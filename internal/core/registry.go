package core

// Registry é uma lista de módulos registrados na aplicação.
type Registry interface {
	All() []Module
}

// staticRegistry é uma implementação simples baseada em slice.
type staticRegistry struct {
	mods []Module
}

func (r staticRegistry) All() []Module { return r.mods }

// NewRegistry cria um Registry imutável com os módulos informados.
func NewRegistry(mods ...Module) Registry {
	return staticRegistry{mods: mods}
}
