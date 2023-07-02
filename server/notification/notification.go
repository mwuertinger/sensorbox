package notification

import (
	"log"
	"sync"
	"time"

	"github.com/mwuertinger/sensorbox/server/ntfy"
)

type Notification interface {
	SendNotification(notificationType, msg string) error
}

type notification struct {
	ntfyClient        ntfy.NtfyClient
	notificationDelay time.Duration
	clock             func() time.Time
	mu                sync.Mutex // protects everything below
	lastNotification  map[string]time.Time
}

func (n *notification) SendNotification(notificationType, msg string) error {
	n.mu.Lock()
	if n.lastNotification[notificationType].After(n.clock().Add(-n.notificationDelay)) {
		n.mu.Unlock()
		log.Printf("SendNotification: last notification too recent for %s", notificationType)
		return nil
	}
	n.lastNotification[notificationType] = n.clock()
	n.mu.Unlock()
	return n.ntfyClient.SendNotification(msg)
}
