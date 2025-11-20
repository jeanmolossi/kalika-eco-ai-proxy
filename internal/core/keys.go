package core

// Default keys used in the container.
// You can follow this pattern to register other modules.

const (
	ConfigModule = "core:config"
	LoggerModule = "core:logger"
	EchoModule   = "core:echo"

	TenantStoreModule    = "platform:tenantstore"
	RateLimiterModule    = "platform:ratelimiter"
	SemanticCacheModule  = "platform:semantic_cache"
	GuardrailsModule     = "platform:guardrails"
	UsagePublisherModule = "platform:usage_publisher"
	AuditPublisherModule = "platform:audit_publisher"
	RouterModule         = "platform:router"
)
