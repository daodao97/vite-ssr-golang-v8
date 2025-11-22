package page

import (
	"fmt"
	"net/http"
	"time"

	"vitego/pkg/locales"
	"vitego/pkg/routematcher"

	"github.com/gin-gonic/gin"
)

func Router(group *gin.RouterGroup) {
	group.GET("/", handleSSRFetch(Home))
	group.GET("/hi/:name", handleSSRFetch(Hi))
	group.GET("/:locale", handleSSRFetch(HomeLocale))
	group.GET("/:locale/hi/:name", handleSSRFetch(HiLocale))
}

func handleSSRFetch(h func(*gin.Context) (routematcher.SSRPayload, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := h(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if payload == nil {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		c.JSON(http.StatusOK, payload.AsMap())
	}
}

func Home(_ *gin.Context) (routematcher.SSRPayload, error) {
	locale := locales.Default

	return homePayload{
		Announcement: announcementByLocale(locale),
		ServerTime:   time.Now().Format(time.RFC1123Z),
		Locale:       locale,
	}, nil
}

func Hi(c *gin.Context) (routematcher.SSRPayload, error) {
	locale := locales.Default
	name := c.Param("name")
	if name == "" {
		name = defaultName(locale)
	}

	salutation := c.Query("title")
	if salutation != "" {
		name = fmt.Sprintf("%s %s", salutation, name)
	}

	return greetingPayload{
		Greeting:    greetingByLocale(locale, name),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Locale:      locale,
	}, nil
}

func HomeLocale(c *gin.Context) (routematcher.SSRPayload, error) {
	locale := locales.Normalize(paramLocale(c))

	return homePayload{
		Announcement: announcementByLocale(locale),
		ServerTime:   time.Now().Format(time.RFC1123Z),
		Locale:       locale,
	}, nil
}

func HiLocale(c *gin.Context) (routematcher.SSRPayload, error) {
	locale := locales.Normalize(paramLocale(c))
	name := c.Param("name")
	if name == "" {
		name = defaultName(locale)
	}

	salutation := c.Query("title")
	if salutation != "" {
		name = fmt.Sprintf("%s %s", salutation, name)
	}

	return greetingPayload{
		Greeting:    greetingByLocale(locale, name),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Locale:      locale,
	}, nil
}

func paramLocale(c *gin.Context) string {
	value := c.Param("locale")
	if value == "" {
		return locales.Default
	}

	return value
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
