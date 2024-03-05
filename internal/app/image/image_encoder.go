package image_encoder

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/helpers"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/logger"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/packets"
)

func EncodeForPrinting(img image.Image) []packets.NiimbotPacket {
	logger.LogInfo("Encoding image for printing")
	grayscaled := imaging.Grayscale(img)
	logger.LogInfo("Image grayscaled")
	inverted := imaging.Invert(grayscaled)
	logger.LogInfo("Image inverted")

	const threshold = 128 // Adjust the threshold value [0, 255] as needed
	binImg := imaging.AdjustFunc(inverted, func(c color.NRGBA) color.NRGBA {
		avg := (int(c.R) + int(c.G) + int(c.B)) / 3
		if avg < threshold {
			return color.NRGBA{0, 0, 0, 255} // Black
		}
		return color.NRGBA{255, 255, 255, 255} // White
	})

	// Save image for debugging
	imaging.Save(binImg, "binarized.png")

	logger.LogInfo("Image binarized")
	byteArray := imageToByteArray(binImg)
	logger.LogInfo("Image converted to byte array")
	niimbotPackets := make([]packets.NiimbotPacket, img.Bounds().Dy())

	for y := 0; y < img.Bounds().Dy(); y++ {
		lineData := make([]byte, 12)
		min := (y * 12)
		max := (y + 1) * 12

		if min >= len(byteArray) {
			// TODO - Check if this is necessary
			break
		}
		if max > len(byteArray) {
			max = len(byteArray)
		}

		copy(lineData, byteArray[min:max])

		var counts [3]int

		for x := 0; x < 3; x++ {
			counts[x] = countBitsOfBytes(lineData[x*4 : (x+1)*4])
		}

		header := make([]byte, 6)
		lineBytes := helpers.ShortToByteArray(y)

		header[0] = lineBytes[0]
		header[1] = lineBytes[1]
		header[2] = byte(counts[0])
		header[3] = byte(counts[1])
		header[4] = byte(counts[2])
		header[5] = byte(1)

		niimbotPackets[y] = packets.NiimbotPacket{
			Type: byte(packets.NiimbotRequestCodePacket.SET_IMAGE_DATA),
			Data: append(header, lineData...),
		}
	}
	logger.LogInfo("Image encoded for printing")
	return niimbotPackets
}

func countBitsOfBytes(data []byte) int {
	n := int(binary.BigEndian.Uint32(data))
	n = (n & 0x55555555) + ((n & 0xAAAAAAAA) >> 1)
	n = (n & 0x33333333) + ((n & 0xCCCCCCCC) >> 2)
	n = (n & 0x0F0F0F0F) + ((n & 0xF0F0F0F0) >> 4)
	n = (n & 0x00FF00FF) + ((n & 0xFF00FF00) >> 8)
	n = (n & 0x0000FFFF) + ((n & 0xFFFF0000) >> 16)
	return n
}

func imageToByteArray(img image.Image) []byte {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	byteArraySize := (width*height + 7) / 8
	byteArray := make([]byte, byteArraySize)

	bytePosition := 0
	bitPosisition := 0
	currentByte := byte(0)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(x, y)

			isWhite := isWhite(pixel)
			if isWhite {
				currentByte = currentByte | byte(1<<(7-bitPosisition))
			}

			bitPosisition++
			if bitPosisition == 8 {
				byteArray[bytePosition] = currentByte
				bytePosition++
				bitPosisition = 0
				currentByte = 0
			}
		}
	}

	if bitPosisition != 0 {
		byteArray[bytePosition] = currentByte
	}

	return byteArray
}

func isWhite(pixel color.Color) bool {
	// get pixels between 0 to 255
	r, g, b, _ := pixel.RGBA()

	r = r >> 8
	g = g >> 8
	b = b >> 8

	return r > 128 && g > 128 && b > 128
}

