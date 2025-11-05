package admin

import (
	"embed"

	"vitego/admin/hook"
	"vitego/adminui"
	"vitego/api"

	"github.com/daodao97/xgo/xadmin"
	"github.com/daodao97/xgo/xapp"
	"github.com/gin-gonic/gin"
)

//go:embed route.jsonc
var routes string

//go:embed schema
var schema embed.FS

func SetupRouter(e *gin.Engine) {
	xadmin.SetRoutes(routes)
	xadmin.InitSchema(schema)
	xadmin.SetAdminPath("/_")
	xadmin.SetUI(adminui.AdminUI)
	xadmin.SetJwt(&xadmin.JwtConf{
		Secret:      "shipnow_admin_jwt_secret",
		TokenExpire: 3600,
	})
	xadmin.SetWebSite(map[string]any{
		"title":         "aicoding",
		"logo":          "/_/claude.svg",
		"defaultAvatar": "/_/claude.svg",
	})
	hook.RegHook()

	g := xadmin.GinRoute(e)

	// team owner create/update
	g.POST("/example", xapp.RegisterAPI(api.Exmaple))
}
