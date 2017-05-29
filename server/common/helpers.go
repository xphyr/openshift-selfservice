package common

import (
	"strconv"
	"log"
	"github.com/gin-gonic/gin"
	"gopkg.in/appleboy/gin-jwt.v2"
	"os"
)

func ValidateIntInput(maxValue string, input string) (bool, string) {
	maxInt, err := strconv.Atoi(maxValue)
	if (err != nil) {
		log.Fatal("Could not parse 'MAX' value of", maxValue)
	}

	inputInt, err := strconv.Atoi(input)
	if (err != nil) {
		return false, "Bitte eine gültige Zahl eintragen"
	}

	if (inputInt > maxInt) {
		return false, "Du kannst maximal " + maxValue + " eintragen"
	}

	return true, ""
}

func GetUserName(c *gin.Context) string {
	jwtClaims := jwt.ExtractClaims(c)
	return jwtClaims["id"].(string)
}

func DebugMode() bool {
	mode := os.Getenv("GIN_MODE")

	return mode != "release"
}