func EncodeForPrintingWithConfirmation(img image.Image) []packets.NiimbotPacket {
	binImg := prepareImage(img)

	niimbotPackets := make([]packets.NiimbotPacket, 0)
	sliceSize := 200

	imgHeight := img.Bounds().Dy()
	imgWidth := img.Bounds().Dx()

	for ySlice := 0; ySlice < imgHeight; ySlice += sliceSize {

		hSlice := min(imgHeight, ySlice+sliceSize)
		minBounds, maxBounds := sliceBounds(ySlice, imgWidth, len(binImg))

		yNext := ySlice
		dataNext := binImg[minBounds:maxBounds]

		for yNext < hSlice {
			y := yNext
			data := dataNext

			yNext++
			for yNext < hSlice {
				minBounds, maxBounds := sliceBounds(yNext, imgWidth, len(binImg))
				dataNext = binImg[minBounds:maxBounds]

				yNext++
				if bytes.Compare(data, dataNext) != 0 {
					break
				}
			}

			niimbotPackets = append(niimbotPackets, packetImageData(y, imgWidth, data, yNext-y))
		}

	}
	return niimbotPackets
}

func packetImageData(y, width int, data []byte, n int) packets.NiimbotPacket {

	buffer := new(bytes.Buffer)
	heightBytes := helpers.ShortToByteArray(y)

	buffer.WriteByte(heightBytes[0])
	buffer.WriteByte(heightBytes[1])

	indexes := make([]int, 0)
	for x := 0; x < width; x += 32 {
		start_indexes := len(indexes)
		for b := 0; b < 32; b++ {
			if len(data) > x+b && data[x+b] != 0 {
				indexes = append(indexes, x+b)
			}
		}
		buffer.WriteByte(byte(len(indexes) - start_indexes))
	}
	buffer.WriteByte(byte(n))

	if len(indexes) == 0 {
		data := make([]byte, 0)
		lineBytes := helpers.ShortToByteArray(y)
		data = append(data, lineBytes[0])
		data = append(data, lineBytes[1])
		data = append(data, byte(n))

		return packets.NiimbotPacket{
			Type: byte(packets.NiimbotRequestCodePacket.IMAGE_CLEAR),
			Data: data,
		}
	}

	// If buffer is small, send indexes instead of bitmaps
	if len(indexes)*2 < width/8 {
		for _, index := range indexes {
			indexBytes := helpers.ShortToByteArray(index)
			buffer.WriteByte(indexBytes[0])
			buffer.WriteByte(indexBytes[1])
		}
		return packets.NiimbotPacket{
			Type: byte(packets.NiimbotRequestCodePacket.SET_IMAGE),
			Data: buffer.Bytes(),
		}
	}

	for x := 0; x < width; x += 8 {
		bits := byte(0)
		for b := 0; b < 8; b++ {
			if len(data) > x+b && data[x+b] != 0 {
				bits |= 1 << (7 - b)
			}
		}
        buffer.WriteByte(bits)
	}


	return packets.NiimbotPacket{
		Type: byte(packets.NiimbotRequestCodePacket.SET_IMAGE_DATA),
		Data: buffer.Bytes(),
	}
}

func prepareImage(img image.Image) []byte {
	grayscaled := imaging.Grayscale(img)
	inverted := imaging.Invert(grayscaled)

    binArray := make([]byte, 0) 

    for y := 0; y < img.Bounds().Dy(); y++ {
        for x := 0; x < img.Bounds().Dx(); x++ {
            pixel := inverted.At(x, y)
            isWhite := isWhite(pixel)
            if isWhite {
                binArray = append(binArray, 1)
            } else {
                binArray = append(binArray, 0)
            }
        }
    }

    return binArray
}

func sliceBounds(ySlice, imgWidth, dataLength int) (int, int) {
	min := (ySlice * imgWidth)
	max := (ySlice + 1) * imgWidth

	if min >= dataLength {
		return dataLength, dataLength
	}
	if max > dataLength {
		max = dataLength
	}
	return min, max
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
