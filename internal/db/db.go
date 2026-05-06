package db

import (
	"context"
	"fmt"
	"log"

	"database/sql"

	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
	"gopkg.in/ini.v1"
)

type DB struct {
	Bun *bun.DB
}

// Define the structure according to the [client] section
type MySQLConfig struct {
	Client struct {
		Database string `ini:"database"`
		Host     string `ini:"host"`
		User     string `ini:"user"`
		Password string `ini:"password"`
		Port     int    `ini:"port"`
	} `ini:"client"`
}

// parseMyCnf parses a my.cnf file and returns a DSN string
func parseMyCnf(path string) (string, error) {

	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalf("Error while loading the file: %v", err)
	}

	var myConfig MySQLConfig
	err = cfg.MapTo(&myConfig)
	if err != nil {
		log.Fatalf("Error while mapping the file: %v", err)
	}

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		myConfig.Client.User, myConfig.Client.Password, myConfig.Client.Host, myConfig.Client.Port, myConfig.Client.Database)
	return dsn, nil
}

func NewDBFromMyCnf(path string) (*DB, error) {
	dsn, err := parseMyCnf(path)
	if err != nil {
		return nil, err
	}
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	bunDB := bun.NewDB(sqldb, mysqldialect.New())
	bunDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(false)))
	return &DB{Bun: bunDB}, nil
}

// GetAPIKey liest den API-Key aus der Datenbank
func (db *DB) GetAPIKey(ctx context.Context) (string, error) {
	type Systemsettings struct {
		Key   string `bun:"key"`
		Value string `bun:"value"`
	}
	var s Systemsettings
	err := db.Bun.NewSelect().
		Model(&s).Column("key", "value").
		Where("`key` = ?", "SUDO_SERVER.API_KEY").
		Scan(ctx)
	if err != nil {
		return "", err
	}
	return s.Value, nil
}
