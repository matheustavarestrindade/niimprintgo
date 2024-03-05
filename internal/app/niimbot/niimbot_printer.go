package niimbot

import (
	"image"
	"time"

	"github.com/matheustavarestrindade/niimprintgo/internal/app/helpers"
	image_encoder "github.com/matheustavarestrindade/niimprintgo/internal/app/image"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/logger"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/packets"
	serialsocket "github.com/matheustavarestrindade/niimprintgo/internal/app/socket"
)

type NiimbotPrinter struct {
	SerialSocket *serialsocket.SerialSocket
}

func NewNiimbotPrinter(comPort string) *NiimbotPrinter {
	printer := &NiimbotPrinter{
		SerialSocket: serialsocket.NewSerialSocket(comPort),
	}
	printer.SerialSocket.Connect()
	return printer
}

func (n *NiimbotPrinter) setLabelType(labelType int) bool {
	if labelType > 3 || labelType < 1 {
		panic("Invalid label type")
	}
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.SET_LABEL_TYPE, []byte{byte(labelType)}, 16)
	if pkt == nil {
		panic("Error setting label type")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) setLabelDensity(density int) bool {
	if density > 3 || density < 0 {
		panic("Invalid label density")
	}
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.SET_LABEL_DENSITY, []byte{byte(density)}, 16)
	if pkt == nil {
		panic("Error setting label density")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) startPrint() bool {
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.START_PRINT, []byte{0x01}, 1)
	if pkt == nil {
		panic("Error starting print")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) endPrint() bool {
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.END_PRINT, []byte{0x01}, 1)
	if pkt == nil {
		panic("Error ending print")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) startPagePrint() bool {
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.START_PAGE_PRINT, []byte{0x01}, 1)
	if pkt == nil {
		panic("Error starting page print")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) endPagePrint() bool {
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.END_PAGE_PRINT, []byte{0x01}, 1)
	if pkt == nil {
		panic("Error ending page print")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) allowPrintClear() bool {
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.ALLOW_PRINT_CLEAR, []byte{0x01}, 16)
	if pkt == nil {
		panic("Error allowing print clear")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) setDimension(w int, h int) bool {

	width := helpers.ShortToByteArray(w)
	height := helpers.ShortToByteArray(h)

	payload := make([]byte, 0)
    payload = append(payload, height...)
	payload = append(payload, width...)

    logger.LogInfo("Payload dimension", payload)

	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.SET_DIMENSION, payload, 1)
	if pkt == nil {
		panic("Error setting dimension")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) setQuantity(quantity int) bool {
	payload := helpers.ShortToByteArray(quantity)
	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.SET_QUANTITY, payload, 1)
	if pkt == nil {
		panic("Error setting quantity")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) getPrintStatus() map[string]int {
	logger.LogInfo("Getting print status\n\n\n")

	pkt := n.SerialSocket.Transcieve(packets.NiimbotRequestCodePacket.GET_PRINT_STATUS, []byte{0x01}, 16)
	if pkt == nil {
		panic("Error getting print status")
	}

	logger.LogInfo("PrintStatusPacket", pkt.ToBytes())
	logger.LogInfo("PrintStatusPacketData", pkt.Data)
	pageBytes := make([]byte, 2)
	pageBytes[0] = pkt.Data[0]
	pageBytes[1] = pkt.Data[1]

	page := int(pageBytes[0])<<8 | int(pageBytes[1])
	progress1 := int(pkt.Data[2])
	progress2 := int(pkt.Data[3])

	logger.LogInfo("PrintStatus: Page", page, "Progress1", progress1, "Progress2", progress2)

	return map[string]int{
		"page":      page,
		"progress1": progress1,
		"progress2": progress2,
	}

}

func (n *NiimbotPrinter) sendImage(pkts []packets.NiimbotPacket) bool {
    pkt := n.SerialSocket.TranscieveBlock(packets.NiimbotRequestCodePacket.IMAGE_CONFIRM, pkts, 0)
    if pkt == nil {
        panic("Error sending image")
    }
    return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) PrintLabel(img image.Image, labelType int, labelDensity int, quantity int) {

	logger.LogInfo("Encoding image")
	imagePackets := image_encoder.EncodeForPrintingWithConfirmation(img)

    sleepTime := 50 * time.Millisecond

    time.Sleep(sleepTime)
    logger.LogInfo("\n\n")
	logger.LogInfo("Setting labelType")
	n.setLabelType(labelType)
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)

	logger.LogInfo("Setting labelDensity")
	n.setLabelDensity(labelDensity)
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)
	logger.LogInfo("Starting print")
	n.startPrint()
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)
	logger.LogInfo("Allowing print clear")
	n.allowPrintClear()
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)
	logger.LogInfo("Starting page print")
	n.startPagePrint()
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)

	logger.LogInfo("Setting dimension width heigth", img.Bounds().Dx(), img.Bounds().Dy())
	n.setDimension(img.Bounds().Dx(), img.Bounds().Dy())
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)
	logger.LogInfo("Setting quantity", quantity)
	n.setQuantity(quantity)
    logger.LogInfo("\n\n")
    time.Sleep(sleepTime)
	logger.LogInfo("Sending image packets")

    n.sendImage(imagePackets)

	logger.LogInfo("Ending page print")
	n.endPagePrint()

	logger.LogInfo("Waiting for print to finish")
	status := n.getPrintStatus()
	logger.LogInfo("PrintStatus", status)

	for status["page"] != quantity {
		logger.LogInfo("Waiting for print to finish", status)
		time.Sleep(100 * time.Millisecond)
		status = n.getPrintStatus()
	}

	n.endPrint()
}
