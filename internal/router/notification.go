package router

import (
	"fmt"
	"push_notification/pkg/models"
)

// getHostNotificationData generates the title, message and icon for a host notification based on the notification type and host data.
func getHostNotificationData(notification models.PostMessage, host *models.Host) (string, string, string) {
	if host == nil {
		return "", "", ""
	}

	var title, message, icon string

	if isAcknowledgement(notification.NotificationType) {
		title = fmt.Sprintf("Acknowledgement for %s (%s)", host.Name, getHostState(notification.State))
		message = fmt.Sprintf("%s: %s", notification.AckAuthor, notification.AckComment)

		return title, message, ""
	}

	if isDowntimeStart(notification.NotificationType) {
		title = fmt.Sprintf("Downtime start for %s", host.Name)
		message = "Host has entered a period of scheduled downtime"

		return title, message, ""
	}

	if isDowntimeEnd(notification.NotificationType) {
		title = fmt.Sprintf("Downtime end for %s", host.Name)
		message = "Host has exited a period of scheduled downtime"

		return title, message, ""
	}

	if isDowntimeCancelled(notification.NotificationType) {
		title = fmt.Sprintf("Downtime cancelled for %s", host.Name)
		message = "Scheduled downtime for host has been cancelled"

		return title, message, ""
	}

	if isFlappingStart(notification.NotificationType) {
		title = fmt.Sprintf("Flapping start for %s", host.Name)
		message = "Host appears to have started flapping"

		return title, message, ""
	}

	if isFlappingStop(notification.NotificationType) {
		title = fmt.Sprintf("Flapping stop for %s", host.Name)
		message = "Host appears to have stopped flapping"

		return title, message, ""
	}

	if isFlappingDisabled(notification.NotificationType) {
		title = fmt.Sprintf("Flapping disabled for %s", host.Name)
		message = "Flapping has been disabled for the host"

		return title, message, ""
	}

	// Default host notification
	title = fmt.Sprintf("%s: %s is %s!", notification.NotificationType, host.Name, getHostState(notification.State))
	message = notification.Output
	icon = getHostNotificationIcon(notification.State)

	return title, message, icon
}

// getServiceNotificationData generates the title, message and icon for a service notification based on the notification type and service data.
func getServiceNotificationData(notification models.PostMessage, host *models.Host, service *models.Service) (string, string, string) {
	if service == nil {
		return "", "", ""
	}

	var title, message, icon string

	if isAcknowledgement(notification.NotificationType) {
		title = fmt.Sprintf("Acknowledgement for %s/%s (%s)", host.Name, service.ServiceName, getServiceState(notification.State))
		message = fmt.Sprintf("%s: %s", notification.AckAuthor, notification.AckComment)

		return title, message, ""
	}

	if isDowntimeStart(notification.NotificationType) {
		title = fmt.Sprintf("Downtime start for %s/%s", host.Name, service.ServiceName)
		message = "Service has entered a period of scheduled downtime"

		return title, message, ""
	}

	if isDowntimeEnd(notification.NotificationType) {
		title = fmt.Sprintf("Downtime end for %s/%s", host.Name, service.ServiceName)
		message = "Service has exited a period of scheduled downtime"

		return title, message, ""
	}

	if isDowntimeCancelled(notification.NotificationType) {
		title = fmt.Sprintf("Downtime cancelled for %s/%s", host.Name, service.ServiceName)
		message = "Scheduled downtime for service has been cancelled"

		return title, message, ""
	}

	if isFlappingStart(notification.NotificationType) {
		title = fmt.Sprintf("Flapping start for %s/%s", host.Name, service.ServiceName)
		message = "Service appears to have started flapping"

		return title, message, ""
	}

	if isFlappingStop(notification.NotificationType) {
		title = fmt.Sprintf("Flapping stop for %s/%s", host.Name, service.ServiceName)
		message = "Service appears to have stopped flapping"

		return title, message, ""
	}

	if isFlappingDisabled(notification.NotificationType) {
		title = fmt.Sprintf("Flapping disabled for %s/%s", host.Name, service.ServiceName)
		message = "Flapping has been disabled for the service"

		return title, message, ""
	}

	// Default service notification
	title = fmt.Sprintf("%s: %s/%s is %s!", notification.NotificationType, host.Name, service.ServiceName, getServiceState(notification.State))
	message = notification.Output
	icon = getServiceNotificationIcon(notification.State)

	return title, message, icon
}

// getServiceNotificationData generates the title and message for a service notification based on the notification type and service data.
func isAcknowledgement(notificationType string) bool {
	return notificationType == "ACKNOWLEDGEMENT"
}

// isFlappingStart checks if the notification type indicates that flapping has started for the host or service.
func isFlappingStart(notificationType string) bool {
	return notificationType == "FLAPPINGSTART"
}

// isFlappingStop checks if the notification type indicates that flapping has stopped for the host or service.
func isFlappingStop(notificationType string) bool {
	return notificationType == "FLAPPINGSTOP"
}

// isFlappingDisabled checks if the notification type indicates that flapping has been disabled for the host or service.
func isFlappingDisabled(notificationType string) bool {
	return notificationType == "FLAPPINGDISABLED"
}

// isDowntimeStart checks if the notification type indicates that a scheduled downtime has started.
func isDowntimeStart(notificationType string) bool {
	return notificationType == "DOWNTIMESTART"
}

// isDowntimeEnd checks if the notification type indicates that a scheduled downtime has ended.
func isDowntimeEnd(notificationType string) bool {
	return notificationType == "DOWNTIMEEND"
}

// isDowntimeCancelled checks if the notification type indicates that a scheduled downtime has been cancelled.
func isDowntimeCancelled(notificationType string) bool {
	return notificationType == "DOWNTIMECANCELLED"
}

// getHostState returns the human-readable state string based on the host state ID.
func getHostState(stateID int) string {
	switch stateID {
	case 0:
		return "UP"
	case 1:
		return "DOWN"
	default:
		return "UNREACHABLE"
	}
}

// getHostNotificationIcon returns the appropriate icon path based on the host state ID.
func getHostNotificationIcon(stateID int) string {
	switch stateID {
	case 0:
		return "/img/push_notifications/wh/HostPushIconUP.png"
	case 1:
		return "/img/push_notifications/wh/HostPushIconDOWN.png"
	case 3:
		return "/img/push_notifications/wh/HostPushIconUNREACHABLE.png"
	}

	// No icon
	return ""
}

// getServiceState returns the human-readable state string based on the service state ID.
func getServiceState(stateID int) string {
	switch stateID {
	case 0:
		return "OK"
	case 1:
		return "WARNING"
	case 2:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// getServiceNotificationIcon returns the appropriate icon path based on the service state ID.
func getServiceNotificationIcon(stateID int) string {
	switch stateID {
	case 0:
		return "/img/push_notifications/wh/ServicePushIconOK.png"
	case 1:
		return "/img/push_notifications/wh/ServicePushIconWARNING.png"
	case 2:
		return "/img/push_notifications/wh/ServicePushIconCRITICAL.png"
	case 3:
		return "/img/push_notifications/wh/ServicePushIconUNKNOWN.png"
	}

	// No icon
	return ""
}
