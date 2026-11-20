package academic

import (
	"errors"
	"image"
	"math"
	"sort"
)

// GridBlock records user-calibrated bubble centers as fractions of image width/height.
type GridBlock struct {
	First       int   `json:"first"`
	Count       int   `json:"count"`
	TopLeft     Point `json:"top_left"`
	TopRight    Point `json:"top_right"`
	BottomLeft  Point `json:"bottom_left"`
	BottomRight Point `json:"bottom_right"`
}
type GridCalibration struct {
	Options int         `json:"options"`
	Radius  float64     `json:"radius"`
	Blocks  []GridBlock `json:"blocks"`
	Area    *GridRect   `json:"area,omitempty"`
	Frames  []GridFrame `json:"frames,omitempty"`
}
type GridRect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}
type GridFrame struct {
	Number int      `json:"number"`
	Rect   GridRect `json:"rect"`
}
type GridRow struct {
	ModelSelection []int     `json:"model_selection,omitempty"`
	ModelState     string    `json:"model_state,omitempty"`
	Number         int       `json:"number"`
	Selection      []int     `json:"selection"`
	Fill           []float64 `json:"fill"`
	Centers        []Point   `json:"centers"`
	Uncertain      bool      `json:"uncertain"`
	Reason         string    `json:"reason"`
}
type GridReading struct {
	Rows   []GridRow `json:"rows"`
	Notice string    `json:"notice"`
}

