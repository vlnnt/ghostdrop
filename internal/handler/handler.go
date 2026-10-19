package handler

import (
	"errors"
	"fmt"
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
	if session.State == nil || session.State.User == nil {
		return errors.New("discord session is not ready")
	}

	appID := session.State.User.ID
	h.log.Info(
		"Registering Discord commands",
		zap.String("guild_id", guildID),
		zap.Bool("global_commands", guildID == ""),
	)

	registered, err := session.ApplicationCommandBulkOverwrite(
		appID,
		guildID,
		commands,
	)
	if err != nil {
		h.log.Error(
			"Failed to register commands",
			zap.Error(err),
		)
		return err
	}

	for _, command := range registered {
		h.log.Info(
			"Registered Command",
			zap.String("command", command.Name),
			zap.String("guild", guildID),
		)
	}

	var cleanupErrors []error
	if guildID != "" {
		if err := h.UnregisterCommand(session, ""); err != nil {
			cleanupErrors = append(
				cleanupErrors,
				fmt.Errorf(
					"failed to remove old global command: %w",
					err,
				),
			)
		}
	}

	session.State.RLock()
	guildIDs := make([]string, 0, len(session.State.Guilds))
	for _, guild := range session.State.Guilds {
		if guild == nil || guild.ID == "" {
			continue
		}

		guildIDs = append(guildIDs, guild.ID)
	}

	session.State.RUnlock()
	for _, oldGuildID := range guildIDs {
		if oldGuildID == guildID {
			continue
		}

		if err := h.UnregisterCommand(
			session,
			oldGuildID,
		); err != nil {
			cleanupErrors = append(
				cleanupErrors,
				fmt.Errorf(
					"failed to remove old command from guild %s: %w",
					oldGuildID,
					err,
				),
			)
		}
	}

	if err := errors.Join(cleanupErrors...); err != nil {
		h.log.Error(
			"Failed to clean up old Discord commands",
			zap.Error(err),
		)
		return err
	}

	h.log.Info(
		"Discord commands synchronized successfully",
		zap.String("guild_id", guildID),
		zap.Bool("global_commands", guildID == ""),
	)

	return nil
}

func (h *handler) UnregisterCommand(
	session *discordgo.Session,
	guildID string,
) error {
	appID := session.State.User.ID
	existingCommands, err := session.ApplicationCommands(
		appID,
		guildID,
	)
	if err != nil {
		h.log.Error(
			"Failed to get existing commands",
			zap.String("guild_id", guildID),
			zap.Error(err),
		)
		return err
	}

	for _, command := range existingCommands {
		if command.Name != "anon" ||
			command.Type != discordgo.ChatApplicationCommand {
			continue
		}

		if err := session.ApplicationCommandDelete(
			appID,
			guildID,
			command.ID,
		); err != nil {
			h.log.Error(
				"Failed to unregister command",
				zap.String("command", command.Name),
				zap.String("guild_id", guildID),
				zap.Error(err),
			)
			return err
		}

		h.log.Info(
			"Unregistered old command",
			zap.String("command", command.Name),
			zap.String("guild_id", guildID),
		)
	}

	return nil
}
