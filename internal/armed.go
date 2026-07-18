// Multi-stripe armed-running sheets shared by the gopher and knight pipelines.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
)

const armedDirectionsPerStripe = 16

type armedFrameFactory func(direction string, aim, frame int) *image.NRGBA

type armedShadowFactory func(direction string, aim int) *image.NRGBA

type armedFrame struct {
	body   *image.NRGBA
	shadow *image.NRGBA
}

type armedAnimationFactory func(direction string, aim, frame int) armedFrame

func writeArmedRunning(
	gfxDir string,
	name string,
	size int,
	frameFactory armedFrameFactory,
	shadowFactory armedShadowFactory,
) error {
	shadows := make([]*image.NRGBA, len(gunMapping))
	if err := writeArmedAnimation(
		gfxDir,
		name+"-running-with-gun",
		size,
		frames,
		func(direction string, aim, frame int) armedFrame {
			body := frameFactory(direction, aim, frame)
			if shadows[aim] == nil {
				shadows[aim] = makeFittedShadow(shadowFactory(direction, aim))
			}
			return armedFrame{body: body, shadow: shadows[aim]}
		},
	); err != nil {
		return err
	}
	return writeFlippedArmedShadows(gfxDir, name, size, shadowFactory)
}

func writeFlippedArmedShadows(
	gfxDir, name string,
	size int,
	shadowFactory armedShadowFactory,
) error {
	for start := 0; start < len(gunMapping); start += armedDirectionsPerStripe {
		end := min(start+armedDirectionsPerStripe, len(gunMapping))
		rows := end - start
		sheet := newCanvas(size*frames, size*rows)
		for row, direction := range gunMapping[start:end] {
			aim := start + row
			shadow := makeFittedShadow(flipHorizontal(shadowFactory(direction, aim)))
			for frame := range frames {
				pasteAt(sheet, shadow, frame*size, row*size)
			}
		}

		stripe := start/armedDirectionsPerStripe + 1
		if err := saveArmedStripe(
			gfxDir,
			name+"-running-with-gun-flipped",
			stripe,
			sheet,
			true,
		); err != nil {
			return err
		}
	}
	return nil
}

func writeArmedAnimation(
	gfxDir, name string,
	size, frameCount int,
	frameFactory armedAnimationFactory,
) error {
	for start := 0; start < len(gunMapping); start += armedDirectionsPerStripe {
		end := min(start+armedDirectionsPerStripe, len(gunMapping))
		rows := end - start
		bodySheet := newCanvas(size*frameCount, size*rows)
		shadowSheet := newCanvas(size*frameCount, size*rows)
		for row, direction := range gunMapping[start:end] {
			aim := start + row
			for frame := range frameCount {
				images := frameFactory(direction, aim, frame)
				pasteAt(bodySheet, images.body, frame*size, row*size)
				pasteAt(shadowSheet, images.shadow, frame*size, row*size)
			}
		}

		stripe := start/armedDirectionsPerStripe + 1
		if err := saveArmedStripe(gfxDir, name, stripe, bodySheet, false); err != nil {
			return err
		}
		if err := saveArmedStripe(gfxDir, name, stripe, shadowSheet, true); err != nil {
			return err
		}
	}
	return nil
}

func saveArmedStripe(
	gfxDir string,
	name string,
	stripe int,
	sheet *image.NRGBA,
	shadow bool,
) error {
	suffix := ""
	kind := "sheet"
	if shadow {
		suffix = "-shadow"
		kind = "shadow sheet"
	}
	out := filepath.Join(gfxDir, fmt.Sprintf("%s-%d%s.png", name, stripe, suffix))
	if err := savePNG(out, sheet); err != nil {
		return fmt.Errorf("save %s %s %d: %w", name, kind, stripe, err)
	}
	slog.Info("wrote sheet", "path", out, "width", sheet.Bounds().Dx(), "height", sheet.Bounds().Dy())
	return nil
}
