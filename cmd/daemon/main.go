package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"push_notification/internal/db"
	"push_notification/internal/hub"
	"push_notification/internal/router"
	"push_notification/internal/watcher"
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

	appCtx := context.Background()

	// Initialize WebHookService which will send Mobile Push Notifications via the public relay server for iOS and Android devices
	whService, err := webhook.NewWebHookService(appCtx, dbConn)
	if err != nil {
		log.Fatalf("Failed to initialize WebHookService: %v", err)
	}

	h := hub.NewHub(whService)
	go h.Run()

	// Start ExportWatcher to monitor for config refresh and broadcast changes to clients
	exportWatcher := watcher.NewExportWatcher(dbConn, h)
	exportWatcher.Start()

	// The Router will handle incoming HTTP requests for WebSocket connections and message inputs.
	r := router.NewRouter(appCtx, h, whService, dbConn)

	// Only bind on 127.0.0.1 for traditional deployments (apt or dnf install)
	listenAddress := "127.0.0.1:8083"
	if os.Getenv("IS_CONTAINER") == "1" {
		// We are running in a Docker container, so we have to bind on all interfaces so that the Mod_Gearman container can reach us.
		// Mod_Gearman will use send_push_notification to send push notifications through THIS server.
		listenAddress = "0.0.0.0:8083"
	}

	srv := &http.Server{
		Addr:    listenAddress,
		Handler: r,
	}

	// Start the HTTP server in a separate goroutine so that it doesn't block the main thread, allowing us to listen for shutdown signals.
	go func() {
		log.Printf("Server started on %s", listenAddress)
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
