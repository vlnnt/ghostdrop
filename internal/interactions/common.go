package interactions

import "go.uber.org/zap"

type Interactions struct {
	log *zap.Logger
}

func NewInteractions(
	log *zap.Logger,
) *Interactions {
	return &Interactions{
		log: log,
	}
}
