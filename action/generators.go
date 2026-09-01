package action

import (
	"fmt"
	"time"
	"uuid"
)

// newAction generates a new value, its input is ignored.
type newAction[O any] struct {
	*erased[[]byte, O]
}

func (a *newAction[O]) Transform(in *Data) (*Data, error) {
	out, err := a.def.Func(a, nil)
	if err != nil {
		return nil, err
	}
	return a.out(out, a, in)
}

func (a *newAction[O]) clone() Action {
	return &newAction[O]{erased: a.erased.clone().(*erased[[]byte, O])}
}

func (a *newAction[O]) rebind(f Format) Action {
	c := *a.erased
	c.def.InputFormat = f
	return &newAction[O]{erased: &c}
}

// newGenerator creates a generator action, the input data is ignored and replaced.
func newGenerator[O any](d Definition[[]byte, O]) *newAction[O] {
	return &newAction[O]{erased: &erased[[]byte, O]{
		def: d,
		out: outputFunc[O](d.OutputFormat),
	}}
}

var uuidV4Action = newGenerator(Definition[[]byte, []byte]{
	Doc:          "New UUID v4, create a new random UUID (version 4)",
	Names:        []string{"uuid4"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(uuid.NewV4().String()), nil
	},
})

var uuidV7Action = newGenerator(Definition[[]byte, []byte]{
	Doc:          "New UUID v7, create a new time ordered UUID (version 7)",
	Names:        []string{"uuid7"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(uuid.NewV7().String()), nil
	},
})

var nowTimeAction = newGenerator(Definition[[]byte, time.Time]{
	Doc:          "New time, create the current time (now)",
	Names:        []string{"now", "newtime"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in []byte) (time.Time, error) {
		return time.Now(), nil
	},
})

const maxRepeatCount = 10000

// repeatAction re-applies the last applied action N times,
// every result is collected in a text list.
type repeatAction struct {
	*erased[[]byte, []byte]
}

func (a *repeatAction) Transform(in *Data) (*Data, error) {
	if len(a.params) != len(a.def.Parameters) {
		return nil, fmt.Errorf("input parameters required for %s", a.Title())
	}
	n, ok := a.params[0].(int)
	if !ok {
		return nil, fmt.Errorf("repeat parameter is not an int")
	}
	if n <= 0 || n > maxRepeatCount {
		return nil, fmt.Errorf("repeat parameter must be between 1 and %d", maxRepeatCount)
	}
	if len(in.Stack) == 0 {
		return nil, fmt.Errorf("no action to repeat, apply an action first")
	}
	last := in.Stack[len(in.Stack)-1]
	if _, ok := last.(*repeatAction); ok {
		return nil, fmt.Errorf("can't repeat %s", last.Title())
	}

	l := make([]string, 0, n)
	cur := in
	for i := 0; i < n; i++ {
		nd, err := last.Transform(cur)
		if err != nil {
			return nil, err
		}
		l = append(l, nd.String())
		cur = nd
	}
	return in.StoreTextListValue(l, a), nil
}

func (a *repeatAction) clone() Action {
	return &repeatAction{erased: a.erased.clone().(*erased[[]byte, []byte])}
}

func (a *repeatAction) rebind(f Format) Action {
	c := *a.erased
	c.def.InputFormat = f
	return &repeatAction{erased: &c}
}

func newRepeatAction() *repeatAction {
	return &repeatAction{erased: &erased[[]byte, []byte]{
		def: Definition[[]byte, []byte]{
			Doc:          "Repeat the last applied action N times, every result is collected in a list",
			Names:        []string{"repeat", "times"},
			Type:         TransformAction,
			InputFormat:  TextFormat,
			OutputFormat: TextListFormat,
			Parameters:   []ActionParameter{{IntParameter, "number of times to repeat the last action"}},
		},
	}}
}

var repeatLastAction = newRepeatAction()
