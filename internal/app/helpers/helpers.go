package helpers


func ShortToByteArray(short int) []byte {
	if short < 0 || short > 65535 {
		panic("Short out of range")
	}
	return []byte{byte(short / 256), byte(short % 256)}
}
