package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Service struct {
	bun.BaseModel           `bun:"table:services"`
	ID                      int       `bun:"id,pk,autoincrement"`
	UUID                    string    `bun:"uuid,unique,notnull"`
	ServicetemplateID       int       `bun:"servicetemplate_id,notnull"`
	HostID                  int       `bun:"host_id,notnull"`
	Name                    *string   `bun:"name,nullzero"`
	Description             *string   `bun:"description,nullzero"`
	ServiceName             string    `bun:",scanonly"`
	CommandID               *int      `bun:"command_id,nullzero"`
	CheckCommandArgs        string    `bun:"check_command_args,notnull"`
	EventhandlerCommandID   *int      `bun:"eventhandler_command_id,nullzero"`
	NotifyPeriodID          *int      `bun:"notify_period_id,nullzero"`
	CheckPeriodID           *int      `bun:"check_period_id,nullzero"`
	CheckInterval           *float64  `bun:"check_interval,nullzero"`
	RetryInterval           *float64  `bun:"retry_interval,nullzero"`
	MaxCheckAttempts        *int      `bun:"max_check_attempts,nullzero"`
	FirstNotificationDelay  *float64  `bun:"first_notification_delay,nullzero"`
	NotificationInterval    *float64  `bun:"notification_interval,nullzero"`
	NotifyOnWarning         *int      `bun:"notify_on_warning,nullzero"`
	NotifyOnUnknown         *int      `bun:"notify_on_unknown,nullzero"`
	NotifyOnCritical        *int      `bun:"notify_on_critical,nullzero"`
	NotifyOnRecovery        *int      `bun:"notify_on_recovery,nullzero"`
	NotifyOnFlapping        *int      `bun:"notify_on_flapping,nullzero"`
	NotifyOnDowntime        *int      `bun:"notify_on_downtime,nullzero"`
	IsVolatile              *int      `bun:"is_volatile,nullzero"`
	FlapDetectionEnabled    *int      `bun:"flap_detection_enabled,nullzero"`
	FlapDetectionOnOk       *int      `bun:"flap_detection_on_ok,nullzero"`
	FlapDetectionOnWarning  *int      `bun:"flap_detection_on_warning,nullzero"`
	FlapDetectionOnUnknown  *int      `bun:"flap_detection_on_unknown,nullzero"`
	FlapDetectionOnCritical *int      `bun:"flap_detection_on_critical,nullzero"`
	LowFlapThreshold        *float64  `bun:"low_flap_threshold,nullzero"`
	HighFlapThreshold       *float64  `bun:"high_flap_threshold,nullzero"`
	ProcessPerformanceData  *int      `bun:"process_performance_data,nullzero"`
	FreshnessChecksEnabled  *int      `bun:"freshness_checks_enabled,nullzero"`
	FreshnessThreshold      *int      `bun:"freshness_threshold,nullzero"`
	PassiveChecksEnabled    *int      `bun:"passive_checks_enabled,nullzero"`
	EventHandlerEnabled     *int      `bun:"event_handler_enabled,nullzero"`
	ActiveChecksEnabled     *int      `bun:"active_checks_enabled,nullzero"`
	NotificationsEnabled    *int      `bun:"notifications_enabled,nullzero"`
	Notes                   *string   `bun:"notes,nullzero"`
	Priority                *int      `bun:"priority,nullzero"`
	Tags                    *string   `bun:"tags,nullzero"`
	OwnContacts             *int      `bun:"own_contacts,nullzero"`
	OwnContactgroups        *int      `bun:"own_contactgroups,nullzero"`
	OwnCustomvariables      *int      `bun:"own_customvariables,nullzero"`
	ServiceURL              *string   `bun:"service_url,nullzero"`
	SlaRelevant             *int      `bun:"sla_relevant,nullzero"`
	ServiceType             int       `bun:"service_type,notnull,default:1"`
	Disabled                int       `bun:"disabled,default:0"`
	UsageFlag               int       `bun:"usage_flag,notnull"`
	Created                 time.Time `bun:"created,notnull"`
	Modified                time.Time `bun:"modified,notnull"`
}
