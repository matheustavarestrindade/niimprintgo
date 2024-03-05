package main

import (
	"flag"

	"github.com/matheustavarestrindade/niimprintgo/internal/app/helpers"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/logger"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/niimbot"
)

type DefaultParameters struct {
	LabelType    int
	LabelDensity int
	Quantity     int
	ImagePath    string
	ComPort      string

	LoggerEnableDebug  bool
	LoggerEnableInfo   bool
	LoggerEnableError  bool
	LoggerEnableColors bool
}

func (dp *DefaultParameters) IsValidConfig() bool {
	if dp.ComPort == "" {
		logger.LogError("COM port is required")
		return false
	}
	if dp.LabelType > 3 || dp.LabelType < 1 {
		logger.LogError("Invalid label type", dp.LabelType)
		return false
	}
	if dp.LabelDensity > 3 || dp.LabelDensity < 0 {
		logger.LogError("Invalid label density", dp)
		return false
	}
	if dp.Quantity < 1 {
		logger.LogError("Invalid quantity", dp.Quantity)
		return false
	}
	if dp.ImagePath == "" {
		logger.LogError("Image path is required")
		return false
	}
	if !helpers.FileExists(dp.ImagePath) {
		logger.LogError("Image file not found", dp.ImagePath)
		return false
	}
	return true
}

func main() {

	initParams := readParams() 

	logger.ConfigureLogger(initParams.LoggerEnableInfo, initParams.LoggerEnableError, initParams.LoggerEnableDebug, initParams.LoggerEnableColors)

	if !initParams.IsValidConfig() {
		return
	}

	logger.LogInfo("Starting Niimprintgo...")
	printer := niimbot.NewNiimbotPrinter(initParams.ComPort)

	img := helpers.GetImageFromFilePath(initParams.ImagePath)
	if img == nil {
		return
	}

	logger.LogInfo("Printing label...")
	printer.PrintLabel(img, initParams.LabelType, initParams.LabelDensity, initParams.Quantity)
}

func readParams() DefaultParameters {

	initParams := DefaultParameters{}
    debug := flag.Bool("debug", false, "Enable debug logs")
    info := flag.Bool("info", true, "Enable info logs")
    error := flag.Bool("error", true, "Enable error logs")
    colors := flag.Bool("colors", true, "Enable colors in logs")
    labelType := flag.Int("labelType", 1, "Label type")
    labelDensity := flag.Int("labelDensity", 2, "Label density")
    quantity := flag.Int("quantity", 1, "Quantity")
    comPort := flag.String("comPort", "", "COM port")
    imagePath := flag.String("imagePath", "", "Image path")


	flag.Parse()

    initParams.LoggerEnableDebug = *debug 
	initParams.LoggerEnableInfo = *info 
	initParams.LoggerEnableError = *error 
	initParams.LoggerEnableColors = *colors 
	initParams.LabelType = *labelType
	initParams.LabelDensity = *labelDensity
	initParams.Quantity = *quantity
	initParams.ComPort = *comPort
	initParams.ImagePath = *imagePath

	return initParams
}
