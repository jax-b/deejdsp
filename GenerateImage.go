package deejdsp

import (
	"image"
	"image/color"

	"github.com/jax-b/iconextract"
	"github.com/jax-b/ssd1306FilePrep"
	"github.com/nfnt/resize"
)

const (
	displaySizeX = 128
	displaySizeY = 64
)

// GetAndConvertIMG returns a byteslice with the converted image
func GetAndConvertIMG(filepath string, index int32, threshold int) [][]byte {
	extractedimage, _ := iconextract.ExtractIcon(filepath, index)
	resizedImage := resize.Thumbnail(uint(displaySizeX), uint(displaySizeY), extractedimage, resize.NearestNeighbor)
	amountToCenter := (displaySizeX - resizedImage.Bounds().Max.X) / 2

	constructedImage := image.NewRGBA(image.Rect(0, 0, displaySizeX, displaySizeY))

	for x := 0; x < constructedImage.Bounds().Max.X; x++ {
		for y := 0; y < constructedImage.Bounds().Max.Y; y++ {
			constructedImage.Set(x, y, color.Black)
		}
	}

	for x := 0; x < resizedImage.Bounds().Max.X; x++ {
		for y := 0; y < resizedImage.Bounds().Max.Y; y++ {
			constructedImage.Set(x+amountToCenter, y, resizedImage.At(x, y))
		}
	}
	bwimage := ssd1306FilePrep.ConvertBW(constructedImage, uint8(threshold))
	bytedIMG := ssd1306FilePrep.ToBWByteSlice(bwimage, uint8(threshold))
	return bytedIMG
}
