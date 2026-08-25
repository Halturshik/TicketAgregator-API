package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"
	"unicode"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func verifyDocument(secret []byte, documentType string, number string, checkedAt time.Time) (string, string) {
	fingerprint := documentFingerprint(secret, documentType, number)
	if documents.IsAutoVerified(documentType) {
		return documents.StatusVerified, fingerprint
	}

	period := checkedAt.UTC().Format(documents.VerificationPeriodLayout)
	digest := hmacDigest(secret, "verification:"+fingerprint+":"+period)
	if binary.BigEndian.Uint16(digest[:2])%documents.VerificationScale < documents.VerificationRejectRate {
		return documents.StatusRejected, fingerprint
	}
	return documents.StatusVerified, fingerprint
}

func documentFingerprint(secret []byte, documentType string, number string) string {
	normalized := strings.ToLower(strings.TrimSpace(documentType)) + ":" + normalizeDocumentNumber(number)
	return hex.EncodeToString(hmacDigest(secret, normalized))
}

func normalizeDocumentNumber(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}
		return -1
	}, value)
}

func hmacDigest(secret []byte, value string) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
