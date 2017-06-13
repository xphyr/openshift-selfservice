package common

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/appleboy/gin-jwt.v2"

	"errors"
	"fmt"
	"os"
)

// ValidateIntInput checks if a the value is bigger than the specified maxValue
func ValidateIntInput(maxValue string, input string) error {
	maxInt, err := strconv.Atoi(maxValue)
	if err != nil {
		log.Fatal("Could not parse 'MAX' value of", maxValue)
	}

	inputInt, err := strconv.Atoi(input)
	if err != nil {
		return errors.New("Bitte eine gültige Zahl eintragen")
	}

	if inputInt > maxInt {
		return fmt.Errorf("Du kannst maximal %v eintragen", maxValue)
	}

	return nil
}

// GetUserName returns the username based of the gin.Context
func GetUserName(c *gin.Context) string {
	jwtClaims := jwt.ExtractClaims(c)
	return jwtClaims["id"].(string)
}

// DebugMode returns if gin is running in debug mode
func DebugMode() bool {
	mode := os.Getenv("GIN_MODE")

	return mode != "release"
}
