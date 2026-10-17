package web

import (
	"fmt"
	"os/exec"
	"runtime"

	"go.uber.org/zap"
)

func OpenBrowser(
	url string,
	log *zap.Logger,
) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(
			"rundll32",
			"url.dll,FileProtocolHandler",
			url,
		)

	case "linux":
		cmd = exec.Command("xdg-open", url)

	case "darwin":
		cmd = exec.Command("open", url)

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		log.Error(
			"Failed to open browser",
			zap.String("url", url),
			zap.Error(err),
		)
		return err
	}

	log.Info(
		"Browser opened",
		zap.String("url", url),
	)

	return nil
}
