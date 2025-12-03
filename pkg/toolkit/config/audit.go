package config

// AuditSink controls where audit events are published.
type AuditSink struct {
	Mode     string `env:"MODE"      envDefault:"file"`
	FilePath string `env:"FILE_PATH" envDefault:"logs/audit-events.log"`
	Topic    string `env:"TOPIC"     envDefault:"ai-proxy.audit.events"`
}
