package valetudo_map_renderer

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/SkaceKamen/valetudo-telegram-bot/assets"
	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/valetudo"
	"github.com/fogleman/gg"
)

var vacuumImage *image.Image
var chargerImage *image.Image

func getLayerOrder(layer valetudo.RobotStateMapLayer) int {
	if layer.Type == "wall" {
		return 3
	}
	if layer.Type == "floor" {
		return 2
	}
	if layer.Type == "segment" {
		return 1
	}

	return 0
}

func getEntityOrder(entity valetudo.RobotStateMapEntity) int {
	if entity.Type == "path" {
		return 0
	}
	if entity.Type == "predicted_path" {
		return 1
	}
	if entity.Type == "charger_location" {
		return 2
	}
	if entity.Type == "robot_position" {
		return 3
	}

	return 4
}

func RenderMap(mapData *valetudo.RobotStateMap) []byte {
	if vacuumImage == nil {
		img, _, err := image.Decode(bytes.NewReader(assets.VacuumImage))
		if err != nil {
			panic(err)
		}
		vacuumImage = &img
	}

	if chargerImage == nil {
		img, _, err := image.Decode(bytes.NewReader(assets.ChargerImage))
		if err != nil {
			panic(err)
		}
		chargerImage = &img
	}

	scale := 2.0

	w := int(math.Round(float64(mapData.Size.X) / float64(mapData.PixelSize)))
	h := int(math.Round(float64(mapData.Size.Y) / float64(mapData.PixelSize)))

	minX := w
	minY := h
	maxX := 0
	maxY := 0

	for _, layer := range mapData.Layers {
		if layer.Dimensions.X.Min < minX {
			minX = layer.Dimensions.X.Min
		}

		if layer.Dimensions.X.Max > maxX {
			maxX = layer.Dimensions.X.Max
		}

		if layer.Dimensions.Y.Min < minY {
			minY = layer.Dimensions.Y.Min
		}

		if layer.Dimensions.Y.Max > maxY {
			maxY = layer.Dimensions.Y.Max
		}
	}

	if minX == w {
		minX = 0
	}

	if minY == h {
		minY = 0
	}

	if maxX == 0 {
		maxX = w
	}

	if maxY == 0 {
		maxY = h
	}

	minX -= int(float64(w) * 0.01)
	minY -= int(float64(h) * 0.01)
	maxX += int(float64(h) * 0.01)
	maxY += int(float64(h) * 0.01)

	resizedW := int(math.Round(float64(maxX-minX) * scale))
	resizedH := int(math.Round(float64(maxY-minY) * scale))

	ctx := gg.NewContext(resizedW, resizedH)

	sort.Slice(mapData.Layers, func(i, j int) bool {
		orderA := getLayerOrder(mapData.Layers[i])
		orderB := getLayerOrder(mapData.Layers[j])

		return orderA < orderB
	})

	for _, layer := range mapData.Layers {
		// Only continue with supported layers
		if layer.Type != "wall" && layer.Type != "floor" && layer.Type != "segment" {
			continue
		}

		layerColor := color.RGBA{0, 0, 0, 255}
		if layer.Type == "wall" {
			layerColor = color.RGBA{0, 0, 0, 255}
		}
		if layer.Type == "floor" {
			layerColor = color.RGBA{200, 200, 200, 255}
		}
		if layer.Type == "segment" {
			layerColor = color.RGBA{128, 128, 128, 255}
		}

		renderLayer(ctx, layer, minX, minY, layerColor, scale)
	}

	sort.Slice(mapData.Entities, func(i, j int) bool {
		orderA := getEntityOrder(mapData.Entities[i])
		orderB := getEntityOrder(mapData.Entities[j])

		return orderA < orderB
	})

	for _, entity := range mapData.Entities {
		x := ((float64((*entity.Points)[0]) / float64(mapData.PixelSize)) - float64(minX)) * scale
		y := ((float64((*entity.Points)[1]) / float64(mapData.PixelSize)) - float64(minY)) * scale

		if entity.Type == "charger_location" {
			ctx.DrawImageAnchored(*chargerImage, int(x), int(y), 0.5, 0.5)
		}

		if entity.Type == "robot_position" {
			ctx.Push()

			if entity.Metadata.Angle != nil {
				ctx.RotateAbout(gg.Degrees(float64(*entity.Metadata.Angle)), x, y)
			}

			ctx.DrawImageAnchored(*vacuumImage, int(x), int(y), 0.5, 0.5)
			ctx.Pop()
		}

		if entity.Type == "path" || entity.Type == "predicted_path" {
			if entity.Points != nil && len(*entity.Points) > 0 {
				if entity.Type == "path" {
					ctx.SetColor(color.RGBA{100, 100, 100, 255})
				} else {
					ctx.SetColor(color.RGBA{100, 100, 100, 100})
				}
				ctx.SetLineWidth(2)
				ctx.MoveTo(x, y)

				for i := 2; i < len(*entity.Points); i += 2 {
					ctx.LineTo(
						((float64((*entity.Points)[i])/float64(mapData.PixelSize))-float64(minX))*scale,
						((float64((*entity.Points)[i+1])/float64(mapData.PixelSize))-float64(minY))*scale,
					)
				}

				ctx.Stroke()
			}
		}
	}

	ctx.Scale(3, 3)
	ctx.Stroke()

	buffer := bytes.Buffer{}
	ctx.EncodePNG(&buffer)

	return buffer.Bytes()
}

func renderLayer(ctx *gg.Context, layer valetudo.RobotStateMapLayer, xOffset int, yOffset int, color color.Color, scale float64) {
	ctx.SetColor(color)

	for i := 0; i < len(layer.Pixels); i += 2 {
		x := layer.Pixels[i]
		y := layer.Pixels[i+1]
		ctx.DrawRectangle(float64(x-xOffset)*scale, float64(y-yOffset)*scale, scale, scale)
		ctx.Fill()
	}

	for i := 0; i < len(layer.CompressedPixels); i += 3 {
		xStart := layer.CompressedPixels[i]
		y := layer.CompressedPixels[i+1]
		count := layer.CompressedPixels[i+2]
		for j := 0; j < count; j++ {
			x := xStart + j
			// ctx.SetPixel(int(float64(x-xOffset)*scale), int(float64(y-yOffset)*scale))
			ctx.DrawRectangle(float64(x-xOffset)*scale, float64(y-yOffset)*scale, scale, scale)
			ctx.Fill()
		}
	}
}
