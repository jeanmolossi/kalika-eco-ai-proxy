package config

type Log struct {
	Level string `env:"LEVEL" envDefault:"info"`
}
