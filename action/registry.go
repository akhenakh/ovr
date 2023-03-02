package action

import (
	"fmt"
	"sort"
	"strings"
)

type ActionRegistry struct {
	m map[string]*Action
}

var all = []Action{upperAction, lowerAction, toHexStringAction, fromHexStringAction, toBase64StringAction, fromBase64StringAction}

func NewRegistry() *ActionRegistry {
	m := make(map[string]*Action)
	r := &ActionRegistry{
		m: m,
	}

	r.RegisterActions(all...)

	return r
}

// RegisterActions registers multiple actions by their input format, names
func (r *ActionRegistry) RegisterActions(actions ...Action) {
	for _, a := range actions {
		r.RegisterAction(a)
	}
}

// RegisterAction registers an action by its input , names
func (r *ActionRegistry) RegisterAction(a Action) {
	for _, name := range a.Names {
		key := a.InputFormat.Name + "," + name
		r.m[key] = &a
	}
}

func (r *ActionRegistry) TextAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[text.Name+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for text input", action)
	}
	return a.Func(in)
}

func (r *ActionRegistry) BinAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for bin input", action)
	}
	return a.Func(in)
}

// ActionsForText returns a list of actions, prefix by search, all if search is empty
// ordered alphabetically
func (r *ActionRegistry) ActionsForText(search string) (actions []*Action) {
	for k, a := range r.m {
		if strings.HasPrefix(k, text.Name+",") {
			actions = append(actions, a)
		}

		sort.Slice(actions, func(i, j int) bool { return actions[i].Names[0] < actions[j].Names[0] })
	}
	return
}
