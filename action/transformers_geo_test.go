//go:build geo

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

	r.RegisterActions(cloneActions(geoActions)...)

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

	r.RegisterActions(cloneActions(geoActions)...)

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

func TestAction_GeoCoverTransform(t *testing.T) {
	r := NewRegistry()

	r.RegisterActions(cloneActions(geoActions)...)

	point := `{"type":"Point","coordinates":[2.2,48.8]}`

	t.Run("s2cover point", func(t *testing.T) {
		a := r.MustActionByName(GeoFormat, "s2cover")
		require.NoError(t, a.SetInputParameters(5, 30, 8))

		g := mustGeoJSON(t, point)
		out, err := a.Transform(NewDataGeom(g))
		require.NoError(t, err)
		require.Equal(t, "47e67b949f715c6b", string(out.RawValue))
	})

	t.Run("s2cover polygon", func(t *testing.T) {
		a := r.MustActionByName(GeoFormat, "s2cover")
		require.NoError(t, a.SetInputParameters(0, 30, 8))

		g := mustGeoJSON(t, `{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,1],[1,0],[0,0]]]}`)
		out, err := a.Transform(NewDataGeom(g))
		require.NoError(t, err)
		require.NotEmpty(t, string(out.RawValue))
	})

	t.Run("h3cover point", func(t *testing.T) {
		a := r.MustActionByName(GeoFormat, "h3cover")
		require.NoError(t, a.SetInputParameters(5))

		g := mustGeoJSON(t, point)
		out, err := a.Transform(NewDataGeom(g))
		require.NoError(t, err)
		require.Equal(t, "851fb463fffffff", string(out.RawValue))
	})

	t.Run("h3cover polygon", func(t *testing.T) {
		a := r.MustActionByName(GeoFormat, "h3cover")
		require.NoError(t, a.SetInputParameters(5))

		g := mustGeoJSON(t, `{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,1],[1,0],[0,0]]]}`)
		out, err := a.Transform(NewDataGeom(g))
		require.NoError(t, err)
		require.NotEmpty(t, string(out.RawValue))
	})
}

func mustGeoJSON(t *testing.T, s string) geom.Geometry {
	t.Helper()
	var g geom.Geometry
	err := json.NewDecoder(strings.NewReader(s)).Decode(&g)
	require.NoError(t, err)
	return g
}

func (r *ActionRegistry) TextGeoAction(name string, in []byte) (geom.Geometry, error) {
	a, ok := r.ActionByName(TextFormat, name)
	if !ok {
		return geom.Geometry{}, fmt.Errorf("action %s does not exist for text input", name)
	}

	out, err := a.Transform(NewDataText(in))
	if err != nil {
		return geom.Geometry{}, err
	}
	g, ok := out.Value.(geom.Geometry)
	if !ok {
		return geom.Geometry{}, fmt.Errorf("output is not a geometry")
	}
	return g, nil
}

func (r *ActionRegistry) GeoTextAction(name string, in geom.Geometry) ([]byte, error) {
	a, ok := r.ActionByName(GeoFormat, name)
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for geo input", name)
	}

	out, err := a.Transform(NewDataGeom(in))
	if err != nil {
		return nil, err
	}
	return out.RawValue, nil
}
