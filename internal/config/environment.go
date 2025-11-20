package config

type Environment string

const (
	Production  Environment = "production"
	Staging     Environment = "staging"
	Development Environment = "development"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) Production() bool {
	return e == Production
}
