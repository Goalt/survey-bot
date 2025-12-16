package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
)

type Server interface {
	Start()
	Stop()
}

type metricsServer struct {
	server *http.Server
	log    logger.Logger
}

func NewMetricsServer(metricsPort int, log logger.Logger) *metricsServer {
	return &metricsServer{
		log: log,
		server: &http.Server{
			Addr:    ":" + fmt.Sprintf("%d", metricsPort),
			Handler: promhttp.Handler(),
		},
	}
}

func (s *metricsServer) Start() {
	if err := s.server.ListenAndServe(); err != nil {
		s.log.Errorf(context.Background(), "failed to start http server: %v", err)
	}
}

func (s *metricsServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Errorf(context.Background(), "failed to stop http server: %v", err)
	}
}
