package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

type JWTClaims struct {
	Payload
	jwt.RegisteredClaims
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, ErrInvalidKeySize
	}

	return &JWTMaker{
		secretKey: secretKey,
	}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload := NewPayload(username, duration)

	claims := JWTClaims{
		*payload,
		jwt.RegisteredClaims{
			Issuer:    "server",
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// check if the token is valid, also convert string to []byte
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	// decode jwt token into JWTClaims struct
	jwtToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	// extract JWTClaims from decoded jwt token
	claims, ok := jwtToken.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return &claims.Payload, nil
}
