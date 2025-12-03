package llm

import "time"

// MetricsRecorder captures latency and error information for LLM calls.
type MetricsRecorder interface {
	ObserveChat(model string, latency time.Duration, err error)
	ObserveEmbed(model string, latency time.Duration, err error)
}

// NoopMetrics is a MetricsRecorder that drops all observations.
type NoopMetrics struct{}

// ObserveChat implements MetricsRecorder.
func (NoopMetrics) ObserveChat(string, time.Duration, error) {}

// ObserveEmbed implements MetricsRecorder.
func (NoopMetrics) ObserveEmbed(string, time.Duration, error) {}
