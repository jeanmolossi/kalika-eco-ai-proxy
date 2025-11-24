package config

// Kafka groups shared Kafka settings.
type Kafka struct {
	Brokers []string `env:"BROKERS" envSeparator:"," envDefault:"kafka:9092"`
}
