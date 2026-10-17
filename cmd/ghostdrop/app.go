package main

import (
	"ghostdrop/internal/handler"
	"ghostdrop/internal/interactions"
	"ghostdrop/internal/settings"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func Run() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Initializing config...")
	config, err := settings.LoadConfig(settings.GetConfigPathFromArgs())
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Initializing config OK!")

	logger.Info("Initializing interactions...")
	interaction := interactions.NewInteractions(logger)

	logger.Info("Initializing interactions OK!")

	logger.Info("Initializing handler...")
	handler := handler.NewHandler(
		logger,
		interaction,
	)

	logger.Info("Initializing handler OK!")

	logger.Info("Initializing token...")
	session, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		return
	}

	logger.Info("Initializing token OK!")

	logger.Info("Opening session...")
	err = session.Open()
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Opening session OK!")

	logger.Info("Initializing discord add handler...")
	session.AddHandler(handler.Handle)

	logger.Info("Initializing discord add handler OK!")

	logger.Info("Initializing register command...")
	handler.RegisterCommand(session, config.Discord.GuildID)

	logger.Info("Initializing register command OK!")
	defer handler.UnregisterCommand(session, config.Discord.GuildID)

	logger.Info("Bot started!")

	result := make(chan os.Signal, 1)
	signal.Notify(result, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-result
}
