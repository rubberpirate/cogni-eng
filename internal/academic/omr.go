package academic

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type BubbleRow struct {
	QuestionID string   `json:"question_id"`
	Type       string   `json:"type"`
	Labels     []string `json:"labels"`
	Centers    []Point  `json:"centers"`
}
type Layout struct {
	Version string      `json:"version"`
	Columns int         `json:"columns"`
	Rows    []BubbleRow `json:"rows"`
}
type ReadRow struct {
	Answer    Answer    `json:"answer"`
	Fill      []float64 `json:"fill"`
	Uncertain bool      `json:"uncertain"`
	Reason    string    `json:"reason"`
}
type ScanRead struct {
	AssignmentID string             `json:"assignment_id"`
	Rows         map[string]ReadRow `json:"rows"`
	Corners      []Point            `json:"corners"`
	Error        string             `json:"error"`
}

func NewLayout(p Paper, columns int) (Layout, error) {
	if columns != 1 && columns != 2 {
		return Layout{}, errors.New("choose one or two columns")
	}
	l := Layout{Version: "registered-bubbles-v1", Columns: columns, Rows: []BubbleRow{}}
	count := 0
	for _, s := range p.Sections {
		count += len(s.Questions)
	}
	if count > 40*columns {
		return l, fmt.Errorf("this layout supports at most %d questions", 40*columns)
	}
	n := 0
	for _, s := range p.Sections {
		for _, q := range s.Questions {
			col, row := n/40, n%40
			x := 110.0 + float64(col)*440
			r := BubbleRow{QuestionID: q.ID, Type: q.Type, Labels: []string{}, Centers: []Point{}}
			if q.Type == "single" || q.Type == "multiple" {
				for i, o := range q.Options {
					r.Labels = append(r.Labels, o.ID)
					r.Centers = append(r.Centers, Point{x + 70 + float64(i)*34, 300 + float64(row)*25})
				}
			}
			l.Rows = append(l.Rows, r)
			n++
		}
	}
	return l, nil
}
func label(img *image.RGBA, x, y int, text string) {
	d := font.Drawer{Dst: img, Src: image.NewUniform(color.Black), Face: basicfont.Face7x13, Dot: fixed.P(x, y)}
	d.DrawString(text)
}
func circle(img *image.RGBA, p Point, r int, filled bool) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			d := x*x + y*y
			if d <= r*r && (filled || d >= (r-1)*(r-1)) {
				img.Set(int(p.X)+x, int(p.Y)+y, color.Black)
			}
		}
	}
}
func SheetImage(l Layout, assignment, fingerprint, title, learner string, filled map[string]Answer) ([]byte, error) {
	code, err := hex.DecodeString(assignment + fingerprint[:16])
	if err != nil || len(code) != 40 {
		return nil, errors.New("invalid sheet identity")
	}
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1400))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	for _, p := range []Point{{40, 40}, {960, 40}, {960, 1360}, {40, 1360}} {
		draw.Draw(img, image.Rect(int(p.X)-10, int(p.Y)-10, int(p.X)+11, int(p.Y)+11), image.NewUniform(color.Black), image.Point{}, draw.Src)
	}
	for i := 0; i < 320; i++ {
		if code[i/8]&(1<<uint(7-i%8)) != 0 {
			x, y := 60+i%32*8, 80+i/32*8
			draw.Draw(img, image.Rect(x, y, x+8, y+8), image.NewUniform(color.Black), image.Point{}, draw.Src)
		}
	}
	label(img, 350, 90, "REGISTERED ANSWER SHEET / v1")
	label(img, 60, 200, title)
	label(img, 60, 225, learner)
	label(img, 60, 250, "Fill circles completely. Keep corner squares and identity grid visible.")
	for i, r := range l.Rows {
		x, y := 110, 300+(i%40)*25
		if i >= 40 {
			x += 440
		}
		label(img, x, y+4, fmt.Sprint(i+1))
		if len(r.Labels) == 0 {
			label(img, x+50, y+4, "Written / numerical: manual review")
			continue
		}
		for j, p := range r.Centers {
			circle(img, p, 9, false)
			label(img, int(p.X)-3, int(p.Y)-13, r.Labels[j])
			for _, selected := range filled[r.QuestionID].Selection {
				if selected == r.Labels[j] {
					circle(img, p, 7, true)
				}
			}
		}
	}
	var out bytes.Buffer
	err = png.Encode(&out, img)
	return out.Bytes(), err
}
func dark(img image.Image, x, y int) bool {
	if !image.Pt(x, y).In(img.Bounds()) {
		return false
	}
	r, g, b, _ := img.At(x, y).RGBA()
	return (299*r+587*g+114*b)/1000 < 26000
}

