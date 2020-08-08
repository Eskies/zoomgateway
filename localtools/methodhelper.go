package localtools

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

func LogThisError(ctx *fasthttp.RequestCtx, message string) {
	fmt.Printf("|%s| ERROR ["+string(ctx.Host())+"] "+string(ctx.RequestURI())+"\n", time.Now().Format("02-01-2006 15:04:05"))
	fmt.Printf("\t-> %s\n", message)
	_, _ = ctx.WriteString(message)
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetConnectionClose()

	file, err := os.OpenFile("error-log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("|%s| ERROR ["+string(ctx.Host())+"] "+string(ctx.RequestURI())+"\n\t-> %s\n", time.Now().Format("02-01-2006 15:04:05"), message))
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
}

func LogThisErrorWCode(ctx *fasthttp.RequestCtx, errcode int, message string) {
	fmt.Printf("|%s| ERROR ["+string(ctx.Host())+"] "+string(ctx.RequestURI())+"\n", time.Now().Format("02-01-2006 15:04:05"))
	fmt.Printf("\t-> %s\n", message)
	_, _ = ctx.WriteString(message)
	ctx.SetStatusCode(errcode)
	ctx.SetConnectionClose()

	file, err := os.OpenFile("error-log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("|%s| ERROR ["+string(ctx.Host())+"] "+string(ctx.RequestURI())+"\n\t-> %s\n", time.Now().Format("02-01-2006 15:04:05"), message))
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
}

func ProblemValue(errormsg string, values []interface{}) (string, bool) {
	for _, e := range values {
		if strings.Contains(errormsg, "'"+e.(string)+"'") {
			return e.(string), true
		}
	}

	return "", false
}
