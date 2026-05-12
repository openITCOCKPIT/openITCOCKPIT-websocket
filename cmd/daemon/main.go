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
	// Create a context that is canceled on SIGINT or SIGTERM for graceful shutdown
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to database
	dbConn, err := db.NewDBFromMyCnf("/opt/openitc/etc/mysql/mysql.cnf")
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}

	h := hub.NewHub()
	go h.Run()

	// Start ExportWatcher to monitor for config refresh and broadcast changes to clients
	exportWatcher := db.NewExportWatcher(dbConn, h)
	exportWatcher.Start()

	appCtx := context.Background()

	// Initialize WebHookService which will send Mobile Push Notifications via the public relay server for iOS and Android devices
	whService, err := webhook.NewWebHookService(appCtx, dbConn)
	if err != nil {
		log.Fatalf("Failed to initialize WebHookService: %v", err)
	}

	// The Router will handle incoming HTTP requests for WebSocket connections and message inputs.
	r := router.NewRouter(appCtx, h, whService, dbConn)

	srv := &http.Server{
		Addr:    "127.0.0.1:8083",
		Handler: r,
	}

	// Start the HTTP server in a separate goroutine so that it doesn't block the main thread, allowing us to listen for shutdown signals.
	go func() {
		log.Println("Server started on 127.0.0.1:8083")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdownCtx.Done()
	log.Println("Shutdown signal received")

	exportWatcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	h.Shutdown()
	log.Println("Server exited gracefully")
}
