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
	logger.LogInfo("Connected to", comPort)
	return printer
}

func (n *NiimbotPrinter) sendCodeAndConfirm(code int, data []byte, offset int) bool {
	pkt := n.SerialSocket.Transcieve(code, data, offset)
	if pkt == nil {
		logger.LogError("Error sending code and confirming", code, data, offset)
		panic("Error sending code and confirming")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) SetLabelType(labelType int) bool {
	if labelType > 3 || labelType < 1 {
		logger.LogError("Invalid label type", labelType)
		panic("Invalid label type")
	}
	logger.LogDebug("Setting label type", labelType)
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.SET_LABEL_TYPE, []byte{byte(labelType)}, 16)
}

func (n *NiimbotPrinter) SetLabelDensity(density int) bool {
	if density > 3 || density < 1 {
		logger.LogError("Invalid label density", density)
		panic("Invalid label density")
	}
	logger.LogDebug("Setting label density", density)
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.SET_LABEL_DENSITY, []byte{byte(density)}, 16)
}

func (n *NiimbotPrinter) StartPrint() bool {
	logger.LogDebug("Starting print")
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.START_PRINT, []byte{0x01}, 1)
}

func (n *NiimbotPrinter) EndPrint() bool {
	logger.LogDebug("Ending print")
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.END_PRINT, []byte{0x01}, 1)
}

func (n *NiimbotPrinter) StartPagePrint() bool {
	logger.LogDebug("Starting page print")
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.START_PAGE_PRINT, []byte{0x01}, 1)
}

func (n *NiimbotPrinter) EndPagePrint() bool {
	logger.LogDebug("Ending page print")
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.END_PAGE_PRINT, []byte{0x01}, 1)
}

func (n *NiimbotPrinter) AllowPrintClear() bool {
	logger.LogDebug("Allowing print clear")
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.ALLOW_PRINT_CLEAR, []byte{0x01}, 16)
}

func (n *NiimbotPrinter) SetDimension(w int, h int) bool {
	width := helpers.ShortToByteArray(w)
	height := helpers.ShortToByteArray(h)

	payload := make([]byte, 0)
	payload = append(payload, height...)
	payload = append(payload, width...)

	logger.LogDebug("Setting dimension", w, h)

	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.SET_DIMENSION, payload, 1)
}

func (n *NiimbotPrinter) SetQuantity(quantity int) bool {
	logger.LogDebug("Setting quantity", quantity)
	return n.sendCodeAndConfirm(packets.NiimbotD11RequestCodePacket.SET_QUANTITY, helpers.ShortToByteArray(quantity), 1)
}

func (n *NiimbotPrinter) GetNextPageUpdate() int {
	logger.LogDebug("Getting print status")
    var pkt *packets.NiimbotPacket

    for i := 0; i < 300; i++ {
        pkt = n.SerialSocket.WaitUntilCode(packets.NiimbotD11ResponseCodePacket.PAGE_PRINT_DONE)
        if pkt !=  nil {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }
    logger.LogDebug("Received page print done packet", pkt.ToBytes())
    return int(pkt.Data[1])
}

func (n *NiimbotPrinter) SendImage(pkts []packets.NiimbotPacket) bool {
	logger.LogDebug("Sending image")
	pkt := n.SerialSocket.TranscieveBlock(packets.NiimbotD11RequestCodePacket.IMAGE_CONFIRM, pkts, 0)
	if pkt == nil {
		panic("Error sending image")
	}
	return int(pkt.Data[0]) != 0
}

func (n *NiimbotPrinter) WaitPrintFinish(pageNumber int) bool {
    for {
        currentPage := n.GetNextPageUpdate()
        if currentPage == pageNumber {
            break
        }
    }

    return true
}

func (n *NiimbotPrinter) PrintLabel(img image.Image, labelType int, labelDensity int, quantity int) {
	if img.Bounds().Dx() > 96 || img.Bounds().Dy() > 330 {
		logger.LogError("Image cannot have more than 96px width and 330px height")
		return
	}

	if img.Bounds().Dx()/img.Bounds().Dy() > 1 {
		logger.LogError("Image must have portrait orientation")
		return
	}

	imagePackets := image_encoder.EncodeForPrintingWithConfirmation(img)

	n.SetLabelType(labelType)
	n.SetLabelDensity(labelDensity)
	n.StartPrint()
	n.AllowPrintClear()

	n.StartPagePrint()

	n.SetDimension(img.Bounds().Dx(), img.Bounds().Dy())
	n.SetQuantity(quantity)
	n.SendImage(imagePackets)

	n.EndPagePrint()

	n.WaitPrintFinish(quantity)
	n.EndPrint()

    logger.LogInfo("Printed", quantity, "labels")
}
