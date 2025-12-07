package core

// Default keys used in the container.
// You can follow this pattern to register other modules.

const (
	ConfigModule     = "core:config"
	LoggerModule     = "core:logger"
	EchoModule       = "core:echo"
	GRPCServerModule = "core:grpc"

	TenantStoreModule    = "tenant:store"
	RateLimiterModule    = "ratelimit:limiter"
	SemanticCacheModule  = "cache:semantic"
	GuardrailsModule     = "guardrails:engine"
	UsagePublisherModule = "observability:usage_publisher"
	AuditPublisherModule = "observability:audit_publisher"
	RouterModule         = "llm:router"
	TokenizerModule      = "llm:tokenizer" //nolint:gosec // container key identifier
)

func GRPCClientModule(name string) string {
	return "grpcclient:" + name
}
