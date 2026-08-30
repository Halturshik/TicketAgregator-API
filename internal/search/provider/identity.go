package provider

import (
	"fmt"

	"github.com/google/uuid"
)

func stableID(seed int64, index int, kind string) string {
	value := fmt.Sprintf("%d:%d:%s", seed, index, kind)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(value)).String()
}
