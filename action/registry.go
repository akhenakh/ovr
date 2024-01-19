package action

import (
	"sort"
	"strings"
	"sync"
)

type ActionRegistry struct {
	m map[string]*Action
}

var (
	defaultRegistry *ActionRegistry
	registryOnce    sync.Once
)

var all = []Action{
	upperAction, lowerAction, titleAction, trimSpaceAction, quoteAction, unquoteAction,
	md5HashAction, sha1HashAction, sha256HashAction, sha512HashAction,
	toHexStringAction, fromHexStringAction, toBase64StringAction, fromBase64StringAction,
	parseJSONDateStringAction, epochTimeAction,
	estTimeAction, etTimeAction, utcTimeAction, isoTimeAction, timeEpochAction,
	commaTextListAction, jwtTextListAction, textListJoinCommaAction, jsonCompactAction,
	textListFirstAction, textListLastAction,
}

func DefaultRegistry() *ActionRegistry {
	registryOnce.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
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

// ActionByName returns an action for an exact name match
func (r *ActionRegistry) ActionByName(name string) (action *Action) {
	a, ok := r.m[name]
	if !ok {
		return nil
	}
	return a
}

// ActionsForText returns a list of actions, prefix by search, all if search is empty
// ordered alphabetically
func (r *ActionRegistry) ActionsForText(search string) (actions []*Action) {
	for k, a := range r.m {
		if strings.HasPrefix(k, TextFormat.Prefix+",") {
			actions = append(actions, a)
		}

		sort.Slice(actions, func(i, j int) bool { return actions[i].Names[0] < actions[j].Names[0] })
	}
	return
}

func (r *ActionRegistry) ActionsForData(data *Data) (actions []*Action) {
	for k, a := range r.m {
		if strings.HasPrefix(k, data.Format.Prefix+",") {
			actions = append(actions, a)
		}

		// in case we have a textList we also want to apply text filter, that can output text
		if data.Format == TextListFormat && a.InputFormat == TextFormat && a.OutputFormat == TextFormat {
			actions = append(actions, a)
		}

		sort.Slice(actions, func(i, j int) bool { return actions[i].Names[0] < actions[j].Names[0] })
	}

	return
}
