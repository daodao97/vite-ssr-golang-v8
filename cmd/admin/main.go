package main

import (
	"log/slog"
	"os"

	"vitego/admin"
	"vitego/conf"
	"vitego/dao"

	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xredis"
	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"
)

var Version string

func init() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "dev" {
		xlog.SetLogger(xlog.StdoutTextPretty(xlog.WithLevel(slog.LevelDebug)))
		xrequest.SetRequestDebug(true)
	} else {
		xlog.SetLogger(xlog.StdoutJson(xlog.WithLevel(slog.LevelInfo)))
		xrequest.SetRequestDebug(false)
	}
}

func main() {
	app := xapp.NewApp().
		AddStartup(
			conf.Init,
			func() error {
				return xredis.Inits(conf.Get().Redis)
			},
			dao.Init,
		).
		AddServer(xapp.NewGinHttpServer(xapp.Args.Bind, h))

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func h() *gin.Engine {
	r := xapp.NewGin(xapp.WithPrintReqeustLog(false))
	admin.SetupRouter(r)

	return r
}
