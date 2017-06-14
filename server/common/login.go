package common

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"net/http"

	"github.com/jtblin/go-ldap-client"
	"gopkg.in/appleboy/gin-jwt.v2"
	jwt3 "gopkg.in/dgrijalva/jwt-go.v3"
)

// GetAuthMiddleware returns a gin middleware for JWT with cookie based auth
func GetAuthMiddleware() *jwt.GinJWTMiddleware {
	key := os.Getenv("SESSION_KEY")

	if len(key) == 0 {
		log.Fatal("Env variable 'SESSION_KEY' must be specified")
	}

	return &jwt.GinJWTMiddleware{
		Realm:         "CLOUD_SSP",
		Key:           []byte(key),
		Timeout:       time.Hour,
		MaxRefresh:    time.Hour,
		Authenticator: ldapAuthenticator,
		Authorizator: func(userId string, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.Abort()
			c.SetCookie("token", "", -1, "", "", false, true)
			if message == "Cookie token empty" {
				message = ""
			}
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Error": message,
			})
		},
		TokenLookup: "cookie:token",
		TimeFunc:    time.Now,
	}
}

func ldapAuthenticator(userID string, password string, c *gin.Context) (string, bool) {
	ldapHost := os.Getenv("LDAP_URL")
	ldapBind := os.Getenv("LDAP_BIND_DN")
	ldapBindPw := os.Getenv("LDAP_BIND_CRED")
	ldapFilter := os.Getenv("LDAP_FILTER")
	ldapSearchBase := os.Getenv("LDAP_SEARCH_BASE")

	client := &ldap.LDAPClient{
		Base:         ldapSearchBase,
		Host:         ldapHost,
		Port:         389,
		UseSSL:       false,
		SkipTLS:      true,
		BindDN:       ldapBind,
		BindPassword: ldapBindPw,
		UserFilter:   ldapFilter,
	}
	// It is the responsibility of the caller to close the connection
	defer client.Close()

	ok, _, err := client.Authenticate(userID, password)
	if err != nil {
		log.Printf("Error authenticating user %s: %+v", userID, err)
	}
	if !ok {
		log.Printf("Authenticating failed for user %s", userID)
	}
	return userID, ok
}

// CookieLoginHandler handles a cookie based JWT token
func CookieLoginHandler(mw *jwt.GinJWTMiddleware, c *gin.Context) {
	// Initial middleware default setting.
	mw.MiddlewareInit()

	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) == 0 || len(password) == 0 {
		mw.Unauthorized(c, http.StatusBadRequest, "Benutzername / Passwort nicht angegeben")
		return
	}

	if mw.Authenticator == nil {
		mw.Unauthorized(c, http.StatusInternalServerError, "Internes Problem")
		return
	}

	userID, ok := mw.Authenticator(username, password, c)

	if !ok {
		mw.Unauthorized(c, http.StatusUnauthorized, "Benutzername / Passwort nicht korrekt")
		return
	}

	// Create the token
	token := jwt3.New(jwt3.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt3.MapClaims)

	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(username) {
			claims[key] = value
		}
	}

	if userID == "" {
		userID = username
	}

	expire := mw.TimeFunc().Add(mw.Timeout)
	claims["id"] = userID
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = mw.TimeFunc().Unix()

	tokenString, err := token.SignedString(mw.Key)

	if err != nil {
		mw.Unauthorized(c, http.StatusUnauthorized, "Token konnte nicht erstellt werden")
		return
	}

	c.SetCookie("token", tokenString, 0, "", "", false, true)

	c.Redirect(http.StatusFound, "/auth/")
}
