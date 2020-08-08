package controller

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"zoomgateway/localtools"

	"github.com/huandu/go-sqlbuilder"
	"github.com/patrickmn/go-cache"

	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
)

func LoginController(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}, jwtkey []byte, poolcache *cache.Cache) {

	nim := string(ctx.FormValue("username"))

	sqlB := sqlbuilder.NewSelectBuilder()
	sqlB.Select("nama")
	sqlB.From("peserta")
	sqlB.Where(sqlB.E("id", nim))

	var namamhs string
	a, b := sqlB.Build()
	err := dbConn.QueryRow(a, b...).Scan(&namamhs)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			ctx.WriteString("ID yang anda masukan tidak terdaftar.")
			ctx.SetStatusCode(404)
		} else {
			localtools.LogThisError(ctx, err.Error())
		}
		return
	}

	if _, found := poolcache.Get("psr-" + nim); found {
		ctx.WriteString("NIM ini hanya bisa digunakan login pada satu komputer saja, silahkan logout terlebih dahulu ditempat lain. Jika tidak sengaja tertutup, mohon tunggu selama 5 menit sebelum mencoba kembali.")
		ctx.SetStatusCode(400)
		return
	} else {
		h := sha1.New()

		h.Write([]byte(fmt.Sprintf("%s-ssc(%s)", nim, time.Now().Format("02-01-2006 15:04:05"))))
		scCode := h.Sum(nil)
		ssCode := fmt.Sprintf("%x", scCode)

		jt := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"nim":    nim,
			"nama":   namamhs,
			"sscode": ssCode,
			"aud":    string(ctx.Host()),
			"exp":    time.Now().Add(2 * time.Hour).Unix(),
		})
		token, err := jt.SignedString(jwtkey)
		if err != nil {
			localtools.LogThisError(ctx, err.Error())
			return
		}

		poolcache.Set("psr-"+nim, ssCode, cache.DefaultExpiration)

		var cookieSC fasthttp.Cookie
		cookieSC.SetKey("auth")
		cookieSC.SetValue(token)
		cookieSC.SetHTTPOnly(true)
		cookieSC.SetPath("/")
		cookieSC.SetMaxAge(1800)
		cookieSC.SetDomain(string(ctx.Host()))
		ctx.Response.Header.SetCookie(&cookieSC)

	}

	ctx.SetStatusCode(200)
	ctx.SetConnectionClose()

	return
}
