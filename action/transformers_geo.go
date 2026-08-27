//go:build geo

package action

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/akhenakh/coord2country"
	"github.com/akhenakh/goh3"
	"github.com/golang/geo/s2"
	"github.com/peterstace/simplefeatures/geom"

	"github.com/akhenakh/ovr/tools"
)

var geoActions = []Action{
	toGeoJSONAction, fromGeoJSONAction, fromWKTAction, toWKTAction, centroidAction, geojsonioAction,
	countryAction, s2Action, h3Action, mapAction,
}

func init() {
	r := DefaultRegistry()

	r.RegisterActions(cloneActions(geoActions)...)
}

var toGeoJSONAction = New(Definition[geom.Geometry, []byte]{
	Doc:          "Transforms a geometry to GeoJSON",
	Names:        []string{"togeojson"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		return json.Marshal(in)
	},
})

var fromGeoJSONAction = New(Definition[[]byte, geom.Geometry]{
	Doc:          "Parse a GeoJSON into a Geometry",
	Names:        []string{"geojson"},
	Type:         ParseAction,
	InputFormat:  TextFormat,
	OutputFormat: GeoFormat,
	Func: func(a Action, in []byte) (geom.Geometry, error) {
		var g geom.Geometry
		err := json.Unmarshal(in, &g)
		return g, err
	},
})

var fromWKTAction = New(Definition[[]byte, geom.Geometry]{
	Doc:          "Parse a WKT into a Geometry",
	Names:        []string{"wkt"},
	Type:         ParseAction,
	InputFormat:  TextFormat,
	OutputFormat: GeoFormat,
	Func: func(a Action, in []byte) (geom.Geometry, error) {
		return geom.UnmarshalWKT(string(in))
	},
})

var toWKTAction = New(Definition[geom.Geometry, []byte]{
	Doc:          "Transforms a geometry to WKT",
	Names:        []string{"towkt"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		return []byte(in.AsText()), nil
	},
})

var centroidAction = New(Definition[geom.Geometry, []byte]{
	Doc:          "Output the centroid of a geometry",
	Names:        []string{"centroid"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		return []byte(in.Centroid().AsText()), nil
	},
})

var geojsonioAction = New(Definition[geom.Geometry, []byte]{
	Doc:          "Open a browser to geojson.io with the geometry",
	Names:        []string{"geojsonio"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		geojson, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}

		query := url.QueryEscape(string(geojson))

		tools.OpenBrowser(fmt.Sprintf("http://geojson.io/#data=data:application/json,%s", query))
		return geojson, nil
	},
})

var countryAction = New(Definition[geom.Geometry, []byte]{
	Doc:          "Returns the centroid's country of the geometry",
	Names:        []string{"country"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		xy, ok := in.Centroid().XY()
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
})

var s2Action = New(Definition[geom.Geometry, []byte]{
	Doc:          "Output the s2 cover of a geometry",
	Names:        []string{"s2cover"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Parameters: []ActionParameter{
		{IntParameter, "min level"},
		{IntParameter, "max level"},
		{IntParameter, "max cells"},
	},

	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		minLevel := a.InputParameters()[0].(int)
		maxLevel := a.InputParameters()[1].(int)
		maxCells := a.InputParameters()[2].(int)

		tokens, err := s2CoverTokens(in, minLevel, maxLevel, maxCells)
		if err != nil {
			return nil, err
		}

		return []byte(strings.Join(tokens, ",")), nil
	},
})

var h3Action = New(Definition[geom.Geometry, []byte]{
	Doc:          "Output the h3 cells of the geometry points at the given level",
	Names:        []string{"h3cover"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{IntParameter, "level of the cells"}},

	Func: func(a Action, in geom.Geometry) ([]byte, error) {
		res := a.InputParameters()[0].(int)

		tokens, err := h3CoverTokens(in, res)
		if err != nil {
			return nil, err
		}

		return []byte(strings.Join(tokens, ",")), nil
	},
})

var mapAction = New(Definition[geom.Geometry, geom.Geometry]{
	Doc:          "Display the geometry on an interactive terminal map",
	Names:        []string{"map"},
	Type:         TransformAction,
	InputFormat:  GeoFormat,
	OutputFormat: GeoFormat,
	Interactive:  true,

	Func: func(a Action, in geom.Geometry) (geom.Geometry, error) {
		return in, nil
	},
})

func s2CoverTokens(g geom.Geometry, minLevel, maxLevel, maxCells int) ([]string, error) {
	if maxCells <= 0 {
		maxCells = 8
	}
	rc := s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel, MaxCells: maxCells}

	var tokens []string
	err := forEachGeometry(g, func(g geom.Geometry) error {
		region, err := s2Region(g)
		if err != nil {
			return err
		}
		if region == nil {
			return nil
		}
		for _, c := range rc.Covering(region) {
			tokens = append(tokens, c.ToToken())
		}
		return nil
	})

	return tokens, err
}

