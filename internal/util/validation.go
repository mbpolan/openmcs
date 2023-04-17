package util

// ValidNameChars is the set of characters which can be used in player and NPC names.
var ValidNameChars = []byte{
	'_', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l',
	'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y',
	'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
}

// ValidChatChars is the set of characters which can be used in player chat messages.
var ValidChatChars = []byte{
	' ', 'e', 't', 'a', 'o', 'i', 'h', 'n', 's', 'r', 'd', 'l', 'u',
	'm', 'w', 'c', 'y', 'f', 'g', 'p', 'b', 'v', 'k', 'x', 'j', 'q',
	'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', '!',
	'?', '.', ',', ':', ';', '(', ')', '-', '&', '*', '\\', '\'', '@',
	'#', '+', '=', '\243', '$', '%', '"', '[', ']',
}

// ChatCharCode returns the code point for a chat character. If the character is not valid, -1 is returned.
func ChatCharCode(ch byte) int {
	for i, c := range ValidChatChars {
		if c == ch {
			return i
		}
	}

	return -1
}