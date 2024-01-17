//go:build geo
// +build geo

package action

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/akhenakh/coord2country"
	"github.com/peterstace/simplefeatures/geom"

	"github.com/akhenakh/ovr/tools"
)

var geoActions = []Action{
	toGeoJSONAction, fromGeoJSONAction, fromWKTAction, toWKTAction, centroidAction, geojsonioAction,
	countryAction,
}

func init() {
	r := DefaultRegistry()

	r.RegisterActions(geoActions...)
}

var toGeoJSONAction = Action{
	Doc:          "Transforms a geometry to GeoJSON",
	Names:        []string{"togeojson"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(in any) (any, error) {
		return json.Marshal(in.(geom.Geometry))
	},
}

var fromGeoJSONAction = Action{
	Doc:          "Parse a GeoJSON into a Geometry",
	Names:        []string{"geojson"},
	Type:         ParseAction,
	InputFormat:  TextFormat,
	OutputFormat: GeoFormat,
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
	InputFormat:  TextFormat,
	OutputFormat: GeoFormat,
	Func: func(in any) (any, error) {
		return geom.UnmarshalWKT(string(in.([]byte)))
	},
}

var toWKTAction = Action{
	Doc:          "Transforms a geometry to WKT",
	Names:        []string{"towkt"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(in any) (any, error) {
		return []byte(in.(geom.Geometry).AsText()), nil
	},
}

var centroidAction = Action{
	Doc:          "Output the centroid of a geometry",
	Names:        []string{"centroid"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(in any) (any, error) {
		return []byte(in.(geom.Geometry).Centroid().AsText()), nil
	},
}

var geojsonioAction = Action{
	Doc:          "Open a browser to geojson.io with the geometry",
	Names:        []string{"geojsonio"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(in any) (any, error) {
		geojson, err := json.Marshal(in.(geom.Geometry))
		if err != nil {
			return nil, err
		}

		query := url.QueryEscape(string(geojson))

		tools.OpenBrowser(fmt.Sprintf("http://geojson.io/#data=data:application/json,%s", query))
		return geojson, nil
	},
}

var countryAction = Action{
	Doc:          "Returns the centroid's country of the geometry",
	Names:        []string{"country"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(in any) (any, error) {
		xy, ok := in.(geom.Geometry).Centroid().XY()
		if !ok {
			return nil, fmt.Errorf("no coordinates for centroid")
		}

		idx, err := coord2country.OpenIndex()
		if err != nil {
			return nil, err
		}
		resp := idx.Query(xy.Y, xy.X)
		countries := make([]string, len(resp))
		for i, l := range resp {
			countries[i] = l.Name
		}
		return []byte(strings.Join(countries, ",")), nil
	},
}
