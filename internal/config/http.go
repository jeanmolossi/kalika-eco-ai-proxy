package config

import "time"

type HTTPServer struct {
	Host        string        `env:"HOST"         envDefault:"0.0.0.0"`
	Port        int           `env:"PORT"         envDefault:"8081"`
	BasePath    string        `env:"BASE_PATH"    envDefault:"/api/v1"`
	ReadTimeout time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
}
