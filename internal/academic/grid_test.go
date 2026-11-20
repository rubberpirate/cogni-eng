package academic

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"
)

func TestCalibratedGridSeparatesBlankAndMultipleWithoutAKey(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	c := GridCalibration{Options: 4, Radius: .006, Blocks: []GridBlock{{First: 1, Count: 3, TopLeft: Point{.1, .1}, TopRight: Point{.4, .1}, BottomLeft: Point{.12, .5}, BottomRight: Point{.42, .5}}}}
	circle(img, Point{300, 100}, 10, true)
	circle(img, Point{120, 500}, 10, true)
	circle(img, Point{420, 500}, 10, true)
	out, err := ReadGrid(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Rows) != 3 || len(out.Rows[0].Selection) != 1 || out.Rows[0].Selection[0] != 3 || out.Rows[0].Uncertain || len(out.Rows[1].Selection) != 0 || !out.Rows[1].Uncertain || len(out.Rows[2].Selection) != 2 || !out.Rows[2].Uncertain {
		t.Fatalf("response accounting %+v", out.Rows)
	}
	for _, tc := range []struct {
		name string
		cal  GridCalibration
	}{{"nan", GridCalibration{Options: 4, Radius: math.NaN(), Blocks: c.Blocks}}, {"duplicate", GridCalibration{Options: 4, Radius: .006, Blocks: append(c.Blocks, c.Blocks...)}}, {"outside", GridCalibration{Options: 4, Radius: .006, Blocks: []GridBlock{{First: 1, Count: 3, TopLeft: Point{-.1, .1}}}}}} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ReadGrid(img, tc.cal); err == nil {
				t.Fatal("unsafe calibration accepted")
			}
		})
	}
}

func TestQuestionFramesOverrideOnlyTheirRows(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	c := GridCalibration{Options: 4, Radius: .006, Area: &GridRect{X: .05, Y: .05, Width: .6, Height: .6}, Blocks: []GridBlock{{First: 5, Count: 2, TopLeft: Point{X: .1, Y: .1}, TopRight: Point{X: .4, Y: .1}, BottomLeft: Point{X: .1, Y: .5}, BottomRight: Point{X: .4, Y: .5}}}, Frames: []GridFrame{{Number: 5, Rect: GridRect{X: .194, Y: .19, Width: .312, Height: .02}}}}
	circle(img, Point{X: 300, Y: 200}, 10, true)
	circle(img, Point{X: 400, Y: 500}, 10, true)
	out, err := ReadGrid(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Rows) != 2 || len(out.Rows[0].Selection) != 1 || out.Rows[0].Selection[0] != 2 || len(out.Rows[1].Selection) != 1 || out.Rows[1].Selection[0] != 4 || out.Rows[0].Uncertain || out.Rows[1].Uncertain {
		t.Fatalf("saved frames did not control real pixels: %+v", out.Rows)
	}
	for _, tc := range []struct {
		name  string
		frame GridFrame
	}{
		{name: "unknown question", frame: GridFrame{Number: 99, Rect: c.Frames[0].Rect}},
		{name: "outside area", frame: GridFrame{Number: 5, Rect: GridRect{X: .7, Y: .2, Width: .2, Height: .03}}},
		{name: "too small", frame: GridFrame{Number: 5, Rect: GridRect{X: .2, Y: .2, Width: .01, Height: .03}}},
		{name: "not finite", frame: GridFrame{Number: 5, Rect: GridRect{X: math.NaN(), Y: .2, Width: .3, Height: .03}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := c
			bad.Frames = []GridFrame{tc.frame}
			if _, err := ReadGrid(img, bad); err == nil {
				t.Fatal("invalid frame accepted")
			}
		})
	}
	t.Run("duplicate frame", func(t *testing.T) {
		bad := c
		bad.Frames = append(append([]GridFrame{}, c.Frames...), c.Frames...)
		if _, err := ReadGrid(img, bad); err == nil {
			t.Fatal("duplicate frame accepted")
		}
	})
}

func TestSingleQuestionGridHasFiniteCenters(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	circle(img, Point{X: 100, Y: 200}, 10, true)
	c := GridCalibration{Options: 4, Radius: .006, Blocks: []GridBlock{{First: 1, Count: 1, TopLeft: Point{X: .1, Y: .2}, TopRight: Point{X: .4, Y: .2}, BottomLeft: Point{X: .1, Y: .2}, BottomRight: Point{X: .4, Y: .2}}}}
	out, err := ReadGrid(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Rows) != 1 || len(out.Rows[0].Selection) != 1 || out.Rows[0].Selection[0] != 1 || math.IsNaN(out.Rows[0].Centers[0].Y) {
		t.Fatalf("single row failed: %+v", out)
	}
}
