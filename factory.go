package mach

type FactoryFunc func() (*Database, error)

var factories = map[string]FactoryFunc{}
var defaultFactoryName string

func RegisterFactory(name string, f FactoryFunc) {
	factories[name] = f
}

func RegisterDefaultFactory(name string, f FactoryFunc) {
	factories[name] = f
	defaultFactoryName = name
}

func NewDatabase() (*Database, error) {
	if len(defaultFactoryName) > 0 {
		if f, ok := factories[defaultFactoryName]; ok {
			return f()
		}
	}
	count := len(factories)
	if count > 0 {
		for _, f := range factories {
			return f()
		}
	}
	return nil, ErrDatabaseNoFactory
}

func NewDatabaseNamed(name string) (*Database, error) {
	if f, ok := factories[name]; ok {
		return f()
	} else {
		return nil, ErrDatabaseFactoryNotFound(name)
	}
}
