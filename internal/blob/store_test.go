package blob

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestAvatarBoundedDecode(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		w, h  int
		valid bool
	}{{"square", 64, 64, true}, {"uncropped", 64, 32, false}, {"too_small", 16, 16, false}, {"excessive_dimensions", 2048, 2, false}} {
		t.Run(test.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, test.w, test.h))
			img.Set(0, 0, color.RGBA{A: 255})
			var b bytes.Buffer
			if err := png.Encode(&b, img); err != nil {
				t.Fatal(err)
			}
			out, err := NormalizeAvatar(b.Bytes())
			if test.valid && (err != nil || len(out) == 0) {
				t.Fatal(err)
			}
			if !test.valid && err == nil {
				t.Fatal("invalid image accepted")
			}
		})
	}
	for _, data := range [][]byte{[]byte("<svg onload='alert(1)'></svg>"), make([]byte, MaxAvatarBytes+1), []byte("not an image")} {
		if _, err := NormalizeAvatar(data); err == nil {
			t.Fatal("untrusted input accepted")
		}
	}
}
