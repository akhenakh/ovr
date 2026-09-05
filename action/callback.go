package action

import "fmt"

// callbackAction runs a UI callback when applied and returns the input data
// unchanged (nothing is pushed on the undo stack). It is input-independent:
// it can be rebound to every format so it is offered for any data.
type callbackAction struct {
	names []string
	doc   string
	fn    func() error
	fmt   Format
}

// NewCallbackAction creates an action that runs fn when applied. The data is
// passed through unchanged. Register it with RegisterActionForAllFormats to
// offer it no matter the data format.
func NewCallbackAction(names []string, doc string, fn func() error) Action {
	return &callbackAction{names: names, doc: doc, fn: fn, fmt: TextFormat}
}

func (c *callbackAction) Doc() string                   { return c.doc }
func (c *callbackAction) Names() []string               { return c.names }
func (c *callbackAction) Type() ActionType              { return TransformAction }
func (c *callbackAction) InputFormat() Format           { return c.fmt }
func (c *callbackAction) OutputFormat() Format          { return c.fmt }
func (c *callbackAction) Parameters() []ActionParameter { return nil }
func (c *callbackAction) InputParameters() []any        { return nil }
func (c *callbackAction) Interactive() bool             { return false }

func (c *callbackAction) SetInputParameters(values ...any) error {
	if len(values) != 0 {
		return fmt.Errorf("%s expects no parameters, got %d", c.Title(), len(values))
	}
	return nil
}

func (c *callbackAction) Transform(in *Data) (*Data, error) {
	if err := c.fn(); err != nil {
		return nil, err
	}
	return in, nil
}

func (c *callbackAction) Title() string           { return capitalize(c.names[0]) }
func (c *callbackAction) Description() string     { return c.doc }
func (c *callbackAction) FullDescription() string { return c.Title() + ": " + c.doc }
func (c *callbackAction) FilterValue() string     { return c.Title() + " " + c.doc }

func (c *callbackAction) clone() Action {
	cp := *c
	return &cp
}

// rebind to another format: the callback ignores its input, so a shallow
// clone with the new format is enough
func (c *callbackAction) rebind(f Format) Action {
	cp := *c
	cp.fmt = f
	return &cp
}
