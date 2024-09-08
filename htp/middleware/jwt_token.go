package middleware

import (
	"errors"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// 一些常量
var (
	TokenExpired     error = errors.New("Token is expired")
	TokenNotValidYet error = errors.New("Token not active yet")
	TokenMalformed   error = errors.New("That's not even a token")
	TokenInvalid     error = errors.New("Token is invalid")
)

// 自定义荷载接口
type IMetaClaimer interface {
	jwt.Claims
	TheUID() string
	TheUName() string
	TheStandardClaims() *jwt.StandardClaims
}

// Create StandardClaims
func StandardClaims(ttl time.Duration) jwt.StandardClaims {
	return jwt.StandardClaims{
		ExpiresAt: int64(time.Now().Add(ttl).Unix()),
		Issuer:    "htp.cloud",
		Subject:   "go",
	}
}

// ==================== JWT签名工具结构
func JWT(cust IMetaClaimer, key []byte) *_jwt { return &_jwt{cust, reflect.TypeOf(cust).Elem(), key} }

type _jwt struct {
	claims   IMetaClaimer
	claimTyp reflect.Type
	signKey  []byte
}

func (j *_jwt) CreateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, j.claims)
	return token.SignedString(j.signKey)
}

func (j *_jwt) ParseToken(tokenString string) (IMetaClaimer, error) {
	// claims := reflect.New(j.claimTyp).Interface()
	token, err := jwt.ParseWithClaims(tokenString, j.claims, func(token *jwt.Token) (interface{}, error) {
		return j.signKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(IMetaClaimer); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

func (j *_jwt) RefreshToken(tokenString string, ttl time.Duration) (string, error) {
	//pClaim := reflect.New(j.claimTyp).Interface()
	token, err := jwt.ParseWithClaims(tokenString, j.claims, func(token *jwt.Token) (interface{}, error) {
		return j.signKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(IMetaClaimer); ok && token.Valid {
		claims.TheStandardClaims().ExpiresAt = time.Now().Add(ttl).Unix()
		j.claims = claims
		return j.CreateToken()
	}
	return "", TokenInvalid
}
