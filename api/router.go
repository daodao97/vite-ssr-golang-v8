package api

import (
	"github.com/daodao97/xgo/xapp"
	"github.com/gin-gonic/gin"

	loginapi "github.com/revenkroz/vite-ssr-golang/api/login"
)

func SetupRouter(r *gin.Engine) {
	g := r.Group("/api")

	g.GET("/example", xapp.HanderFunc(Exmaple))
	g.POST("/auth/login/google", xapp.HanderFunc(loginapi.AuthLoginGoogle))
	g.POST("/auth/login/email/request", xapp.HanderFunc(loginapi.AuthRequestEmailCode))
	g.POST("/auth/login/email/verify", xapp.HanderFunc(loginapi.AuthVerifyEmailCode))
	g.POST("/auth/logout", xapp.HanderFunc(loginapi.AuthLogout))
	g.GET("/auth/session", xapp.HanderFunc(loginapi.AuthSession))
}

type ReqExample struct{}

type RespExample struct{}

func Exmaple(ctx *gin.Context, req *ReqExample) (*ReqExample, error) {
	return &ReqExample{}, nil
}
