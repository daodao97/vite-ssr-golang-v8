package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xredis"
	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"
	"github.com/revenkroz/vite-ssr-golang/admin"
	"github.com/revenkroz/vite-ssr-golang/api"
	"github.com/revenkroz/vite-ssr-golang/conf"
	"github.com/revenkroz/vite-ssr-golang/dao"
	"github.com/revenkroz/vite-ssr-golang/job"
	"github.com/revenkroz/vite-ssr-golang/pkg"
	"github.com/revenkroz/vite-ssr-golang/pkg/locales"
	"github.com/revenkroz/vite-ssr-golang/pkg/routematcher"
)

//go:embed all:dist/client
var frontendDist embed.FS

//go:embed all:dist/server
var serverDist embed.FS

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
		AddServer(xapp.NewGinHttpServer(xapp.Args.Bind, h)).
		AddServer(job.NewCronServer())

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func h() *gin.Engine {
	r := xapp.NewGin(xapp.WithPrintReqeustLog(false))
	defer func() {
		xapp.GenerateOpenAPIDoc(
			r,
			xapp.WithInfo("team-gpt", Version, ""),
			xapp.WithBearerAuth(),
		)
	}()
	vueSsr(r)
	api.SetupRouter(r)
	admin.SetupRouter(r)

	return r
}

func vueSsr(r *gin.Engine) {
	fsysFrontend, _ := fs.Sub(frontendDist, "dist/client")
	fsysServer, _ := fs.Sub(serverDist, "dist/server")

	matcher := routematcher.New([]routematcher.Route{
		routematcher.RouteOf("/", func(_ context.Context, _ map[string]string, _ url.Values) (homePayload, error) {
			locale := locales.Default

			return homePayload{
				Announcement: announcementByLocale(locale),
				ServerTime:   time.Now().Format(time.RFC1123Z),
				Locale:       locale,
			}, nil
		}),
		routematcher.RouteOf("/hi/:name", func(_ context.Context, params map[string]string, query url.Values) (greetingPayload, error) {
			locale := locales.Default
			name := params["name"]
			if name == "" {
				name = defaultName(locale)
			}

			salutation := query.Get("title")
			if salutation != "" {
				name = fmt.Sprintf("%s %s", salutation, name)
			}

			return greetingPayload{
				Greeting:    greetingByLocale(locale, name),
				GeneratedAt: time.Now().Format(time.RFC3339),
				Locale:      locale,
			}, nil
		}),
		routematcher.RouteOf("/:locale", func(_ context.Context, params map[string]string, _ url.Values) (homePayload, error) {
			locale := locales.Normalize(paramsLocale(params))

			return homePayload{
				Announcement: announcementByLocale(locale),
				ServerTime:   time.Now().Format(time.RFC1123Z),
				Locale:       locale,
			}, nil
		}),
		routematcher.RouteOf("/:locale/hi/:name", func(_ context.Context, params map[string]string, query url.Values) (greetingPayload, error) {
			locale := locales.Normalize(paramsLocale(params))
			name := params["name"]
			if name == "" {
				name = defaultName(locale)
			}

			salutation := query.Get("title")
			if salutation != "" {
				name = fmt.Sprintf("%s %s", salutation, name)
			}

			return greetingPayload{
				Greeting:    greetingByLocale(locale, name),
				GeneratedAt: time.Now().Format(time.RFC3339),
				Locale:      locale,
			}, nil
		}),
	})

	pkg.RunBlocking(
		r,
		pkg.FrontendBuild{
			FrontendDist: fsysFrontend,
			ServerDist:   fsysServer,
		},
		matcher.Fetch,
	)
}

type greetingPayload struct {
	Greeting    string
	GeneratedAt string
	Locale      string
}

func (g greetingPayload) AsMap() map[string]any {
	return map[string]any{
		"greeting":    g.Greeting,
		"generatedAt": g.GeneratedAt,
		"locale":      g.Locale,
	}
}

type homePayload struct {
	Announcement string
	ServerTime   string
	Locale       string
}

func (h homePayload) AsMap() map[string]any {
	return map[string]any{
		"announcement": h.Announcement,
		"serverTime":   h.ServerTime,
		"locale":       h.Locale,
	}
}

func paramsLocale(params map[string]string) string {
	if params == nil {
		return locales.Default
	}

	if value, ok := params["locale"]; ok {
		return value
	}

	return locales.Default
}

func announcementByLocale(locale string) string {
	switch locale {
	case "zh-CN":
		return "欢迎体验 Go + Vite SSR 示例"
	default:
		return "Welcome to the Go + Vite SSR demo"
	}
}

func defaultName(locale string) string {
	switch locale {
	case "zh-CN":
		return "朋友"
	default:
		return "friend"
	}
}

func greetingByLocale(locale string, name string) string {
	switch locale {
	case "zh-CN":
		return fmt.Sprintf("你好，%s！", name)
	default:
		return fmt.Sprintf("Hello, %s!", name)
	}
}
