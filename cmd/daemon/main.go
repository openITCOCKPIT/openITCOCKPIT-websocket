package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"push_notification/internal/db"
	"push_notification/internal/hub"
	"push_notification/internal/router"
	"push_notification/internal/webhook"
)

func main() {
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Datenbankverbindung herstellen
	dbConn, err := db.NewDBFromMyCnf("/opt/openitc/etc/mysql/mysql.cnf")
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}

	h := hub.NewHub()
	go h.Run()

	// ExportWatcher starten (nutzt Hub als Broadcaster)
	exportWatcher := db.NewExportWatcher(dbConn, h)
	exportWatcher.Start()

	// WebHook-Gateways können hier konfiguriert werden
	gateways := []string{
		// "https://example.com/webhook"
	}
	whService := webhook.NewWebHookService(gateways)

	r := router.NewRouter(h, whService, dbConn)

	srv := &http.Server{
		Addr:    ":8083",
		Handler: r,
	}

	go func() {
		log.Println("Server started on :8083")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-shutdownCtx.Done()
	log.Println("Shutdown signal received")

	// ExportWatcher stoppen
	exportWatcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	h.Shutdown()
	log.Println("Server exited gracefully")
}
