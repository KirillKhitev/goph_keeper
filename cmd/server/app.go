// Сервер для приема данных пользователей.
package main

import (
	"context"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/gzip"
	"github.com/KirillKhitev/goph_keeper/internal/handlers"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// app основная структура приложения.
type app struct {
	store  store.Store
	server http.Server
}

// newApp конструктор приложения.
func newApp(s store.Store) *app {
	instance := &app{
		store: s,
	}

	return instance
}

// Close корректно останавливает приложение.
func (a *app) Close() error {
	if err := a.shutdownServer(); err != nil {
		return fmt.Errorf("error by Server shutdown: %w", err)
	}

	if err := a.store.Close(); err != nil {
		return fmt.Errorf("error by closing Store: %w", err)
	}

	log.Println("Store graceful shutdown complete.")

	return nil
}

// StartServer запускает сервер.
func (a *app) StartServer() error {
	log.Printf("Running server: %v", config.ConfigServer)

	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Handle("/register", &handlers.Register{
			Store: a.store,
		})
		r.Handle("/login", &handlers.Login{
			Store: a.store,
		})
	})

	r.Route("/api/data", func(r chi.Router) {
		r.Handle("/update", &handlers.Update{
			Store: a.store,
		})
		r.Handle("/list", &handlers.List{
			Store: a.store,
		})
		r.Handle("/get", &handlers.Get{
			Store: a.store,
		})
	})

	handler := gzip.Middleware(r)

	a.server = http.Server{
		Addr:    config.ConfigServer.AddrRun,
		Handler: handler,
	}

	return a.server.ListenAndServe()
}

// shutdownServer останавливает сервер.
func (a *app) shutdownServer() error {
	shutdownCtx, shutdownRelease := context.WithCancel(context.TODO())
	defer shutdownRelease()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP shutdown error: %w", err)
	}

	log.Println("HTTP graceful shutdown complete.")

	return nil
}

// CatchTerminateSignal ловит сигналы остановки сервера.
func (a *app) CatchTerminateSignal() error {
	terminateSignals := make(chan os.Signal, 1)

	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignals

	if err := a.Close(); err != nil {
		return err
	}

	log.Println("Terminate app complete")

	return nil
}
