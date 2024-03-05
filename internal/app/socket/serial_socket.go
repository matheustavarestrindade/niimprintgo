package serialsocket

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/matheustavarestrindade/niimprintgo/internal/app/logger"
	"github.com/matheustavarestrindade/niimprintgo/internal/app/packets"
	"go.bug.st/serial"
)

type SerialSocket struct {
	ComPort string

	connection serial.Port
	pktBuffer  []byte
	bytesRead  int
	bufferPos  int
	startBytes []byte
	endBytes   []byte
}

func NewSerialSocket(comPort string) *SerialSocket {
	return &SerialSocket{
		ComPort: comPort,

		pktBuffer:  make([]byte, 1024),
		bytesRead:  0,
		bufferPos:  0,
		startBytes: []byte{0x55, 0x55},
		endBytes:   []byte{0xAA, 0xAA},
	}
}

func (ss *SerialSocket) Connect() {
	if ss.connection != nil {
		ss.connection.Close()
	}

	port, err := serial.Open(ss.ComPort, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		log.Fatal(err)
	}
	ss.connection = port
	fmt.Println("Connected to", ss.ComPort)
}

func (ss *SerialSocket) Send(data []byte) {
	_, err := ss.connection.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (ss *SerialSocket) Read() {
	n, err := ss.connection.Read(ss.pktBuffer)
	logger.LogInfo("Read", n, "bytes", ss.pktBuffer[:n])
	if err != nil {
		log.Fatal(err)
	}
	ss.bufferPos = 0
	ss.bytesRead = n
}

func (ss *SerialSocket) Close() {
	ss.connection.Close()
}

func (ss *SerialSocket) TranscieveBlock(code int, data []packets.NiimbotPacket, responseOffset int) *packets.NiimbotPacket {
    logger.LogInfo("Transcieve ", code, data, responseOffset)
	responseCode := responseOffset + code

	logger.LogInfo("Waiting for response code", responseCode)

	logger.LogInfo("\nSending packet block") 
    for _, pkt := range data {
        logger.LogInfo("Sending packet", pkt.ToBytes())
        ss.Send(pkt.ToBytes())
    }
    logger.LogInfo("Sent packet block\n")

	for i := 0; i < 6; i++ {
		for _, pkt := range ss.recv() {
			switch int(pkt.Type) {
			case 219:
				panic("Error: IllegalArgument")
			case 0:
				panic("Error: NotImplement")
			case responseCode:
				return &pkt
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
func (ss *SerialSocket) Transcieve(code int, data []byte, responseOffset int) *packets.NiimbotPacket {
    logger.LogInfo("Transcieve ", code, data, responseOffset)
	responseCode := responseOffset + code

	logger.LogInfo("Waiting for response code", responseCode)
	packet := packets.NiimbotPacket{
		Type: byte(code),
		Data: data,
	}

	logger.LogInfo("Sending packet", packet.ToBytes())
	ss.Send(packet.ToBytes())
	for i := 0; i < 6; i++ {
		for _, pkt := range ss.recv() {
			switch int(pkt.Type) {
			case 219:
				panic("Error: IllegalArgument")
			case 0:
				panic("Error: NotImplement")
			case responseCode:
				return &pkt
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (ss *SerialSocket) readUntil(target []byte) bool {
	logger.LogInfo("Reading until", target)
	localBuffer := make([]byte, len(target))

	for ss.bufferPos < len(ss.pktBuffer) {
		// Shift old bytes
		for i := 0; i < len(target)-1; i++ {
			localBuffer[i] = localBuffer[i+1]
		}
		localBuffer[len(target)-1] = ss.readByteFromBuffer()
		if bytes.Equal(localBuffer, target) {
			logger.LogInfo("Found target")
			return true
		}
	}
	return false
}

func (ss *SerialSocket) readByteFromBuffer() byte {
	if ss.bufferPos >= ss.bytesRead && ss.bytesRead == len(ss.pktBuffer) {
		ss.Read()
	}
	pkt := ss.pktBuffer[ss.bufferPos]
	ss.bufferPos++
	return pkt
}

func (ss *SerialSocket) recv() []packets.NiimbotPacket {
	pkts := []packets.NiimbotPacket{}
	for {
		logger.LogInfo("Reading from serial socket")
		ss.Read()
		if ss.bytesRead <= 0 {
			logger.LogInfo("No bytes read")
			break
		}
		logger.LogInfo("Bytes read", ss.bytesRead)

		for ss.readUntil(ss.startBytes) {
			packet := ss.extractPacketFromBuffer()
			if packet == nil {
				continue
			}
			pkt, err := packets.FromBytes(packet)
			if err != nil {
				log.Fatal(err)
			}
			pkts = append(pkts, *pkt)
		}
		if ss.bufferPos >= ss.bytesRead && ss.bytesRead < len(ss.pktBuffer) {
			break
		}
	}
	return pkts
}

func (ss *SerialSocket) extractPacketFromBuffer() []byte {
	localBuffer := []byte{}
	localBuffer = append(localBuffer, ss.startBytes...)

	for {
		byt := ss.readByteFromBuffer()
		localBuffer = append(localBuffer, byt)

		if len(localBuffer) >= 4 && bytes.Compare(localBuffer[len(localBuffer)-2:], ss.endBytes) == 0 {
			break
		}

		if ss.bufferPos >= ss.bytesRead {
			return nil
		}
	}
	return localBuffer
}
