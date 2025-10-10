package dao

import (
	"github.com/revenkroz/vite-ssr-golang/conf"

	"github.com/daodao97/xgo/xdb"
	_ "github.com/go-sql-driver/mysql"
)

func Init() error {
	err := xdb.Inits(conf.Get().Database)
	if err != nil {
		return err
	}

	ProjectUser = xdb.New("project_user")
	return nil
}
