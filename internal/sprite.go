// Low-level NRGBA PNG helpers used by the procedural sheet generators in
// this package: load/save plus pixel-buffer transforms (clone, paste,
// overlay, line drawing, shift, resize, alpha scale, blacken, shear).

package gopher

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

// loadPNG reads a PNG file and returns it as *image.NRGBA. Non-NRGBA inputs
// are converted.
func loadPNG(path string) (img *image.NRGBA, err error) {
	file, err := openRootedFile(path)
	if err != nil {
		return nil, err
	}
	defer closeWithErr(file, &err)

	src, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	if nrgba, ok := src.(*image.NRGBA); ok {
		return nrgba, nil
	}

	out := image.NewNRGBA(src.Bounds())
	draw.Copy(out, image.Point{}, src, src.Bounds(), draw.Src, nil)
	return out, nil
}

// savePNG writes an image as PNG to path.
func savePNG(path string, img image.Image) (err error) {
	cleanPath, err := writablePath(path)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Clean(cleanPath))
	if err != nil {
		return err
	}
	defer closeWithErr(f, &err)
	return png.Encode(f, img)
}

// newCanvas returns a transparent NRGBA canvas of the given size.
func newCanvas(w, h int) *image.NRGBA {
	return image.NewNRGBA(image.Rect(0, 0, w, h))
}

// clone returns an independent copy of src.
func clone(src *image.NRGBA) *image.NRGBA {
	out := image.NewNRGBA(src.Bounds())
	copy(out.Pix, src.Pix)
	return out
}

func flipHorizontal(src *image.NRGBA) *image.NRGBA {
	bounds := src.Bounds()
	out := newCanvas(bounds.Dx(), bounds.Dy())
	for y := range bounds.Dy() {
		for x := range bounds.Dx() {
			out.SetNRGBA(bounds.Dx()-1-x, y, src.NRGBAAt(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return out
}

// pasteAt draws src into dst at (dx, dy), replacing existing pixels.
func pasteAt(dst *image.NRGBA, src image.Image, dx, dy int) {
	b := src.Bounds()
	r := image.Rect(dx, dy, dx+b.Dx(), dy+b.Dy())
	draw.Copy(dst, r.Min, src, b, draw.Src, nil)
}

// overlay copies non-transparent pixels of src onto dst (source-over with no
// alpha blending; sprites use solid alpha so this is fine).
func overlay(dst, src *image.NRGBA) {
	for i := 0; i < len(dst.Pix); i += 4 {
		if src.Pix[i+3] == 0 {
			continue
		}
		dst.Pix[i] = src.Pix[i]
		dst.Pix[i+1] = src.Pix[i+1]
		dst.Pix[i+2] = src.Pix[i+2]
		dst.Pix[i+3] = src.Pix[i+3]
	}
}

func drawThickLine(
	img *image.NRGBA,
	x0, y0, x1, y1, radius int,
	colour color.NRGBA,
) {
	dx := x1 - x0
	dy := y1 - y0
	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		drawSquare(img, x0, y0, radius, colour)
		return
	}
	for step := 0; step <= steps; step++ {
		x := x0 + dx*step/steps
		y := y0 + dy*step/steps
		drawSquare(img, x, y, radius, colour)
	}
}

func drawSquare(img *image.NRGBA, cx, cy, radius int, colour color.NRGBA) {
	bounds := img.Bounds()
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if image.Pt(x, y).In(bounds) {
				img.SetNRGBA(x, y, colour)
			}
		}
	}
}

// shiftUp returns a new image with content shifted up by amount pixels;
// vacated bottom rows are transparent.
func shiftUp(src *image.NRGBA, amount int) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))

	if amount <= 0 {
		copy(out.Pix, src.Pix)
		return out
	}
	if amount >= h {
		return out
	}

	copy(out.Pix, src.Pix[amount*src.Stride:])
	return out
}

// resize returns a new image scaled to (w, h) using Catmull-Rom interpolation.
func resize(src *image.NRGBA, w, h int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(out, out.Bounds(), src, src.Bounds(), draw.Over, nil)
	return out
}

// scaleAlpha multiplies every pixel's alpha by factor/255.
func scaleAlpha(img *image.NRGBA, factor uint8) {
	for i := 3; i < len(img.Pix); i += 4 {
		scaled := uint16(img.Pix[i]) * uint16(factor) / 255
		img.Pix[i] = clampByte(scaled)
	}
}

func clampByte(value uint16) uint8 {
	if value > math.MaxUint8 {
		return math.MaxUint8
	}
	return uint8(value)
}

// blacken sets every pixel's RGB to 0 (preserves alpha).
func blacken(img *image.NRGBA) {
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 0
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
	}
}

// shearRight shears the image so each row's content shifts right by
// (h - y) * tan pixels. Anchored at the bottom row.
func shearRight(src *image.NRGBA, tan float64) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))

	for y := range h {
		offset := int(float64(h-y) * tan)
		for x := range w {
			sx := x - offset
			if sx < 0 || sx >= w {
				continue
			}
			i := y*src.Stride + sx*4
			j := y*out.Stride + x*4
			out.Pix[j] = src.Pix[i]
			out.Pix[j+1] = src.Pix[i+1]
			out.Pix[j+2] = src.Pix[i+2]
			out.Pix[j+3] = src.Pix[i+3]
		}
	}
	return out
}
