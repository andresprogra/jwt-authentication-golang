package authentication

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/andresprogra/jwtauth/models"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

// Se firman los token con una llave privada
// Se verifican con una llave pública

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func init() {
	privateBytes, err := ioutil.ReadFile("./private.rsa")
	if err != nil {
		log.Fatal("No se pudo leer el archivo privado")
	}

	publicBytes, err := ioutil.ReadFile("./public.rsa.pub")
	if err != nil {
		log.Fatal("No se pudo leer el archivo público")
	}

	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err != nil {
		log.Fatal("No se pudo hacer el parse a privatekey")
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err != nil {
		log.Fatal("No se pudo hacer el parse a publickey")
	}
}

func GenerateJWT(user models.User) string {
	claims := models.Claim{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
			Issuer:    "Auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	result, err := token.SignedString(privateKey)
	if err != nil {
		log.Fatal("No se pudo firmar el token")
	}

	return result
}
func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		fmt.Fprintln(w, "Error al leer el usuario %s", err)
		return
	}
	if user.Name == "andres" && user.Password == "andres" {
		user.Password = ""
		user.Role = "admin"

		token := GenerateJWT(user)
		result := models.ResponseToken{token}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintln(w, "No se pudo hacer el marshal del json %s", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResult)
	} else {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Usuario o clave inválidos")
	}
}

func ValidateToken(w http.ResponseWriter, r *http.Request) {
	token, err := request.ParseFromRequestWithClaims(r, request.OAuth2Extractor, &models.Claim{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		switch err.(type) {
		case *jwt.ValidationError:
			vErr := err.(*jwt.ValidationError)
			switch vErr.Errors {
			case jwt.ValidationErrorExpired:
				fmt.Fprintln(w, "Su token ha expirado")
				return
			case jwt.ValidationErrorSignatureInvalid:
				fmt.Fprintln(w, "La firma del token no coincide")
				return
			default:
				fmt.Fprintln(w, "Su token no es válido")
				return
			}
		default:
			fmt.Fprintln(w, "Su token no es válido")
			return
		}
	}

	if token.Valid {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Bienvenido al sistema")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Su token no es válido")
	}
}
