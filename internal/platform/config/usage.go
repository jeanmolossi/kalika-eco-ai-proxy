package config

// UsageSink controls where usage events are published.
type UsageSink struct {
	Mode     string `env:"MODE"      envDefault:"file"`
	FilePath string `env:"FILE_PATH" envDefault:"logs/usage-events.log"`
	Topic    string `env:"TOPIC"     envDefault:"ai-proxy.usage.events"`
}
