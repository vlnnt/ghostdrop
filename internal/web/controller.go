package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

//go:embed static
var staticFiles embed.FS

type SettingsSaver func(token, guildID string) error

type Handler struct {
	log          *zap.Logger
	saveSettings SettingsSaver
	configPath   string
}

func NewHandler(
	log *zap.Logger,
	saveSettings SettingsSaver,
	configPath string,
) *Handler {
	return &Handler{
		log:          log,
		saveSettings: saveSettings,
		configPath:   configPath,
	}
}

func (h *Handler) InitAPI(
	e *echo.Echo,
) {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(static))

	e.GET("/", echo.WrapHandler(fileServer))
	e.GET("/static/*", echo.WrapHandler(
		http.StripPrefix("/static/", fileServer),
	))

	e.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.GET("/api/settings", h.getSettings)
	e.POST("/api/settings", h.createSettings)
	e.DELETE("/api/settings", h.clearSettings)
}
