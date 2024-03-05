package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/matheustavarestrindade/niimprintgo/internal/app/logger"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/niimbot"
)

func main() {
    fmt.Println("Starting...")
	printer := niimbot.NewNiimbotPrinter("COM4")
    fmt.Println("Printer created...")


	defaultLabelType := 1
	defaultLabelDensity := 2
	defaultQuantity := 1

	var img image.Image
	// Load image from file
    fmt.Println("Loading image...")
	file, err := os.Open("./image.jpg")
	if err != nil {
		log.Fatal(err)
	}
	img, _, err = image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

    imgWidth := img.Bounds().Dx()
    imgHeight := img.Bounds().Dy()
    logger.LogInfo("Loaded image with width: ", imgWidth, " and height: ", imgHeight, "px")
    // if imgWidth != 96 || imgHeight >=600 {
    //     fmt.Println("Image must have 96px width and 600px height max")
    //     return
    // }
    if imgWidth / imgHeight > 1 {
        fmt.Println("Image must have portrait orientation")
        return
    }


    fmt.Println("Image loaded...")
	printer.PrintLabel(img, defaultLabelType, defaultLabelDensity, defaultQuantity)
    fmt.Println("Label printed...")

}
