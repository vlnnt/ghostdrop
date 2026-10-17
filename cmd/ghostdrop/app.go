package main

import (
	"context"
	"errors"
	"ghostdrop/internal/handler"
	"ghostdrop/internal/interactions"
	"ghostdrop/internal/settings"
	"ghostdrop/internal/web"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func Run() {
	logger, err := zap.NewProduction()
	if err != nil {
		return
	}
	defer func() {
		_ = logger.Sync()
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	configPath := settings.GetConfigPathFromArgs()
	settingsSaved := make(chan struct{}, 1)

	logger.Info("Initializing web handler...")
	webHandler := web.NewHandler(
		logger,
		func(token, guildID string) error {
			cfg := &settings.BotSettings{
				Discord: settings.DiscordSettings{
					Token:   token,
					GuildID: guildID,
				},
			}

			if err := settings.SaveConfig(cfg, configPath); err != nil {
				logger.Error(
					"Failed to save configuration",
					zap.Error(err),
				)
				return err
			}

			logger.Info(
				"Configuration saved successfully",
				zap.Bool("global_commands", guildID == ""),
			)

			select {
			case settingsSaved <- struct{}{}:
			default:
			}

			return nil
		},
		configPath,
	)

	logger.Info("Initializing HTTP server...")
	httpServer := web.NewHTTPServer(
		8080,
		webHandler,
		logger,
	)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	defer func() {
		logger.Info("Shutting down HTTP server...")
		if err := httpServer.GracefulShutdown(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"HTTP shutdown failed",
				zap.Error(err),
			)
		}
	}()

	logger.Info("Loading GhostDrop configuration...")
	cfg, err := settings.LoadConfig(configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Error(
				"Invalid configuration",
				zap.Error(err),
			)
			return
		}

		logger.Info("Configuration not found, opening setup...")
		if os.Getenv("GHOSTDROP_DISABLE_BROWSER") != "true" {
			if err := web.OpenBrowser(
				"http://127.0.0.1:8080",
				logger,
			); err != nil {
				logger.Warn(
					"Failed to open browser",
					zap.Error(err),
				)
			}
		} else {
			logger.Info(
				"Browser opening disabled",
				zap.String("url", "http://127.0.0.1:8080"),
			)
		}

		select {
		case <-settingsSaved:
			logger.Info(
				"Settings saved, starting Discord bot...",
			)

		case <-ctx.Done():
			logger.Info("Shutdown requested")
			return

		case err := <-serverErrors:
			logger.Error(
				"HTTP server stopped during setup",
				zap.Error(err),
			)
			return
		}

		cfg, err = settings.LoadConfig(configPath)
		if err != nil {
			logger.Error(
				"Failed to reload configuration",
				zap.Error(err),
			)
			return
		}
	}

	logger.Info("Initializing interactions...")
	interaction := interactions.NewInteractions(logger)

	logger.Info("Initializing Discord handler...")
	discordHandler := handler.NewHandler(
		logger,
		interaction,
	)

	logger.Info("Initializing Discord session...")
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		logger.Error(
			"Failed to create Discord session",
			zap.Error(err),
		)
		return
	}

	session.ShouldReconnectOnError = true
	session.AddHandler(discordHandler.Handle)

	defer func() {
		logger.Info("Closing Discord session...")
		if err := session.Close(); err != nil {
			logger.Error(
				"Failed to close Discord session",
				zap.Error(err),
			)
		}
	}()

	connectionResult := make(chan error, 1)
	go func() {
		connectionResult <- handler.OpenSessionWithRetry(
			ctx,
			session,
			logger,
		)
	}()

	logger.Info("Waiting for Discord connection...")
	select {
	case err := <-connectionResult:
		if err != nil {
			logger.Error(
				"Failed to establish Discord connection",
				zap.Error(err),
			)
			return
		}

	case <-ctx.Done():
		logger.Info(
			"Shutdown requested during Discord connection",
		)
		return

	case err := <-serverErrors:
		logger.Error(
			"HTTP server stopped during Discord connection",
			zap.Error(err),
		)
		return
	}

	logger.Info("Registering Discord commands...")
	if err := discordHandler.RegisterCommand(
		session,
		cfg.Discord.GuildID,
	); err != nil {
		logger.Error(
			"Failed to register Discord commands",
			zap.Error(err),
		)
		return
	}

	logger.Info(
		"GhostDrop started successfully!",
		zap.Bool("global_commands", cfg.Discord.GuildID == ""),
	)

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")

	case err := <-serverErrors:
		if err != nil {
			logger.Error(
				"HTTP server stopped",
				zap.Error(err),
			)
		}
	}

	logger.Info("Shutting down GhostDrop...")
}
