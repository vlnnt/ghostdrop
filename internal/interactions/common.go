package interactions

import (
	"ghostdrop/internal/settings"

	"go.uber.org/zap"
)

type Interactions struct {
	log             *zap.Logger
	channelSettings settings.ChannelSettings
}

func NewInteractions(
	log *zap.Logger,
	channelSettings settings.ChannelSettings,
) *Interactions {
	return &Interactions{
		log:             log,
		channelSettings: channelSettings,
	}
}
