package action

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ActionRegistry struct {
	m map[string]*Action
}

var all = []Action{
	upperAction, lowerAction, titleAction,
	md5HashAction, sha1HashAction, sha256HashAction, sha512HashAction,
	toHexStringAction, fromHexStringAction, toBase64StringAction, fromBase64StringAction,
	parseJSONDateStringAction,
}

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
		key := a.InputFormat.Prefix + "," + name
		r.m[key] = &a
	}
}

func (r *ActionRegistry) TextAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[textFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for text input", action)
	}

	ab, err := a.Func(in)
	return ab.([]byte), err
}

func (r *ActionRegistry) BinAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[binFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for bin input", action)
	}
	ab, err := a.Func(in)
	return ab.([]byte), err
}

func (r *ActionRegistry) TextTimeAction(action string, in []byte) (time.Time, error) {
	a, ok := r.m[textFormat.Prefix+","+action]
	if !ok {
		return time.Time{}, fmt.Errorf("action %s does not exist for text input", action)
	}
	ab, err := a.Func(in)
	return ab.(time.Time), err
}

func (r *ActionRegistry) TimeAction(action string, in time.Time) (time.Time, error) {
	a, ok := r.m[timeFormat.Prefix+","+action]
	if !ok {
		return time.Time{}, fmt.Errorf("action %s does not exist for time input", action)
	}
	ab, err := a.Func(in)
	return ab.(time.Time), err
}

// ActionsForText returns a list of actions, prefix by search, all if search is empty
// ordered alphabetically
func (r *ActionRegistry) ActionsForText(search string) (actions []*Action) {
	for k, a := range r.m {
		if strings.HasPrefix(k, textFormat.Prefix+",") {
			actions = append(actions, a)
		}

		sort.Slice(actions, func(i, j int) bool { return actions[i].Names[0] < actions[j].Names[0] })
	}
	return
}

// ActionsForTime returns a list of actions, prefix by search, all if search is empty
// ordered alphabetically
func (r *ActionRegistry) ActionsForTime(search string) (actions []*Action) {
	for k, a := range r.m {
		if strings.HasPrefix(k, timeFormat.Prefix+",") {
			actions = append(actions, a)
		}

		sort.Slice(actions, func(i, j int) bool { return actions[i].Names[0] < actions[j].Names[0] })
	}
	return
}
