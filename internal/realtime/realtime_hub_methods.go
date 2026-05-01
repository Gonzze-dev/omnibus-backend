package realtime

const (
	hubMethodSendToFrontend        = "SendToFrontend"
	hubMethodSendToFrontendGlobal  = "SendToFrontendGlobal"
	hubMethodNotifyDelayBus        = "NotifyDelayBus"
	hubMethodNotifyAdminFromCamera = "NotifyAdminFromCamera"
	hubMethodDeleteNotification    = "DeleteNotification"
)

// RealtimeHubMethods holds SignalR hub method names invoked by NotificationService.
type RealtimeHubMethods struct {
	SendToFrontend        string
	SendToFrontendGlobal  string
	NotifyDelayBus        string
	NotifyAdminFromCamera string
	DeleteNotification    string
}

// DefaultRealtimeHubMethods returns hub method names matching the default server hub.
func DefaultRealtimeHubMethods() RealtimeHubMethods {
	return RealtimeHubMethods{
		SendToFrontend:        hubMethodSendToFrontend,
		SendToFrontendGlobal:  hubMethodSendToFrontendGlobal,
		NotifyDelayBus:        hubMethodNotifyDelayBus,
		NotifyAdminFromCamera: hubMethodNotifyAdminFromCamera,
		DeleteNotification:    hubMethodDeleteNotification,
	}
}
