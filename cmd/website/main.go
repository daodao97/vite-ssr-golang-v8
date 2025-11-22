package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"vitego/admin"
	"vitego/api"
	"vitego/api/page"
	"vitego/conf"
	"vitego/dao"
	"vitego/job"
	"vitego/pkg"
	"vitego/pkg/routematcher"
	"vitego/webssr"

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

	if os.Getenv("CRON_ENABLE") == "true" {
		app.AddServer(job.NewCronServer())
	}

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
	fsyFrontend, _ := fs.Sub(webssr.FrontendDist, "dist/client")
	fsyServer, _ := fs.Sub(webssr.ServerDist, "dist/server")

	pkg.RunBlocking(
		r,
		pkg.FrontendBuild{
			FrontendDist: fsyFrontend,
			ServerDist:   fsyServer,
		},
		registerSSRFetchRoutes(r),
	)
}

const ssrFetchPrefix = pkg.DefaultSSRFetchPrefix

func registerSSRFetchRoutes(r *gin.Engine) pkg.BackendDataFetcher {
	group := r.Group(ssrFetchPrefix)
	page.Router(group)

	return func(ctx context.Context, req *http.Request) (routematcher.SSRPayload, error) {
		cloned := req.Clone(ctx)

		originalPath := req.URL.Path
		if !strings.HasPrefix(originalPath, "/") {
			originalPath = "/" + originalPath
		}

		ssrPath := ssrFetchPrefix + originalPath
		cloned.URL.Path = ssrPath
		cloned.URL.RawPath = ""
		cloned.URL.RawQuery = req.URL.RawQuery
		cloned.RequestURI = ssrPath
		if cloned.URL.RawQuery != "" {
			cloned.RequestURI += "?" + cloned.URL.RawQuery
		}

		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, cloned)

		res := recorder.Result()
		defer res.Body.Close()

		if res.StatusCode == http.StatusNotFound {
			return mapPayload{}, nil
		}

		if res.StatusCode >= 400 {
			body, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf("ssr fetch %s returned %d: %s", ssrPath, res.StatusCode, strings.TrimSpace(string(body)))
		}

		var data map[string]any
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			return nil, err
		}

		return mapPayload(data), nil
	}
}

type mapPayload map[string]any

func (m mapPayload) AsMap() map[string]any {
	return m
}
