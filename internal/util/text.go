package util

import "fmt"

// ValidNameChars is the set of characters which can be used in player and NPC names.
var validNameChars = []byte{
	'_', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l',
	'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y',
	'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
}

// ValidChatChars is the set of characters which can be used in player chat messages.
var validChatChars = []byte{
	' ', 'e', 't', 'a', 'o', 'i', 'h', 'n', 's', 'r', 'd', 'l', 'u',
	'm', 'w', 'c', 'y', 'f', 'g', 'p', 'b', 'v', 'k', 'x', 'j', 'q',
	'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', '!',
	'?', '.', ',', ':', ';', '(', ')', '-', '&', '*', '\\', '\'', '@',
	'#', '+', '=', '\243', '$', '%', '"', '[', ']',
}

// validNameCharSetLen defines how many unique characters comprise the set of valid username characters.
var validNameCharSetLen = uint64(len(validNameChars))

// ChatCharCode returns the code point for a chat character. If the character is not valid, -1 is returned.
func ChatCharCode(ch byte) int {
	for i, c := range validChatChars {
		if c == ch {
			return i
		}
	}

	return -1
}

// EncodeName encodes a player name into a long integer.
func EncodeName(text string) uint64 {
	name := uint64(0)

	for _, ch := range text {
		name *= validNameCharSetLen

		if ch >= 'A' && ch <= 'Z' {
			name += uint64((ch + 1) - 'A')
		} else if ch >= 'a' && ch <= 'z' {
			name += uint64((ch + 1) - 'a')
		} else if ch >= '0' && ch <= '9' {
			name += uint64((ch + 27) - '0')
		}
	}

	// normalize the value
	for name%validNameCharSetLen == 0 && name != 0 {
		name /= validNameCharSetLen
	}

	return name
}

// DecodeName decodes a player name from a long integer.
func DecodeName(n uint64) (string, error) {
	var bytes []byte

	for n != 0 {
		code := int(n - (n/validNameCharSetLen)*validNameCharSetLen)
		if code >= len(validNameChars) {
			return "", fmt.Errorf("invalid name code: %d", n)
		}

		bytes = append([]byte{validNameChars[code]}, bytes...)
		n /= validNameCharSetLen
	}

	return string(bytes), nil
}

// EncodeChat encodes a chat message using the dictionary of valid chat characters.
func EncodeChat(text string) []byte {
	var encoded []byte

	lastCh := -1
	for _, ch := range text {
		code := ChatCharCode(byte(ch))

		if code > 12 {
			code += 0xC3
		}

		if lastCh == -1 {
			if code < 13 {
				lastCh = code
			} else {
				encoded = append(encoded, byte(code))
			}
		} else if code < 13 {
			encoded = append(encoded, byte(lastCh<<4|code))
			lastCh = -1
		} else {
			encoded = append(encoded, byte(lastCh<<4|(code>>4)))
			lastCh = code & 0x0F
		}
	}

	if lastCh != -1 {
		encoded = append(encoded, byte(lastCh<<4))
	}

	return encoded
}

// DecodeChat decodes a chat message using the dictionary of valid chat characters.
func DecodeChat(raw []byte) (string, error) {
	var text []byte

	lastCh := byte(0x00)
	for i := 0; i < len(raw); i++ {
		ch := raw[i]

		// each byte contains up to two distinct characters
		hi := ch >> 4
		lo := ch & 0x0F

		// if the last byte had a continuation, form a single code point from the two parts
		// if the high bits are a value greater than 13, treat the entire byte as a single code point
		// otherwise treat the high bits as a single code point
		code := -1
		if lastCh > 0 {
			code = int(((lastCh << 4) | hi) - 0xC3)
			text = append(text, validChatChars[code])

			lastCh = 0x00
		} else if hi >= 13 {
			code = int(ch - 0xC3)
			text = append(text, validChatChars[code])

			// skip the rest of the byte
			continue
		} else {
			code = int(hi)
			text = append(text, validChatChars[code])
		}

		// if the low bits are a value greater than 13, store them and expect the next byte to complete the code point
		// otherwise treat the low bits as a single code point
		if lo >= 13 {
			lastCh = lo
		} else {
			code = int(lo)
			text = append(text, validChatChars[code])
		}
	}

	// if some bits are left over, form a code point from them
	if lastCh > 0x00 {
		code := int(lastCh)
		text = append(text, validChatChars[code])
	}

	return string(text), nil
}
