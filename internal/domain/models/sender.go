package models

import (
	"strings"
)

const (
	// Icelandic merchants ISK.
	rbSenderPrefix   = "RB"
	isbSenderPrefix  = "ISB"
	saxoSenderPrefix = "SAXO"
)

func IsISLSender(sender string) bool {
	return strings.HasPrefix(sender, rbSenderPrefix) || strings.HasPrefix(sender, isbSenderPrefix)
}

func IsSaxoSender(sender string) bool {
	return strings.HasPrefix(sender, saxoSenderPrefix)
}
