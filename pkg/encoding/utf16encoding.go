package encoding

import (
	"encoding/base64"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func EncodeStringToUtf16(data string) string {
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	utf16Bytes, _, _ := transform.Bytes(encoder, []byte(data))
	encodedCommand := base64.StdEncoding.EncodeToString(utf16Bytes)
	return encodedCommand

	// utf16Bytes := utf16.Encode([]rune(data))
	// encodedCommand := base64.StdEncoding.EncodeToString(encodeUTF16LE(utf16Bytes))
	// return encodedCommand
}

// func encodeUTF16LE(utf16Bytes []uint16) []byte {
// 	bytes := make([]byte, len(utf16Bytes)*2)
// 	for i, val := range utf16Bytes {
// 		bytes[i*2] = byte(val)
// 		bytes[i*2+1] = byte(val >> 8)
// 	}
// 	return bytes
// }
