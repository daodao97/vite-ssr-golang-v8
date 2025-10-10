package conf

import (
	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xredis"
)

var (
	GitTag string
)

type Conf struct {
	Database    []xdb.Config      `json:"database" yaml:"database" envPrefix:"DATABASE"`
	Redis       []xredis.Options  `json:"redis" yaml:"redis" envPrefix:"REDIS"`
	GoogleOAuth GoogleOAuthConfig `json:"google_oauth" yaml:"google_oauth" envPrefix:"GOOGLE_OAUTH"`
	Project     ProjectConfig     `json:"project" yaml:"project" envPrefix:"PROJECT"`
	Email       EmailConfig       `json:"email" yaml:"email" envPrefix:"EMAIL"`
}

type GoogleOAuthConfig struct {
	ClientID     string `json:"client_id" yaml:"client_id" env:"CLIENT_ID"`
	ClientSecret string `json:"client_secret" yaml:"client_secret" env:"CLIENT_SECRET"`
	RedirectURI  string `json:"redirect_uri" yaml:"redirect_uri" env:"REDIRECT_URI"`
}

type ProjectConfig struct {
	AppID string `json:"app_id" yaml:"app_id" env:"APP_ID"`
}

type EmailConfig struct {
	ResendAPIKey string `json:"resend_api_key" yaml:"resend_api_key" env:"RESEND_API_KEY"`
	FromAddress  string `json:"from_address" yaml:"from_address" env:"FROM_ADDRESS"`
	Subject      string `json:"subject" yaml:"subject" env:"SUBJECT"`
}

var ConfInstance *Conf

func Init() error {
	ConfInstance = &Conf{}

	err := xapp.InitConf(ConfInstance)
	if err != nil {
		return err
	}

	xlog.Debug("GitTag", xlog.Any("GitTag", GitTag))

	return nil
}

func Get() *Conf {
	return ConfInstance
}