func s2Region(g geom.Geometry) (s2.Region, error) {
	switch {
	case g.IsPoint():
		p, ok := g.AsPoint()
		if !ok {
			return nil, fmt.Errorf("not a point")
		}
		xy, ok := p.XY()
		if !ok {
			return nil, fmt.Errorf("no coordinates for point")
		}
		return s2.PointFromLatLng(s2.LatLngFromDegrees(xy.Y, xy.X)), nil

	case g.IsLineString():
		ls, ok := g.AsLineString()
		if !ok {
			return nil, fmt.Errorf("not a linestring")
		}
		seq := ls.Coordinates()
		pts := make([]s2.Point, 0, seq.Length())
		for i := range seq.Length() {
			xy := seq.GetXY(i)
			pts = append(pts, s2.PointFromLatLng(s2.LatLngFromDegrees(xy.Y, xy.X)))
		}
		p := s2.Polyline(pts)
		return &p, nil

	case g.IsPolygon():
		poly, ok := g.AsPolygon()
		if !ok {
			return nil, fmt.Errorf("not a polygon")
		}
		loop := s2Loop(poly.ExteriorRing())
		if loop.Area() > 2*math.Pi {
			loop.Invert()
		}
		return loop, nil

	default:
		return nil, fmt.Errorf("unsupported geometry type for s2 cover: %s", g.Type())
	}
}

func s2Loop(ring geom.LineString) *s2.Loop {
	seq := ring.Coordinates()
	n := seq.Length()
	pts := make([]s2.Point, 0, n)
	for i := range n {
		// drop the closing vertex, s2 loops are not closed
		if i == n-1 && n > 1 && seq.GetXY(i) == seq.GetXY(0) {
			continue
		}
		xy := seq.GetXY(i)
		pts = append(pts, s2.PointFromLatLng(s2.LatLngFromDegrees(xy.Y, xy.X)))
	}
	return s2.LoopFromPoints(pts)
}

func h3CoverTokens(g geom.Geometry, res int) ([]string, error) {
	var tokens []string
	seen := make(map[h3.Cell]bool)
	add := func(xy geom.XY) {
		c := h3.LatLngToCell(h3.LatLng{Lat: xy.Y, Lng: xy.X}, res)
		if !seen[c] {
			seen[c] = true
			tokens = append(tokens, c.String())
		}
	}

	err := forEachGeometry(g, func(g geom.Geometry) error {
		switch {
		case g.IsPoint():
			p, ok := g.AsPoint()
			if !ok {
				return fmt.Errorf("not a point")
			}
			xy, ok := p.XY()
			if !ok {
				return fmt.Errorf("no coordinates for point")
			}
			add(xy)

		case g.IsLineString():
			ls, ok := g.AsLineString()
			if !ok {
				return fmt.Errorf("not a linestring")
			}
			seq := ls.Coordinates()
			for i := range seq.Length() {
				add(seq.GetXY(i))
			}

		case g.IsPolygon():
			poly, ok := g.AsPolygon()
			if !ok {
				return fmt.Errorf("not a polygon")
			}
			for _, ring := range poly.DumpRings() {
				seq := ring.Coordinates()
				for i := range seq.Length() {
					add(seq.GetXY(i))
				}
			}
			if xy, ok := poly.Centroid().XY(); ok {
				add(xy)
			}

		default:
			return fmt.Errorf("unsupported geometry type for h3 cover: %s", g.Type())
		}
		return nil
	})

	return tokens, err
}

func forEachGeometry(g geom.Geometry, fn func(geom.Geometry) error) error {
	switch {
	case g.IsGeometryCollection():
		gc, ok := g.AsGeometryCollection()
		if !ok {
			return fmt.Errorf("not a geometry collection")
		}
		for i := range gc.NumGeometries() {
			if err := forEachGeometry(gc.GeometryN(i), fn); err != nil {
				return err
			}
		}
		return nil
	case g.IsMultiPoint():
		mp, ok := g.AsMultiPoint()
		if !ok {
			return fmt.Errorf("not a multipoint")
		}
		for i := range mp.NumPoints() {
			if err := fn(mp.PointN(i).AsGeometry()); err != nil {
				return err
			}
		}
		return nil
	case g.IsMultiLineString():
		mls, ok := g.AsMultiLineString()
		if !ok {
			return fmt.Errorf("not a multilinestring")
		}
		for i := range mls.NumLineStrings() {
			if err := fn(mls.LineStringN(i).AsGeometry()); err != nil {
				return err
			}
		}
		return nil
	case g.IsMultiPolygon():
		mp, ok := g.AsMultiPolygon()
		if !ok {
			return fmt.Errorf("not a multipolygon")
		}
		for i := range mp.NumPolygons() {
			if err := fn(mp.PolygonN(i).AsGeometry()); err != nil {
				return err
			}
		}
		return nil
	default:
		return fn(g)
	}
}
