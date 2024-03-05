package packets

import (
	"errors"
	"fmt"
)

var NiimbotInfoPacket = newNiimbotInfoPackets()
var NiimbotRequestCodePacket = newNiimbotRequestCodePackets()

type NiimbotInfoPackets struct {
	DENSITY          int
	PRINTSPEED       int
	LABELTYPE        int
	LANGUAGETYPE     int
	AUTOSHUTDOWNTIME int
	DEVICETYPE       int
	SOFTVERSION      int
	BATTERY          int
	DEVICESERIAL     int
	HARDVERSION      int
}

func newNiimbotInfoPackets() *NiimbotInfoPackets {
	return &NiimbotInfoPackets{
		DENSITY:          1,
		PRINTSPEED:       2,
		LABELTYPE:        3,
		LANGUAGETYPE:     6,
		AUTOSHUTDOWNTIME: 7,
		DEVICETYPE:       8,
		SOFTVERSION:      9,
		BATTERY:          10,
		DEVICESERIAL:     11,
		HARDVERSION:      12,
	}
}

type NiimbotRequestCodePackets struct {
	GET_INFO          int
	GET_RFID          int
	HEARTBEAT         int
	SET_LABEL_TYPE    int
	SET_LABEL_DENSITY int
	START_PRINT       int
	END_PRINT         int
	START_PAGE_PRINT  int
	END_PAGE_PRINT    int
	ALLOW_PRINT_CLEAR int
	SET_DIMENSION     int
	SET_QUANTITY      int
	GET_PRINT_STATUS  int
	SET_IMAGE         int
	IMAGE_CLEAR       int
	SET_IMAGE_DATA    int
    IMAGE_CONFIRM     int
}

func newNiimbotRequestCodePackets() *NiimbotRequestCodePackets {
	return &NiimbotRequestCodePackets{
		GET_INFO:          64,
		GET_RFID:          26,
		HEARTBEAT:         220,
		SET_LABEL_TYPE:    35,
		SET_LABEL_DENSITY: 33,
		START_PRINT:       1,
		END_PRINT:         243,
		START_PAGE_PRINT:  3,
		END_PAGE_PRINT:    227,
		ALLOW_PRINT_CLEAR: 32,
		SET_DIMENSION:     19,
		SET_QUANTITY:      21,
		GET_PRINT_STATUS:  163,
		SET_IMAGE:         131,
		IMAGE_CLEAR:       132,
		SET_IMAGE_DATA:    133,
        IMAGE_CONFIRM:     211,
	}
}

type NiimbotPacket struct {
	Type byte
	Data []byte
}

func (np *NiimbotPacket) ToBytes() []byte {
	checksum := int(np.Type) ^ len(np.Data)
	for _, b := range np.Data {
		checksum = checksum ^ int(b)
	}

	// Start of packet
	packet := []byte{0x55, 0x55, np.Type}
	packet = append(packet, byte(len(np.Data)))
	packet = append(packet, np.Data...)
	packet = append(packet, byte(checksum), byte(0xAA), byte(0xAA))

	return packet
}

func (np *NiimbotPacket) ToString() {
	fmt.Println("NiimbotPacket -> Type:", np.Type, "Data:", np.Data)
}

func FromBytes(packet []byte) (*NiimbotPacket, error) {
	if packet[0] != 0x55 || packet[1] != 0x55 {
		fmt.Println("NiimbotPacket -> Invalid packet", packet)
		return nil, errors.New("Invalid packet")
	}
	if packet[len(packet)-1] != 0xaa || packet[len(packet)-2] != 0xaa {
		fmt.Println("NiimbotPacket -> Invalid packet", packet)
		return nil, errors.New("Invalid packet")
	}
	np := &NiimbotPacket{
		Type: packet[2],
		Data: packet[4 : len(packet)-3],
	}

	length := int(packet[3])

	checksum := int(np.Type) ^ length
	for _, b := range np.Data {
		checksum ^= int(b)
	}
	if byte(checksum) != packet[len(packet)-3] {
		fmt.Println("NiimbotPacket -> Invalid checksum", packet)
		return nil, errors.New("Invalid checksum")
	}
	return np, nil
}
