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
