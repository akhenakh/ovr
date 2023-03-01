package action

import "fmt"

type ActionRegistry struct {
	m map[string]*Action
}

func NewRegistry() *ActionRegistry {
	m := make(map[string]*Action)

	for _, action := range []Action{upperAction, lowerAction, toHexStringAction, fromHexStringAction} {
		a := action
		for _, name := range a.Names {
			m[name] = &a
		}
	}

	return &ActionRegistry{
		m: m,
	}
}

func (r *ActionRegistry) TextAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", action)
	}
	if a.InputFormat != text {
		return nil, fmt.Errorf("action %s does not take text input", action)
	}
	return a.Func(in)
}

func (r *ActionRegistry) BinAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", action)
	}
	if a.InputFormat != bin {
		return nil, fmt.Errorf("action %s does not take binary input", action)
	}
	return a.Func(in)
}
