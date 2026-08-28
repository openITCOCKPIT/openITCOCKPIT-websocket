package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"push_notification/pkg/models"
	"time"
)

// This is a simple CLI tool to send push notifications to the push_notification service.
// This command will be used by Naemon to trigger push notifications based on monitoring events.
// Usage example:
// go run send_push_notification.go --type=host --notificationtype=PROBLEM --hostuuid=1234-5678-90ab-cdef --state=2 --output="Host is down" --user-id=1
func main() {
	// Read the address from openITCOCKPIT Server from the environment, in case we run in docker
	address := "127.0.0.1"

	if envAddress := os.Getenv("OITC_ADDRESS"); envAddress != "" {
		address = envAddress
	}

	url := fmt.Sprintf("http://%s:8083/message", address)

	// Define flags
	msgType := flag.String("type", "", "Type of the notification: host or service")
	msgTypeShort := flag.String("t", "", "Short for --type")
	notificationType := flag.String("notificationtype", "", "Notification type of monitoring engine => $NOTIFICATIONTYPE$")
	hostUuid := flag.String("hostuuid", "", "Host uuid you want to send a notification => $HOSTNAME$")
	serviceUuid := flag.String("serviceuuid", "", "Service uuid you want to send a notification => $SERVICEDESC$")
	state := flag.Int("state", 0, "Current host/service state => $HOSTSTATEID$/$SERVICESTATEID$")
	output := flag.String("output", "", "Host/service output => $HOSTOUTPUT$/$SERVICEOUTPUT$")
	ackAuthor := flag.String("ackauthor", "", "Acknowledgement author => $NOTIFICATIONAUTHOR$")
	ackComment := flag.String("ackcomment", "", "Acknowledgement comment => $NOTIFICATIONCOMMENT$")
	userId := flag.Int64("user-id", 0, "openITCOCKPIT User Id")

	flag.Parse()

	finalType := *msgType
	if finalType == "" {
		finalType = *msgTypeShort
	}

	if *userId <= 0 {
		fmt.Fprintln(os.Stderr, "user-id is required and must be > 0")
		os.Exit(1)
	}

	postMsg := models.PostMessage{
		Timestamp:        time.Now().Unix(),
		UserId:           *userId,
		Type:             finalType,
		HostUuid:         *hostUuid,
		ServiceUuid:      *serviceUuid,
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

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintln(os.Stderr, "HTTP request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseData))

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}

	os.Exit(0)
}
