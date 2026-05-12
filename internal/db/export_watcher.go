package db

import (
	"context"
	"database/sql"
	"log"
	"push_notification/internal/hub"
	"sync"
	"time"

	"github.com/uptrace/bun"
)

type ExportWatcher struct {
	db     *DB
	stopCh chan struct{}
	wg     sync.WaitGroup
	hub    *hub.Hub
}

func NewExportWatcher(db *DB, hub *hub.Hub) *ExportWatcher {
	return &ExportWatcher{
		db:     db,
		stopCh: make(chan struct{}),
		hub:    hub,
	}
}

func (w *ExportWatcher) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-w.stopCh:
				log.Println("ExportWatcher stopped")
				return
			case <-ticker.C:
				w.checkExportStarted()
			}
		}
	}()
}

func (w *ExportWatcher) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

type ExportRecord struct {
	bun.BaseModel `bun:"table:exports"`

	ID           int       `bun:"id"`
	Task         string    `bun:"task"`
	Text         string    `bun:"text"`
	Finished     int       `bun:"finished"`
	Successfully int       `bun:"successfully"`
	Created      time.Time `bun:"created"`
	Modified     time.Time `bun:"modified"`
}

func (w *ExportWatcher) checkExportStarted() {
	if !w.hub.HasClients() {
		// If no clients are connected, we can skip the database check to reduce load
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var record ExportRecord
	err := w.db.Bun.NewSelect().Model(&record).Where("task = ?", "export_started").Order("id DESC").Limit(1).Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("ExportWatcher DB error: %v", err)
		return
	}
	if err == nil {
		//log.Printf("ExportWatcher: export_started record found: %+v", record)
		exportRunning := false
		if record.Finished == 0 {
			exportRunning = true
		}

		// Send message to all WebSocket clients
		if w.hub != nil {
			w.hub.Broadcast(hub.ResponseMessage{
				Type:    hub.ResponseExportStatus,
				Message: "", // Empty message to keep the payload small
				Payload: exportRunning,
			})
		}
	}
}
