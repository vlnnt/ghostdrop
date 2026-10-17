package handler

import (
	"ghostdrop/internal/interactions"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type handler struct {
	log         *zap.Logger
	interaction *interactions.Interactions
}

func NewHandler(
	log *zap.Logger,
	interaction *interactions.Interactions,
) *handler {
	return &handler{
		log:         log,
		interaction: interaction,
	}
}

func (h *handler) Handle(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) {
	if interaction.Member != nil &&
		interaction.Member.User.ID == session.State.User.ID {
		return
	}

	var err error
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err = h.handleApplicationComponent(session, interaction)
	}

	if err != nil {
		h.log.Error("Handle", zap.Error(err))
	}
}

func (h *handler) RegisterCommand(
	session *discordgo.Session,
	guildID string,
) error {
	h.log.Info("RegisterCommand")
	cmd, err := session.ApplicationCommandBulkOverwrite(
		session.State.User.ID,
		guildID,
		commands,
	)
	if err != nil {
		h.log.Error("Failed to register commands", zap.Error(err))
		return err
	}

	for _, command := range cmd {
		h.log.Info(
			"Registered Command",
			zap.String("command", command.Name),
			zap.String("guild", guildID),
		)
	}

	return nil
}

func (h *handler) UnregisterCommand(
	session *discordgo.Session,
	guildID string,
) error {
	h.log.Info("UnregisterCommand")
	existingCommands, err := session.ApplicationCommands(session.State.User.ID, guildID)
	if err != nil {
		h.log.Error(err.Error())
		return err
	}

	for _, command := range existingCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, guildID, command.ID)
		if err != nil {
			h.log.Error(err.Error())
			return err
		}
		h.log.Info(
			"Unregistered Command",
			zap.String("command", command.Name), zap.String("guild", guildID))
	}

	return nil
}
