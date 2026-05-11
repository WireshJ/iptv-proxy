package utils

import (
	"log"
	"os"
	"strings"
)

func DebugLog(format string, v ...interface{}) {
	if strings.EqualFold(os.Getenv("DEBUG"), "true") {
		log.Printf("[DEBUG] "+format, v...)
	}
}
