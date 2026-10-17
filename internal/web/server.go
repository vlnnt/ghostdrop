package web

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type HTTPServer struct {
	router *echo.Echo
	addr   string
	log    *zap.Logger
}

func NewHTTPServer(
	port int,
	handler *Handler,
	log *zap.Logger,
) *HTTPServer {
	router := echo.New()
	router.HideBanner = true
	router.HidePort = true

	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.Recover())
	router.Use(middleware.BodyLimit("16K"))
	router.Use(middleware.Secure())

	router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			expectedHost := net.JoinHostPort(
				"127.0.0.1",
				strconv.Itoa(port),
			)

			if c.Request().Host != expectedHost {
				return c.NoContent(http.StatusForbidden)
			}

			if c.Request().Method == http.MethodPost {
				if c.Request().Header.Get("Origin") !=
					"http://"+expectedHost {
					return c.NoContent(http.StatusForbidden)
				}
			}

			return next(c)
		}
	})

	webHost := os.Getenv("GHOSTDROP_WEB_HOST")
	if webHost == "" {
		webHost = "127.0.0.1"
	}

	handler.InitAPI(router)
	return &HTTPServer{
		router: router,
		addr:   net.JoinHostPort(webHost, strconv.Itoa(port)),
		log:    log,
	}
}

func (s *HTTPServer) ListenAndServe() error {
	s.log.Info(
		"Starting GhostDrop HTTP server",
		zap.String("address", s.addr),
	)

	err := s.router.Start(s.addr)
	if err != nil && err != http.ErrServerClosed {
		s.log.Error("HTTP server failed", zap.Error(err))
		return err
	}

	return nil
}

func (s *HTTPServer) GracefulShutdown() error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	if err := s.router.Shutdown(ctx); err != nil {
		s.log.Error(
			"HTTP server shutdown failed",
			zap.Error(err),
		)
		return err
	}

	s.log.Info("HTTP server stopped!")
	return nil
}
