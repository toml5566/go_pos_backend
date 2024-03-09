package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/toml5566/go_pos_backend/utils"
)

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker(utils.RandString(32))
	require.NoError(t, err)

	username := utils.RandString(6)
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := time.Now().Add(duration)

	jwtToken, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	payload, err := maker.VerifyToken(jwtToken)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotEmpty(t, payload.ID)
	require.Equal(t, payload.Username, username)
	require.WithinDuration(t, payload.IssuedAt, issuedAt, time.Second)
	require.WithinDuration(t, payload.ExpiredAt, expiredAt, time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(utils.RandString(32))
	require.NoError(t, err)

	username := utils.RandString(6)
	duration := -time.Minute

	jwtToken, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	payload, err := maker.VerifyToken(jwtToken)
	require.EqualError(t, err, "token has invalid claims: token is expired")
	require.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	maker, err := NewJWTMaker(utils.RandString(32))
	require.NoError(t, err)

	// CreateToken with unsafe alg
	payload := NewPayload(utils.RandString(6), time.Minute)

	claims := JWTClaims{
		*payload,
		jwt.RegisteredClaims{
			Issuer:    "server",
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)          // changed signing method to none
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType) // changed to unsafe secret key
	require.NoError(t, err)

	// Verify token
	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, "token is unverifiable: error while executing keyfunc: token is invalid")
	require.Nil(t, payload)
}

func TestInvalidKeySize(t *testing.T) {
	maker, err := NewJWTMaker(utils.RandString(16))
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidKeySize.Error())
	require.Nil(t, maker)
}
