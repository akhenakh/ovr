//go:build geo
// +build geo

package action

import (
	"encoding/json"

	"github.com/peterstace/simplefeatures/geom"
)

var geoActions = []Action{
	toGeoJSONAction, fromGeoJSONAction, fromWKTAction, toWKTAction,
}

func init() {
	r := DefaultRegistry()

	r.RegisterActions(geoActions...)
}

var toGeoJSONAction = Action{
	Doc:          "Transforms a geometry to GeoJSON",
	Names:        []string{"togeojson"},
	Type:         TransformAction,
	InputFormat:  geoFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return json.Marshal(in.(geom.Geometry))
	},
}

var fromGeoJSONAction = Action{
	Doc:          "Parse a GeoJSON into a Geometry",
	Names:        []string{"geojson"},
	Type:         ParseAction,
	InputFormat:  textFormat,
	OutputFormat: geoFormat,
	Func: func(in any) (any, error) {
		var g geom.Geometry
		err := json.Unmarshal(in.([]byte), &g)
		return g, err
	},
}

var fromWKTAction = Action{
	Doc:          "Parse a WKT into a Geometry",
	Names:        []string{"wkt"},
	Type:         ParseAction,
	InputFormat:  textFormat,
	OutputFormat: geoFormat,
	Func: func(in any) (any, error) {
		return geom.UnmarshalWKT(string(in.([]byte)))
	},
}

var toWKTAction = Action{
	Doc:          "Transforms a geometry to WKT",
	Names:        []string{"towkt"},
	Type:         TransformAction,
	InputFormat:  geoFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(in.(geom.Geometry).AsText()), nil
	},
}
