//go:build geo
// +build geo

package action

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/stretchr/testify/require"
)

func TestAction_TextGeoTransform(t *testing.T) {
	r := NewRegistry()

	r.RegisterActions(geoActions...)

	tests := []struct {
		action  string
		in      string
		want    string
		wantErr bool
	}{
		{"geojson", `{"type":"Point","coordinates":[-74,40.7]}`, "Point[XY] with 1 point", false},
		{"geojson", `{"coordinates":[[[-84.4839056753157,42.33121043490431],[-84.48325592871463,42.31896267658158],[-84.46668739038469,42.31896267658158],[-84.46636251708416,42.33169069054807],[-84.4839056753157,42.33121043490431]]],"type":"Polygon"}`, "Polygon[XY] with 1 ring consisting of 5 total points", false},
		{"wkt", `POLYGON((0 0,0 1,1 1,1 0,0 0))`, "Polygon[XY] with 1 ring consisting of 5 total points", false},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextGeoAction(tt.action, []byte(tt.in))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestAction_GeoTextTransform(t *testing.T) {
	r := NewRegistry()

	r.RegisterActions(geoActions...)

	tests := []struct {
		action       string
		validGeoJSON string
		want         string
		wantErr      bool
	}{
		{
			"togeojson",
			`{"type":"Point","coordinates":[-74,40.7]}`,
			`{"type":"Point","coordinates":[-74,40.7]}`,
			false,
		},
		{
			"towkt",
			`{"type":"Point","coordinates":[-74,40.7]}`,
			`POINT(-74 40.7)`,
			false,
		},
		{
			"centroid",
			`{"type":"Point","coordinates":[-74,40.7]}`,
			"POINT(-74 40.7)",
			false,
		},

		{
			"country",
			`{"type":"Point","coordinates":[2.2,48.8]}`,
			"France",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			var g geom.Geometry
			err := json.NewDecoder(strings.NewReader(tt.validGeoJSON)).Decode(&g)
			require.NoError(t, err)

			got, err := r.GeoTextAction(tt.action, g)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, string(got))
			}
		})
	}
}

func (r *ActionRegistry) TextGeoAction(action string, in []byte) (geom.Geometry, error) {
	a, ok := r.m[textFormat.Prefix+","+action]
	if !ok {
		return geom.Geometry{}, fmt.Errorf("action %s does not exist for text input", action)
	}
	ab, err := a.Func(in)
	return ab.(geom.Geometry), err
}

func (r *ActionRegistry) GeoTextAction(action string, in geom.Geometry) ([]byte, error) {
	a, ok := r.m[geoFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for geo input", action)
	}
	ab, err := a.Func(in)
	return ab.([]byte), err
}
