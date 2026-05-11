package db

import (
	"context"
	"fmt"
	"log"
	"push_notification/pkg/models"

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
	debugQuery := true
	bunDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(debugQuery)))
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

// GetHostByUUID queries a host by its UUID
func (db *DB) GetHostByUUID(ctx context.Context, uuid string) (*models.Host, error) {
	var host models.Host
	err := db.Bun.NewSelect().
		Model(&host).
		Column("host.id", "host.uuid", "host.name", "host.address").
		Where("host.uuid = ?", uuid).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// GetServiceByUUID queries a service by its UUID
func (db *DB) GetServiceByUUID(ctx context.Context, uuid string) (*models.Service, error) {
	var service models.Service
	err := db.Bun.NewSelect().
		Model(&service).
		Column("service.id", "service.uuid", "service.name", "service.servicetemplate_id").
		ColumnExpr("IF((service.name IS NULL OR service.name=\"\"), servicetemplate.name, service.name) AS service_name").
		Join("INNER JOIN servicetemplates AS servicetemplate ON service.servicetemplate_id = servicetemplate.id").
		Where("service.uuid = ?", uuid).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (db *DB) IsMobilePushNotificationRelayEnabled(ctx context.Context) (bool, error) {
	var relay models.PushNotificationsRelay
	err := db.Bun.NewSelect().
		Model(&relay).
		Where("enabled = ?", true).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return false, err
	}
	return relay.Enabled, nil
}

func (db *DB) GetMobilePushRelay(ctx context.Context) (models.PushNotificationsRelay, error) {
	var relay models.PushNotificationsRelay
	err := db.Bun.NewSelect().
		Model(&relay).
		Where("enabled = ?", true).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return models.PushNotificationsRelay{}, err
	}
	return relay, nil
}

func (db *DB) GetUserMobileDevices(ctx context.Context, userId int64) ([]models.MobileDevice, error) {
	var devices []models.MobileDevice
	err := db.Bun.NewSelect().
		Model(&devices).
		Where("user_id = ?", userId).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (db *DB) DeleteMobileDeviceByDeviceID(ctx context.Context, deviceId string) error {
	_, err := db.Bun.NewDelete().
		Model((*models.MobileDevice)(nil)).
		Where("device_id = ?", deviceId).
		Exec(ctx)
	return err
}
