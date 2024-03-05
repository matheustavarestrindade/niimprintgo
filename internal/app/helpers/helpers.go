package helpers

import (
	"image"
	"os"

	"github.com/matheustavarestrindade/niimprintgo/internal/app/logger"
)

func ShortToByteArray(short int) []byte {
	if short < 0 || short > 65535 {
		panic("Short out of range")
	}
	return []byte{byte(short / 256), byte(short % 256)}
}

func GetImageFromFilePath(path string) image.Image {
    logger.LogDebug("Opening image file", path)
	file, err := os.Open(path)
	if err != nil {
        logger.LogError("Error opening image file", path)
        return nil
	}
    img, _, err := image.Decode(file)
	if err != nil {
        logger.LogError("Error decoding image file", path)
        return nil
	}
    return img
}

func FileExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

