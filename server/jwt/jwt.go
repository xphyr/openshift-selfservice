package jwt

import (
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"time"
	"fmt"
)

func validate(protectedPage http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){
		//Validate the token and if it passes call the protected handler below.
		protectedPage(res, req)
	})
}

func validate(protectedPage http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){

		// If no Auth cookie is set then return a 404 not found
		cookie, err := req.Cookie("Auth")
		if err != nil {
			http.NotFound(res, req)
			return
		}

		// Return a Token using the cookie
		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error){
			// Make sure token's signature wasn't changed
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected siging method")
			}
			return []byte("secret"), nil
		})
		if err != nil {
			http.NotFound(res, req)
			return
		}

		// Grab the tokens claims and pass it into the original request
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), MyKey, *claims)
			page(res, req.WithContext(ctx))
		} else {
			http.NotFound(res, req)
			return
		}
	})
}

func setToken(res http.ResponseWriter, req *http.Request) {

	// Expires the token and cookie in 1 hour
	expireToken := time.Now().Add(time.Hour * 1).Unix()
	expireCookie := time.Now().Add(time.Hour * 1)

	// We'll manually assign the claims but in production you'd insert values from a database
	claims := Claims {
		"myusername",
		jwt.StandardClaims {
			ExpiresAt: expireToken,
			Issuer:    "localhost:9000",
		},
	}

	// Create the token using your claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signs the token with a secret.
	signedToken, _ := token.SignedString([]byte("secret"))

	// Place the token in the client's cookie
	cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true}
	http.SetCookie(res, &cookie)

	// Redirect the user to his profile
	http.Redirect(res, req, "/profile", 307)
}
