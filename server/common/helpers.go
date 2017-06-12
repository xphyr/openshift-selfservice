package common

import (
	"strconv"
	"log"
	"github.com/gin-gonic/gin"
	"gopkg.in/appleboy/gin-jwt.v2"
	"os"
	"errors"
	"fmt"
)

func ValidateIntInput(maxValue string, input string) (error) {
	maxInt, err := strconv.Atoi(maxValue)
	if (err != nil) {
		log.Fatal("Could not parse 'MAX' value of", maxValue)
	}

	inputInt, err := strconv.Atoi(input)
	if (err != nil) {
		return errors.New("Bitte eine gültige Zahl eintragen")
	}

	if (inputInt > maxInt) {
		return fmt.Errorf("Du kannst maximal %v eintragen", maxValue)
	}

	return nil
}

func GetUserName(c *gin.Context) string {
	jwtClaims := jwt.ExtractClaims(c)
	return jwtClaims["id"].(string)
}

func DebugMode() bool {
	mode := os.Getenv("GIN_MODE")

	return mode != "release"
}
