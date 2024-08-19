package object

func NewClosedEnv(outer *Env) *Env {
	env := NewEnvironment()
	env.outer = outer

	return env
}

func NewEnvironment() *Env {
	s := make(map[string]Object)
	return &Env{store: s}
}

type Env struct {
	store map[string]Object
	outer *Env
}

func (e *Env) Get(name string) (Object, bool) {
	obj, ok := e.store[name]

	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}

	return obj, ok
}

func (e *Env) Set(name string, val Object) Object {
	e.store[name] = val

	return val
}
