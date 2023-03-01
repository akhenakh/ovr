package action

const (
	TransformAction ActionType = iota
	ParseAction
)

type Action struct {
	Doc          string
	Names        []string // command and aliases
	Type         ActionType
	InputFormat  Format
	OutputFormat Format
	Func         func([]byte) ([]byte, error)
}

type Data struct {
	Format           Format
	UnstructuredData []any
	StructuredData   map[string]any
}

type ActionType uint16

type Format struct {
	Name string
}

var (
	text = Format{"text"}
	bin  = Format{"bin"}
)

func (a *Action) TextTransform(in []byte) ([]byte, error) {
	return a.Func(in)
}