// Registration uses connected filled corner components; bubble sampling excludes printed outlines.
func corners(img image.Image) ([]Point, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := []Point{}
	for _, quadrant := range [][2]int{{0, 0}, {1, 0}, {1, 1}, {0, 1}} {
		x0, y0 := quadrant[0]*w*3/4, quadrant[1]*h*3/4
		ww, hh := w/4, h/4
		seen := make([]bool, ww*hh)
		best := 0
		bestDistance := math.MaxFloat64
		var center Point
		for y := 0; y < hh; y++ {
			for x := 0; x < ww; x++ {
				index := y*ww + x
				if seen[index] || !dark(img, x+x0, y+y0) {
					continue
				}
				queue := []int{index}
				seen[index] = true
				minx, maxx, miny, maxy := x, x, y, y
				sumx, sumy := 0, 0
				for k := 0; k < len(queue); k++ {
					v := queue[k]
					xx, yy := v%ww, v/ww
					sumx += xx
					sumy += yy
					minx = min(minx, xx)
					maxx = max(maxx, xx)
					miny = min(miny, yy)
					maxy = max(maxy, yy)
					for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
						nx, ny := xx+d[0], yy+d[1]
						if nx >= 0 && nx < ww && ny >= 0 && ny < hh {
							ni := ny*ww + nx
							if !seen[ni] && dark(img, nx+x0, ny+y0) {
								seen[ni] = true
								queue = append(queue, ni)
							}
						}
					}
				}
				bw, bh := maxx-minx+1, maxy-miny+1
				area := bw * bh
				if bw > 5 && bh > 5 && float64(bw)/float64(bh) > .55 && float64(bw)/float64(bh) < 1.8 && float64(len(queue))/float64(area) > .75 && float64(area) < float64(w*h)*.003 {
					candidate := Point{float64(x0) + float64(sumx)/float64(len(queue)), float64(y0) + float64(sumy)/float64(len(queue))}
					dx, dy := candidate.X/float64(w), candidate.Y/float64(h)
					if quadrant[0] == 1 {
						dx = 1 - dx
					}
					if quadrant[1] == 1 {
						dy = 1 - dy
					}
					distance := dx*dx + dy*dy
					// The identity grid can contain larger filled squares; registration is the eligible component nearest the page corner.
					if distance < bestDistance {
						bestDistance = distance
						best = len(queue)
						center = candidate
					}
				}
			}
		}
		if best == 0 {
			return nil, errors.New("four registration squares could not be located; rescan the full page")
		}
		out = append(out, center)
	}
	return out, nil
}
func projection(dst []Point) (func(Point) Point, error) {
	return projectionBetween([]Point{{40, 40}, {960, 40}, {960, 1360}, {40, 1360}}, dst)
}
func projectionBetween(src, dst []Point) (func(Point) Point, error) {
	var a [8][9]float64
	for i, p := range src {
		q := dst[i]
		a[i*2] = [9]float64{p.X, p.Y, 1, 0, 0, 0, -q.X * p.X, -q.X * p.Y, q.X}
		a[i*2+1] = [9]float64{0, 0, 0, p.X, p.Y, 1, -q.Y * p.X, -q.Y * p.Y, q.Y}
	}
	for k := 0; k < 8; k++ {
		pivot := k
		for j := k + 1; j < 8; j++ {
			if math.Abs(a[j][k]) > math.Abs(a[pivot][k]) {
				pivot = j
			}
		}
		a[k], a[pivot] = a[pivot], a[k]
		if math.Abs(a[k][k]) < 1e-9 {
			return nil, errors.New("invalid registration geometry")
		}
		v := a[k][k]
		for j := k; j < 9; j++ {
			a[k][j] /= v
		}
		for i := 0; i < 8; i++ {
			if i != k {
				v = a[i][k]
				for j := k; j < 9; j++ {
					a[i][j] -= v * a[k][j]
				}
			}
		}
	}
	return func(p Point) Point {
		d := a[6][8]*p.X + a[7][8]*p.Y + 1
		return Point{(a[0][8]*p.X + a[1][8]*p.Y + a[2][8]) / d, (a[3][8]*p.X + a[4][8]*p.Y + a[5][8]) / d}
	}, nil
}
func Recognize(img image.Image, l Layout, fingerprint string) ScanRead {
	out := ScanRead{Rows: map[string]ReadRow{}}
	if img.Bounds().Dx() < 300 || img.Bounds().Dy() < 400 || int64(img.Bounds().Dx())*int64(img.Bounds().Dy()) > 32_000_000 {
		out.Error = "Page dimensions are outside the supported 32 megapixel bound."
		return out
	}
	cs, err := corners(img)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	var transform func(Point) Point
	for rotation := 0; rotation < 4; rotation++ {
		ordered := append(append([]Point{}, cs[rotation:]...), cs[:rotation]...)
		fn, e := projection(ordered)
		if e != nil {
			continue
		}
		code := make([]byte, 40)
		for i := 0; i < 320; i++ {
			p := fn(Point{64 + float64(i%32)*8, 84 + float64(i/32)*8})
			if dark(img, int(math.Round(p.X)), int(math.Round(p.Y))) {
				code[i/8] |= 1 << uint(7-i%8)
			}
		}
		if hex.EncodeToString(code[32:]) == fingerprint[:16] {
			out.AssignmentID = hex.EncodeToString(code[:32])
			out.Corners = ordered
			transform = fn
			break
		}
	}
	if transform == nil {
		out.Error = "Sheet identity or layout fingerprint is unreadable. Confirm the assigned learner and rescan using this layout."
		return out
	}
	for _, r := range l.Rows {
		reading := ReadRow{Fill: []float64{}, Answer: Answer{Selection: []string{}, Visited: true}}
		if len(r.Centers) == 0 {
			reading.Uncertain = true
			reading.Reason = "Written and numerical responses require manual transcription/review."
		}
		for i, c := range r.Centers {
			total, black := 0, 0
			for dy := -5; dy <= 5; dy++ {
				for dx := -5; dx <= 5; dx++ {
					if dx*dx+dy*dy <= 25 {
						p := transform(Point{c.X + float64(dx), c.Y + float64(dy)})
						total++
						if dark(img, int(math.Round(p.X)), int(math.Round(p.Y))) {
							black++
						}
					}
				}
			}
			ratio := float64(black) / float64(total)
			reading.Fill = append(reading.Fill, ratio)
			if ratio >= .65 {
				reading.Answer.Selection = append(reading.Answer.Selection, r.Labels[i])
			} else if ratio > .12 {
				reading.Uncertain = true
				reading.Reason = "Partial or faint fill needs reviewer confirmation."
			}
		}
		if r.Type == "single" && len(reading.Answer.Selection) > 1 {
			reading.Uncertain = true
			reading.Reason = "Multiple circles in a single-selection question."
		}
		out.Rows[r.QuestionID] = reading
	}
	return out
}
