package image_encoder

import (
	"bytes"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/helpers"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/packets"
)

func EncodeForPrintingWithConfirmation(img image.Image) []packets.NiimbotPacket {
	binImg := convertImageToBinary(img)

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
			Type: byte(packets.NiimbotD11RequestCodePacket.IMAGE_CLEAR),
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
			Type: byte(packets.NiimbotD11RequestCodePacket.SET_IMAGE),
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
		Type: byte(packets.NiimbotD11RequestCodePacket.SET_IMAGE_DATA),
		Data: buffer.Bytes(),
	}
}

func convertImageToBinary(img image.Image) []byte {
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

func isWhite(pixel color.Color) bool {
	// get pixels between 0 to 255
	r, g, b, _ := pixel.RGBA()

	r = r >> 8
	g = g >> 8
	b = b >> 8

	return r > 128 && g > 128 && b > 128
}
