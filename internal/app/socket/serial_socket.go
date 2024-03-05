package serialsocket

import (
	"bytes"
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
		logger.LogError("Error opening serial port", ss.ComPort)
		panic(err)
	}
	ss.connection = port
}

func (ss *SerialSocket) Send(data []byte) {
	_, err := ss.connection.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (ss *SerialSocket) Read() {

    err := ss.connection.SetReadTimeout(200 * time.Millisecond)
    if err != nil {
        log.Fatal(err)
    }
	n, err := ss.connection.Read(ss.pktBuffer)
	logger.LogDebug("Read", n, "bytes", ss.pktBuffer[:n])
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
	logger.LogDebug("TranscieveBlock ", code, data, responseOffset)
	responseCode := responseOffset + code

	logger.LogDebug("Waiting for response code", responseCode)

	logger.LogDebug("\nSending packet block")
	for _, pkt := range data {
		logger.LogDebug("Sending packet", pkt.ToBytes())
		ss.Send(pkt.ToBytes())
	}
	logger.LogDebug("Sent packet block\n")

	for i := 0; i < 6; i++ {
		for _, pkt := range ss.recv() {
			switch int(pkt.Type) {
			case 219:
				logger.LogError("Error: IllegalArgument")
				panic("Error: IllegalArgument")
			case 0:
				logger.LogError("Error: NotImplement")
				panic("Error: NotImplement")
			case responseCode:
				return &pkt
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (ss *SerialSocket) TranscieveTimeout(code int, data []byte, responseOffset int, timeout time.Duration) *packets.NiimbotPacket {

	logger.LogDebug("Transcieve ", code, data, responseOffset)
	responseCode := responseOffset + code

	logger.LogDebug("Waiting for response code", responseCode)
	packet := packets.NiimbotPacket{
		Type: byte(code),
		Data: data,
	}

	logger.LogDebug("Sending packet", packet.ToBytes())
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

func (ss *SerialSocket) Transcieve(code int, data []byte, responseOffset int) *packets.NiimbotPacket {

	logger.LogDebug("Transcieve ", code, data, responseOffset)
	responseCode := responseOffset + code

	logger.LogDebug("Waiting for response code", responseCode)
	packet := packets.NiimbotPacket{
		Type: byte(code),
		Data: data,
	}

	logger.LogDebug("Sending packet", packet.ToBytes())
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

func (ss *SerialSocket) WaitUntilCode(code int) *packets.NiimbotPacket {
	for i := 0; i < 6; i++ {
		for _, pkt := range ss.recv() {
			switch int(pkt.Type) {
			case 219:
				panic("Error: IllegalArgument")
			case 0:
				panic("Error: NotImplement")
			case code:
				return &pkt
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
    return nil
}

func (ss *SerialSocket) readUntil(target []byte) bool {
	localBuffer := make([]byte, len(target))

	for ss.bufferPos < len(ss.pktBuffer) {
		for i := 0; i < len(target)-1; i++ {
			localBuffer[i] = localBuffer[i+1]
		}
		localBuffer[len(target)-1] = ss.readByteFromBuffer()
		if bytes.Equal(localBuffer, target) {
			logger.LogDebug("Found target", target)
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
		ss.Read()
		if ss.bytesRead <= 0 {
			break
		}
		logger.LogDebug("Read bytes from port", ss.bytesRead)
		logger.LogDebug("Reading until start bytes")
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
