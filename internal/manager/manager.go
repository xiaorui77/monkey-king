package manager

import (
	"context"
	"github.com/xiaorui77/goutils/httpr"
	"github.com/xiaorui77/goutils/logx"
	"github.com/xiaorui77/monker-king/internal/engine/types"
	"net/http"
	"time"
)

type Manager struct {
	collector types.Collect

	server  *http.Server
	router  *httpr.Httpr
	runChan chan struct{}
}

func NewManager(c types.Collect) *Manager {
	m := &Manager{
		collector: c,
		router:    httpr.NewEngine(),
		runChan:   make(chan struct{}),
	}
	m.server = &http.Server{
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  15 * time.Second,

		Handler: m.router,
	}

	m.router.POST("/api/v1/task", m.HandleAddTask)
	m.router.GET("/api/v1/tasks", m.HandleListTask)

	return m
}

// Run the server in blocking mode.
func (m *Manager) Run(ctx context.Context) {
	go func() {
		logx.Errorf("HTTP Server start at %v", m.server.Addr)
		if err := m.server.ListenAndServe(); err != nil {
			logx.Errorf("HTTP Server crashed: %v", err)
		}
		logx.Infof("HTTP Server stopped")
		close(m.runChan)
	}()

	select {
	case <-ctx.Done():
		logx.Info("HTTP Server shutdown...")
		if err := m.server.Shutdown(ctx); err != nil {
			logx.Errorf("HTTP Server shutdown error: %v", err)
		} else {
			logx.Info("HTTP Server shutdown success")
		}
	case <-m.runChan:
		logx.Errorf("HTTP Server unexpected stopped")
	}
}