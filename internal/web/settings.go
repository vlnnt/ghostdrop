package web

import (
	"errors"
	"fmt"
	"ghostdrop/internal/settings"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type SettingsRequest struct {
	Token   string `json:"token"`
	GuildID string `json:"guild_id"`
}

type SettingsResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type CurrentSettingsResponse struct {
	HasToken bool   `json:"has_token"`
	GuildID  string `json:"guild_id"`
}

var guildIDPattern = regexp.MustCompile(`^[0-9]{17,20}$`)

func (h *Handler) createSettings(
	c echo.Context,
) error {
	start := time.Now()
	h.log.Info(
		"GhostDrop settings request received",
		zap.String("method", c.Request().Method),
		zap.String("path", c.Request().URL.Path),
	)

	if c.Request().Header.Get(echo.HeaderContentType) !=
		echo.MIMEApplicationJSON {
		return c.JSON(http.StatusUnsupportedMediaType, SettingsResponse{
			Error: "content type must be application/json",
		})
	}

	var request SettingsRequest
	if err := c.Bind(&request); err != nil {
		h.log.Warn(
			"Failed to bind settings request",
			zap.Error(err),
		)

		return c.JSON(http.StatusBadRequest, SettingsResponse{
			Error: "invalid settings request",
		})
	}

	request.Token = strings.TrimSpace(request.Token)
	request.GuildID = strings.TrimSpace(request.GuildID)

	if request.Token == "" {
		cfg, err := settings.LoadConfig(h.configPath)
		if err != nil {
			h.log.Warn(
				"Discord token is required",
				zap.Error(err),
			)

			return c.JSON(http.StatusBadRequest, SettingsResponse{
				Error: "discord token is required",
			})
		}

		request.Token = cfg.Discord.Token
	}

	if request.GuildID != "" &&
		!guildIDPattern.MatchString(request.GuildID) {
		h.log.Warn("Invalid Discord guild ID")

		return c.JSON(http.StatusBadRequest, SettingsResponse{
			Error: "invalid Discord guild ID",
		})
	}

	if h.saveSettings == nil {
		err := fmt.Errorf("settings storage is not configured")
		h.log.Error("Failed to save settings", zap.Error(err))

		return c.JSON(http.StatusInternalServerError, SettingsResponse{
			Error: "settings storage is unavailable",
		})
	}

	if err := h.saveSettings(
		request.Token,
		request.GuildID,
	); err != nil {
		h.log.Error(
			"Failed to save GhostDrop settings",
			zap.Error(err),
		)

		return c.JSON(http.StatusInternalServerError, SettingsResponse{
			Error: "failed to save settings",
		})
	}

	h.log.Info(
		"GhostDrop settings saved successfully",
		zap.Bool("global_commands", request.GuildID == ""),
		zap.Duration("duration", time.Since(start)),
	)

	return c.JSON(http.StatusOK, SettingsResponse{
		Message: "settings saved successfully",
	})
}

func (h *Handler) getSettings(
	c echo.Context,
) error {
	cfg, err := settings.LoadConfig(h.configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c.JSON(http.StatusOK, CurrentSettingsResponse{})
		}

		h.log.Error("Failed to load settings", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, SettingsResponse{
			Error: "failed to load settings",
		})
	}

	return c.JSON(http.StatusOK, CurrentSettingsResponse{
		HasToken: cfg.Discord.Token != "",
		GuildID:  cfg.Discord.GuildID,
	})
}

func (h *Handler) clearSettings(
	c echo.Context,
) error {
	if err := os.Remove(h.configPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			h.log.Error("Failed to clear settings", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, SettingsResponse{
				Error: "failed to clear settings",
			})
		}
	}

	h.log.Info("GhostDrop configuration cleared!")
	return c.JSON(http.StatusOK, SettingsResponse{
		Message: "configuration cleared successfully!",
	})
}
