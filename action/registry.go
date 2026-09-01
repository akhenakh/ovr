package action

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type ActionRegistry struct {
	m map[string]Action
}

var (
	defaultRegistry *ActionRegistry
	registryOnce    sync.Once
)

var all = []Action{
	upperAction, lowerAction, titleAction, trimSpaceAction, quoteAction, unquoteAction, calcAction,
	md5HashAction, sha1HashAction, sha256HashAction, sha512HashAction, crc32HashAction, hmacSha256Action,
	toHexStringAction, fromHexStringAction, toBase64StringAction, fromBase64StringAction,
	parseDateStringAction, epochTimeAction, addDurationTimeAction,
	tzTimeAction, isoTimeAction, toJSONDateStringAction, timeEpochAction,
	spaceTextListAction, pipeTextListAction, splitTextListAction,
	commaTextListAction, textListJoinNewLineAction, jwtTextListAction, textListJoinCommaAction, textListCharJoinAction, jsonCompactAction, jsonPrettifyAction,
	textListSortAction, textListReverseAction,
	textListFirstAction, textListLastAction, textListIndexAction, textListCountAction, unescapeTextAction, stripWhitespaceAction, pipeCommandAction,
	parseCSVAction, tableSortColumnAction, tableToCSVAction,
}

func init() {
	all = slices.Concat(all, timezoneActions)
}

// generatorActions ignore their input and can be applied to any data
var generatorActions = []Action{uuidV4Action, uuidV7Action, nowTimeAction, repeatLastAction, newCalcAction}

// inputFormats is every format generator actions are registered for,
// text list data already offers text to text actions so it is not needed there
var inputFormats = []Format{TextFormat, BinFormat, TimeFormat, JSONFormat, GeoFormat}

// rebinder is implemented by actions that can register for another input format
type rebinder interface {
	rebind(Format) Action
}

func DefaultRegistry() *ActionRegistry {
	registryOnce.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
}

func NewRegistry() *ActionRegistry {
	m := make(map[string]Action)
	r := &ActionRegistry{
		m: m,
	}

	r.RegisterActions(cloneActions(all)...)

	for _, a := range generatorActions {
		rb, ok := a.(rebinder)
		if !ok {
			panic(fmt.Sprintf("generator action %s can't be registered for every format", a.Title()))
		}
		for _, f := range inputFormats {
			r.RegisterAction(rb.rebind(f))
		}
	}

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
	for _, name := range a.Names() {
		key := a.InputFormat().Prefix + "," + name
		if _, exist := r.m[key]; exist {
			panic(fmt.Sprintf("registering action conflict for %s", key))
		}
		r.m[key] = a
	}
}

// ActionByName returns an action for an exact name match
func (r *ActionRegistry) ActionByName(format Format, name string) (Action, bool) {
	a, ok := r.m[format.Prefix+","+name]
	return a, ok
}

// MustActionByName returns an action for an exact name match, panics otherwise
func (r *ActionRegistry) MustActionByName(format Format, name string) Action {
	a, ok := r.ActionByName(format, name)
	if !ok {
		panic(fmt.Sprintf("no action %s,%s", format.Prefix, name))
	}
	return a
}

// ActionsForText returns a list of actions, prefix by search, all if search is empty
// ordered alphabetically
func (r *ActionRegistry) ActionsForText(search string) (actions []Action) {
	added := make(map[Action]bool)
	for k, a := range r.m {
		if strings.HasPrefix(k, TextFormat.Prefix+",") {
			if !added[a] {
				actions = append(actions, a)
				added[a] = true
			}
		}
	}
	sortActions(actions)
	return
}

func (r *ActionRegistry) ActionsForData(data *Data) (actions []Action) {
	added := make(map[Action]bool)
	for k, a := range r.m {
		isApplicable := false

		// Action's input format matches the data's format
		if strings.HasPrefix(k, data.Format.Prefix+",") {
			isApplicable = true
		} else if data.Format == TextListFormat && a.InputFormat() == TextFormat && a.OutputFormat() == TextFormat {
			// Special case: text-to-text actions can be applied to each item of a text list
			isApplicable = true
		}

		if isApplicable {
			if !added[a] {
				actions = append(actions, a)
				added[a] = true
			}
		}
	}

	sortActions(actions)

	return
}

func sortActions(actions []Action) {
	slices.SortFunc(actions, func(a, b Action) int {
		return strings.Compare(a.Names()[0], b.Names()[0])
	})
}

func cloneActions(actions []Action) []Action {
	out := make([]Action, len(actions))
	for i, a := range actions {
		out[i] = a.clone()
	}
	return out
}
