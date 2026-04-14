package service

const (
	hubMethodSendToFrontend        = "SendToFrontend"
	hubMethodSendToFrontendGlobal  = "SendToFrontendGlobal"
	hubMethodNotifyDelayBus        = "NotifyDelayBus"
	hubMethodNotifyAdminFromCamera = "NotifyAdminFromCamera"
)

// RealtimeHubMethods holds SignalR hub method names invoked by NotificationService.
type RealtimeHubMethods struct {
	SendToFrontend        string
	SendToFrontendGlobal  string
	NotifyDelayBus        string
	NotifyAdminFromCamera string
}

// DefaultRealtimeHubMethods returns hub method names matching the default server hub.
func DefaultRealtimeHubMethods() RealtimeHubMethods {
	return RealtimeHubMethods{
		SendToFrontend:        hubMethodSendToFrontend,
		SendToFrontendGlobal:  hubMethodSendToFrontendGlobal,
		NotifyDelayBus:        hubMethodNotifyDelayBus,
		NotifyAdminFromCamera: hubMethodNotifyAdminFromCamera,
	}
}