// ReadGrid measures pixels without a key. It never assigns a learner or awards marks.
func ReadGrid(img image.Image, c GridCalibration) (GridReading, error) {
	out := GridReading{Rows: []GridRow{}, Notice: "Calibrated preview only. Confirm every response and learner identity before using it as academic evidence."}
	b := img.Bounds()
	if b.Dx() < 100 || b.Dy() < 100 || int64(b.Dx())*int64(b.Dy()) > 32_000_000 {
		return out, errors.New("use an image between 100 pixels and 32 megapixels")
	}
	if c.Options < 2 || c.Options > 8 || len(c.Blocks) < 1 || len(c.Blocks) > 8 || math.IsNaN(c.Radius) || math.IsInf(c.Radius, 0) || c.Radius < .001 || c.Radius > .03 {
		return out, errors.New("calibrate 1–8 blocks, 2–8 options and a radius from 0.001 to 0.03 of image width")
	}
	seen := map[int]bool{}
	frames := map[int]GridRect{}
	validRect := func(r GridRect) bool {
		for _, n := range []float64{r.X, r.Y, r.Width, r.Height} {
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return false
			}
		}
		return r.X >= 0 && r.Y >= 0 && r.Width > 0 && r.Height > 0 && r.X+r.Width <= 1+1e-8 && r.Y+r.Height <= 1+1e-8
	}
	inside := func(p Point, r GridRect) bool {
		return p.X >= r.X-1e-8 && p.X <= r.X+r.Width+1e-8 && p.Y >= r.Y-1e-8 && p.Y <= r.Y+r.Height+1e-8
	}
	if c.Area != nil && !validRect(*c.Area) || len(c.Frames) > 500 {
		return out, errors.New("draw an answer area inside the image and at most 500 question frames")
	}
	for _, frame := range c.Frames {
		r := frame.Rect
		_, duplicate := frames[frame.Number]
		if frame.Number < 1 || duplicate || !validRect(r) || r.Width <= 2*c.Radius || r.Height+1e-8 < 2*c.Radius*float64(b.Dx())/float64(b.Dy()) {
			return out, errors.New("question frames must be unique, inside the image and large enough for the sampling circles")
		}
		if c.Area != nil && (!inside(Point{r.X, r.Y}, *c.Area) || !inside(Point{r.X + r.Width, r.Y + r.Height}, *c.Area)) {
			return out, errors.New("question frames must be inside the selected answer area")
		}
		frames[frame.Number] = r
	}
	total := 0
	for _, block := range c.Blocks {
		total += block.Count
		if block.First < 1 || block.Count < 1 || block.Count > 200 || total > 500 {
			return out, errors.New("calibration must contain 1–500 numbered responses")
		}
		for _, p := range []Point{block.TopLeft, block.TopRight, block.BottomLeft, block.BottomRight} {
			if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) || p.X < 0 || p.Y < 0 || p.X > 1 || p.Y > 1 {
				return out, errors.New("calibration centers must be inside the image")
			}
		}
		if block.TopRight.X-block.TopLeft.X < .01 || block.BottomRight.X-block.BottomLeft.X < .01 || block.Count > 1 && (block.BottomLeft.Y-block.TopLeft.Y < .01 || block.BottomRight.Y-block.TopRight.Y < .01) {
			return out, errors.New("mark the four bubble-center corners in reading order")
		}
		for n := 0; n < block.Count; n++ {
			number := block.First + n
			if seen[number] {
				return out, errors.New("calibration contains overlapping question numbers")
			}
			seen[number] = true
			row := GridRow{Number: number, Selection: []int{}, Fill: []float64{}, Centers: []Point{}}
			v := float64(n) / float64(max(1, block.Count-1))
			for opt := 0; opt < c.Options; opt++ {
				u := float64(opt) / float64(c.Options-1)
				p := Point{(1-v)*((1-u)*block.TopLeft.X+u*block.TopRight.X) + v*((1-u)*block.BottomLeft.X+u*block.BottomRight.X), (1-v)*((1-u)*block.TopLeft.Y+u*block.TopRight.Y) + v*((1-u)*block.BottomLeft.Y+u*block.BottomRight.Y)}
				if frame, ok := frames[number]; ok {
					p = Point{frame.X + c.Radius + u*(frame.Width-2*c.Radius), frame.Y + frame.Height/2}
				}
				ry := c.Radius * float64(b.Dx()) / float64(b.Dy())
				if c.Area != nil && (!inside(Point{p.X - c.Radius, p.Y - ry}, *c.Area) || !inside(Point{p.X + c.Radius, p.Y + ry}, *c.Area)) {
					return out, errors.New("a sampling circle leaves the answer area; adjust the grid or redraw its question frame")
				}
				row.Centers = append(row.Centers, p)
				r := max(1, int(c.Radius*float64(b.Dx())))
				black, pixels := 0, 0
				for y := -r; y <= r; y++ {
					for x := -r; x <= r; x++ {
						if x*x+y*y > r*r {
							continue
						}
						px := int(p.X*float64(b.Dx())) + x + b.Min.X
						py := int(p.Y*float64(b.Dy())) + y + b.Min.Y
						if !image.Pt(px, py).In(b) {
							return out, errors.New("a sampling circle leaves the image; correct the calibration")
						}
						pixels++
						red, green, blue, _ := img.At(px, py).RGBA()
						if red > 35000 && red > green+6000 && red > blue+6000 {
							continue
						}
						if (299*red+587*green+114*blue)/1000 < 26000 {
							black++
						}
					}
				}
				ratio := float64(black) / float64(pixels)
				row.Fill = append(row.Fill, ratio)
				if ratio >= .65 {
					row.Selection = append(row.Selection, opt+1)
				} else if ratio > .25 {
					row.Uncertain = true
					row.Reason = "Partial fill or calibration overlap; inspect the source."
				}
			}
			if len(row.Selection) != 1 {
				row.Uncertain = true
				if len(row.Selection) == 0 {
					row.Reason = "No clear filled bubble. Confirm blank or correct the reading."
				} else {
					row.Reason = "Multiple filled bubbles. Confirm the intended response."
				}
			}
			out.Rows = append(out.Rows, row)
		}
	}
	for number := range frames {
		if !seen[number] {
			return out, errors.New("a question frame does not belong to a configured response block")
		}
	}
	sort.Slice(out.Rows, func(i, j int) bool { return out.Rows[i].Number < out.Rows[j].Number })
	return out, nil
}
