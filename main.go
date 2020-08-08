package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"zoomgateway/controller"
	"zoomgateway/localtools"

	"github.com/dgrijalva/jwt-go"
	"github.com/patrickmn/go-cache"

	"github.com/benthor/clustersql"
	"github.com/fasthttp/router"
	"github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
)

var (
	dbuser        string
	dbpassword    string
	dbport        string
	dbname        string
	listenAdr     string
	jwtkey        []byte
	databaseHosts map[string]interface{}
	dbconnection  *sql.DB
	poolcache     *cache.Cache

	pagesettings map[string]interface{}
	zoomsettings map[string]interface{}
)

func main() {
	loadSettings()

	//OPEN DATABASE
	mysqldriver := mysql.MySQLDriver{}
	clusterDriver := clustersql.NewDriver(mysqldriver)
	for hostname, hostip := range databaseHosts {
		clusterDriver.AddNode(hostname, dbuser+":"+dbpassword+"@tcp("+hostip.(string)+":"+dbport+")/"+dbname)
	}
	sql.Register("db-cluster", clusterDriver)
	dbConn, err := sql.Open("db-cluster", "-")
	dbconnection = dbConn
	if err != nil {
		panic(err.Error)
	}

	poolcache = cache.New(6*time.Minute, 3*time.Minute)

	var panicrouteHandler = func(ctx *fasthttp.RequestCtx, info interface{}) {
		localtools.LogThisError(ctx, info.(string))
		fmt.Printf("|%s| Panic n Recover:\n", time.Now().Format("02-01-2006 15:04:05"))
		ctx.WriteString("Sorry we are too busy right now!")
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetConnectionClose()
	}

	r := router.New()
	r.RedirectFixedPath = true
	r.RedirectTrailingSlash = true
	r.PanicHandler = panicrouteHandler

	r.GET("/", func(ctx *fasthttp.RequestCtx) {
		controller.IndexController(dbConn, ctx, pagesettings, zoomsettings, jwtkey, poolcache)
	})

	r.POST("/login", func(ctx *fasthttp.RequestCtx) {
		controller.LoginController(dbConn, ctx, pagesettings, zoomsettings, jwtkey, poolcache)
	})

	r.GET("/logout", func(ctx *fasthttp.RequestCtx) {
		//APIKEY CHECK
		tokenstring := string(ctx.Request.Header.Cookie("auth"))
		if len(tokenstring) > 0 {
			token, _ := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
				return jwtkey, nil
			})

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if string(ctx.Host()) == claims["aud"] {
					nimmhs := claims["nim"].(string)
					poolcache.Delete("psr-" + nimmhs)
				}
			}
		}
		ctx.Redirect("/", 307)
		logOut(ctx)
		ctx.SetConnectionClose()

	})

	r.GET("/seminar", apiAuth(func(ctx *fasthttp.RequestCtx) {
		controller.SeminarPageController(dbConn, ctx, pagesettings, zoomsettings)
	}))

	r.GET("/joinseminar/{idsesi}", apiAuth(func(ctx *fasthttp.RequestCtx) {
		controller.JoinSeminarController(dbConn, ctx, pagesettings, zoomsettings)
	}))

	r.GET("/tokenseminar/{idsesi}", apiAuth(func(ctx *fasthttp.RequestCtx) {
		controller.TokenSeminarController(dbConn, ctx, pagesettings, zoomsettings)
	}))

	r.GET("/retoken", apiAuth(func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(200)
	}))

	s := &fasthttp.Server{
		Handler:           r.Handler,
		Name:              "ZoomGateway by Ngurah Ady K. (ady_kusuma@stikom-bali.ac.id)",
		ReduceMemoryUsage: true,
		ReadTimeout:       time.Duration(30 * time.Second),
	}

	fmt.Println("============================================")
	fmt.Printf("|%s| ZoomGateway  - Starting\n", time.Now().Format("02-01-2006 15:04:05"))
	fmt.Println("Developed by by Ngurah Ady K. (ady_kusuma@stikom-bali.ac.id)")

	fmt.Printf("Title: %s\nDescription: %s\n", pagesettings["title"].(string), pagesettings["description"].(string))
	fmt.Printf("|%s| Listening on: %s\n", time.Now().Format("02-01-2006 15:04:05"), listenAdr)
	log.Fatal(s.ListenAndServe(listenAdr))
}

func loadSettings() {
	var configloaded map[string]interface{}

	plan, _ := ioutil.ReadFile("settings.json")
	err := json.Unmarshal(plan, &configloaded)

	if err != nil {
		panic(fmt.Sprintf("Error opening config file: %s", err.Error()))
	} else {
		pagesettings = configloaded["pagesettings"].(map[string]interface{})
		zoomsettings = configloaded["zoomsettings"].(map[string]interface{})
		config := configloaded["config"].(map[string]interface{})

		databaseHosts = config["dbhost"].(map[string]interface{})
		dbuser = config["dbuser"].(string)
		dbpassword = config["dbpassword"].(string)
		dbport = config["dbport"].(string)
		dbname = config["dbname"].(string)
		listenAdr = config["listenAddress"].(string)
		jwtkey = []byte(config["jwtkey"].(string))
	}
}

func apiAuth(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		//APIKEY CHECK
		tokenstring := string(ctx.Request.Header.Cookie("auth"))
		if len(tokenstring) > 0 {
			token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
				return jwtkey, nil
			})

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if string(ctx.Host()) == claims["aud"] {
					nimmhs := claims["nim"].(string)
					namamhs := claims["nama"].(string)
					sscodeReq := claims["sscode"].(string)
					if x, found := poolcache.Get("psr-" + nimmhs); found {
						if sscodeReq == x.(string) {
							ctx.SetUserValue("namamhs", namamhs)
							ctx.SetUserValue("nimmhs", nimmhs)

							next(ctx)

							poolcache.Delete("psr-" + nimmhs)
							poolcache.Set("psr-"+nimmhs, x.(string), cache.DefaultExpiration)
						} else {
							ctx.Error("Anda terdeteksi melakukan double login!", fasthttp.StatusForbidden)
							ctx.Redirect("/", 307)
							logOut(ctx)
							ctx.SetConnectionClose()
						}
					} else {
						ctx.Error("Silahkan login terlebih dahulu", fasthttp.StatusForbidden)
						ctx.Redirect("/", 307)
						logOut(ctx)
						ctx.SetConnectionClose()
					}
				} else {
					ctx.Error("Kami tidak dapat memverifikasi anda", fasthttp.StatusForbidden)
					ctx.Redirect("/", 307)
					logOut(ctx)
					ctx.SetConnectionClose()
				}
			} else {
				ctx.Error(err.Error(), fasthttp.StatusBadRequest)
				ctx.Redirect("/", 307)
				logOut(ctx)
				ctx.SetConnectionClose()
			}
		} else {
			ctx.Error("Who are you, please relogin or refresh this page?", fasthttp.StatusForbidden)
			ctx.Redirect("/", 307)
			ctx.SetConnectionClose()
		}

	}
}
func logOut(ctx *fasthttp.RequestCtx) {
	var cookieSC fasthttp.Cookie
	cookieSC.SetKey("auth")
	cookieSC.SetValue("")
	cookieSC.SetHTTPOnly(true)
	cookieSC.SetPath("/")
	cookieSC.SetMaxAge(1)
	cookieSC.SetDomain(string(ctx.Host()))
	ctx.Response.Header.SetCookie(&cookieSC)
}
