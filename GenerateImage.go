package deejdsp

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/jax-b/deej/pkg/deej"
	"github.com/jax-b/iconfinderapi"
	"github.com/jax-b/ssd1306FilePrep"
	"github.com/nfnt/resize"
)

const (
	displaySizeX = 128
	displaySizeY = 64
)

// GetIconFromAPI gets an icon from online
// This calls the API from icon finder and tryes to get the first icon that matches the requirements
// It only looks up 3 icons
// It filters on flat icons, it cannot be a icon that needs to be bought
// It cannot be a vector image
func GetIconFromAPI(icofdr *iconfinderapi.Iconfinder, keyword string) (image.Image, error) {
	search, err := icofdr.SearchIcons(keyword, 3, -1, 0, 0, "", "", "flat")
	if err != nil {
		return nil, err
	}
	for _, results := range search.Icons {
		for _, size := range results.Rasters {
			// Looks for the first icon that is equal to or bigger than the display size
			// We resize it anyways so we are just looking for an icon with the most detail
			if size.SizeHeight >= displaySizeY {
				for _, format := range size.Formats {
					if format.Format == "png" {
						return icofdr.DownloadIcon(format), nil
					}
				}
				// if we cannot find a png (preferential) then we download the jpg
				return icofdr.DownloadIcon(size.Formats[0]), nil
			}
		}
	}

	return nil, errors.New("Unable to find a Compatable image")
}

// ConvertImage returns a byteslice with the converted image
// Basicly a copy of the main test program in my ssd1306 file prep lib but we dont write it to a file
func ConvertImage(srcimg image.Image, index int32, threshold int) ([][]byte, error) {
	if srcimg == nil {
		return nil, errors.New("srcimg equal to nil")
	}
	resizedImage := resize.Thumbnail(uint(displaySizeX), uint(displaySizeY), srcimg, resize.NearestNeighbor)

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
	return bytedIMG, nil
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
