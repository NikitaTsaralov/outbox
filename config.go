package outbox

type Config struct {
	MessageRelay     MessageRelayConfig
	GarbageCollector GarbageCollectorConfig
}
