package logger

import (
	"strings"
	"unicode/utf8"
)

func MaskEmail(email string) string {
	at := strings.LastIndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return "***"
	}
	local := email[:at]
	_, firstRuneSize := utf8.DecodeRuneInString(local)
	visible := local[:firstRuneSize]
	if firstRuneSize < len(local) {
		visible += "***"
	}
	return visible + email[at:]
}
