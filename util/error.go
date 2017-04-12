package util

import (
	"log"
)

// CheckErr logs a fatal error if err is not nil
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
