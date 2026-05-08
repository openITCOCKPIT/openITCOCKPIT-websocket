package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

type PostMessage struct {
	Timestamp        time.Time `json:"timestamp"`
	UserId           int64     `json:"userId"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	Type             string    `json:"type"`
	HostUuid         string    `json:"hostUuid"`
	ServiceUuid      string    `json:"serviceUuid"`
	Icon             string    `json:"icon"`
	State            int       `json:"state"`
	NotificationType string    `json:"notificationtype,omitempty"`
	Output           string    `json:"output,omitempty"`
	AckAuthor        string    `json:"ackauthor,omitempty"`
	AckComment       string    `json:"ackcomment,omitempty"`
}

func main() {
	// Define flags
	msgType := flag.String("type", "", "Type of the notification: host or service")
	msgTypeShort := flag.String("t", "", "Short for --type")
	notificationType := flag.String("notificationtype", "", "Notification type of monitoring engine")
	hostUuid := flag.String("hostuuid", "", "Host uuid you want to send a notification")
	serviceUuid := flag.String("serviceuuid", "", "Service uuid you want to send a notification")
	state := flag.Int("state", 0, "Current host/service state")
	output := flag.String("output", "", "Host/service output")
	ackAuthor := flag.String("ackauthor", "", "Acknowledgement author")
	ackComment := flag.String("ackcomment", "", "Acknowledgement comment")
	userId := flag.Int64("user-id", 0, "openITCOCKPIT User Id")
	title := flag.String("title", "", "Title of the notification")
	message := flag.String("message", "", "Message body")
	icon := flag.String("icon", "", "Icon path")
	url := flag.String("url", "http://localhost:8083/message", "Push notification endpoint URL")

	flag.Parse()

	finalType := *msgType
	if finalType == "" {
		finalType = *msgTypeShort
	}

	if *userId <= 0 {
		fmt.Fprintln(os.Stderr, "user-id is required and must be > 0")
		os.Exit(1)
	}

	postMsg := PostMessage{
		Timestamp:        time.Now().UTC(),
		UserId:           *userId,
		Title:            *title,
		Message:          *message,
		Type:             *msgType,
		HostUuid:         *hostUuid,
		ServiceUuid:      *serviceUuid,
		Icon:             *icon,
		State:            *state,
		NotificationType: *notificationType,
		Output:           *output,
		AckAuthor:        *ackAuthor,
		AckComment:       *ackComment,
	}

	jsonData, err := json.Marshal(postMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to marshal JSON:", err)
		os.Exit(1)
	}

	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintln(os.Stderr, "HTTP request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)
}

func isAcknowledgement(notificationtype string) bool {
	return notificationtype == "ACKNOWLEDGEMENT"
}

func isFlappingStart(notificationtype string) bool {
	return notificationtype == "FLAPPINGSTART"
}

func isFlappingStop(notificationtype string) bool {
	return notificationtype == "FLAPPINGSTOP"
}

func isFlappingDisabled(notificationtype string) bool {
	return notificationtype == "FLAPPINGDISABLED"
}

func isDowntimeStart(notificationtype string) bool {
	return notificationtype == "DOWNTIMESTART"
}

func isDowntimeEnd(notificationtype string) bool {
	return notificationtype == "DOWNTIMEEND"
}

func isDowntimeCancelled(notificationtype string) bool {
	return notificationtype == "DOWNTIMECANCELLED"
}
