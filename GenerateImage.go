package deejdsp

import (
	"crypto/sha1"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/jax-b/deej"
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
	imagebounds := constructedImage.Bounds().Max
	for x := 0; x < imagebounds.X; x++ {
		for y := 0; y < imagebounds.Y; y++ {
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

// CreateFileName creates a truncated 8 character filename using sha1 and the .b ending
func CreateFileName(processname string) string {
	h := sha1.New()
	h.Write([]byte(processname))
	namehashed := h.Sum(nil)
	name := fmt.Sprintf("%x", namehashed)
	return name[0:7] + ".B"
}

// CreateAutoMap creates a automatic mapping of sessions to the displays
func CreateAutoMap(SliderMap *deej.SliderMap, SessionMap *deej.SessionMap) map[int]string {
	AutoMap := make(map[int]string)
	SliderMap.Iterate(func(index int, targets []string) {
		for _, mapping := range targets {
			session, _ := SessionMap.Get(strings.ToLower(mapping))
			if session != nil {
				AutoMap[index] = mapping
				break
			}
		}
	})
	return AutoMap
}
