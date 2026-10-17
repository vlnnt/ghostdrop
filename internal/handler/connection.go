package handler

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func OpenSessionWithRetry(
	ctx context.Context,
	session *discordgo.Session,
	log *zap.Logger,
) error {
	const (
		initialDelay = 2 * time.Second
		maxDelay     = 60 * time.Second
	)

	delay := initialDelay
	attempt := 0

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		attempt++
		log.Info(
			"Connecting to Discord Gateway...",
			zap.Int("attempt", attempt),
		)

		err := session.Open()
		if err == nil {
			log.Info(
				"Discord session opened successfully",
				zap.Int("attempt", attempt),
			)

			return nil
		}

		if !isRetryableConnectionError(err) {
			log.Error(
				"Non-retryable Discord connection error",
				zap.Error(err),
			)

			return err
		}

		log.Warn(
			"Discord connection failed, retrying...",
			zap.Int("attempt", attempt),
			zap.Duration("retry_in", delay),
			zap.Error(err),
		)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()

		case <-timer.C:
		}

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

func isRetryableConnectionError(
	err error,
) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") ||
		strings.Contains(message, "connection refused") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "temporary failure") ||
		strings.Contains(message, "eof")
}
