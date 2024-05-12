package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	VKEventsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "viktig_vk_events_received"},
		[]string{"type"},
	)
	MessagesForwarded = promauto.NewCounter(
		prometheus.CounterOpts{Name: "viktig_messages_forwarded"},
	)
)
