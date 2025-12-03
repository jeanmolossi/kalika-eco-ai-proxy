package config

// ChoosePostgres returns the primary Postgres configuration when it is
// populated, otherwise falls back to the default configuration.
func ChoosePostgres(primary, fallback Postgres) Postgres {
	if primary.DSN != "" || primary.Database.Database != "" {
		return primary
	}

	return fallback
}
