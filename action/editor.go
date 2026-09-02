package action

import "fmt"

// Editor is implemented by actions that edit the entry with an external
// editor. UIs should detect them before calling Transform: they suspend
// rendering, run the editor with EditContent and feed the edited content
// back as the action input parameter.
type Editor interface {
	Action

	// EditContent returns the entry as the string to edit,
	// non text entries are converted to their string representation first.
	EditContent(d *Data) string
}

// editAction stores the content edited in the external editor as its input
// parameter, Transform replays it so the action stays undoable without
// launching the editor again.
type editorAction struct {
	*erased[[]byte, []byte]
}

func (a *editorAction) EditContent(d *Data) string {
	if d.Format == TextFormat {
		return string(d.RawValue)
	}
	return d.String()
}

func (a *editorAction) Transform(in *Data) (*Data, error) {
	if len(a.InputParameters()) != len(a.def.Parameters) {
		return nil, fmt.Errorf("input parameters required for %s", a.Title())
	}
	p, ok := a.InputParameters()[0].(string)
	if !ok {
		return nil, fmt.Errorf("edit parameter is not a string")
	}
	return in.StoreTextValue([]byte(p), a), nil
}

func (a *editorAction) clone() Action {
	return &editorAction{erased: a.erased.clone().(*erased[[]byte, []byte])}
}

func (a *editorAction) rebind(f Format) Action {
	c := *a.erased
	c.def.InputFormat = f
	return &editorAction{erased: &c}
}

var editAction = &editorAction{erased: &erased[[]byte, []byte]{
	def: Definition[[]byte, []byte]{
		Doc:          "Edit the entry in $EDITOR, non text entries are converted to text first",
		Names:        []string{"edit"},
		Type:         TransformAction,
		InputFormat:  TextFormat,
		OutputFormat: TextFormat,
		Parameters:   []ActionParameter{{StringParameter, "edited content"}},
	},
}}
