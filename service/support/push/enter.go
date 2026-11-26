// Package push is the offline delivery path for support chat. The SSE fanout reaches users who
// are connected; this package consumes the same Kafka topic in a single shared consumer group and
// sends a system push notification to the recipients whose device is not currently connected.
package push

// ServiceGroup groups the push services; embedded into service/support.ServiceGroup.
type ServiceGroup struct {
	PushServices
}
