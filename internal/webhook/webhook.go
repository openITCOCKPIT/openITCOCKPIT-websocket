package webhook

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"push_notification/internal/db"
	"push_notification/pkg/models"
)

type WebHookService struct {
	appCtx                   context.Context
	db                       *db.DB
	isMobilePushRelayEnabled bool
	relay                    models.PushNotificationsRelay
	authKey                  string
}

func NewWebHookService(appCtx context.Context, dbConn *db.DB) (*WebHookService, error) {
	webHook := &WebHookService{
		appCtx: appCtx,
		db:     dbConn,
	}

	enabled, err := dbConn.IsMobilePushNotificationRelayEnabled(appCtx)
	if err != nil {
		return nil, err
	}
	webHook.isMobilePushRelayEnabled = enabled

	relay, err := dbConn.GetMobilePushRelay(appCtx)
	if err != nil {
		return nil, err
	}
	webHook.relay = relay

	return webHook, nil
}

func (w *WebHookService) IsMobilePushNotificationRelayEnabled() bool {
	return w.isMobilePushRelayEnabled
}

// SendMobilePush sends a mobile push notification to all devices of a user using the relay settings.
// The send is executed in a goroutine for each device.
func (w *WebHookService) SendMobilePush(userId int64, title, message, icon string, notification models.PostMessage) {
	if !w.isMobilePushRelayEnabled {
		return
	}

	devices, err := w.db.GetUserMobileDevices(w.appCtx, userId)
	if err != nil {
		log.Printf("Failed to get mobile devices for user %d: %v", userId, err)
		return
	}

	endpoint := fmt.Sprintf("%s:%d/send-notification", w.relay.Address, w.relay.Port)
	authKey := w.relay.AuthKey

	for _, device := range devices {
		data := map[string]interface{}{
			"title":       title,
			"body":        message,
			"token":       device.DeviceID,
			"icon":        icon,
			"type":        notification.Type,
			"hostUuid":    notification.HostUuid,
			"serviceUuid": notification.ServiceUuid,
			"userId":      userId,
			"auth":        authKey,
		}

		go func(deviceId string) {
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Failed to marshal push data for device %s: %v", deviceId, err)
				return
			}

			req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Failed to create request for device %s: %v", deviceId, err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Api-Key", authKey)

			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to send push to device %s: %v", deviceId, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusGone { // 410
				// Remove device from DB if gone
				err := w.db.DeleteMobileDeviceByDeviceID(w.appCtx, deviceId)
				if err != nil {
					log.Printf("Failed to delete device %s: %v", deviceId, err)
				}
			}
		}(device.DeviceID)
	}
}
