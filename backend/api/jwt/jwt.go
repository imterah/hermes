package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"git.terah.dev/imterah/hermes/backend/api/db"
	"github.com/golang-jwt/jwt/v5"
)

var (
	DevelopmentModeTimings = time.Duration(60*24) * time.Minute
	NormalModeTimings      = time.Duration(3) * time.Minute
)

type JWTCore struct {
	Key            []byte
	Database       *db.DB
	TimeMultiplier time.Duration
}

func New(key []byte, database *db.DB, timeMultiplier time.Duration) *JWTCore {
	jwtCore := &JWTCore{
		Key:            key,
		Database:       database,
		TimeMultiplier: timeMultiplier,
	}

	return jwtCore
}

func (jwtCore *JWTCore) Parse(tokenString string, options ...jwt.ParserOption) (*jwt.Token, error) {
	return jwt.Parse(tokenString, jwtCore.jwtKeyCallback, options...)
}

func (jwtCore *JWTCore) GetUserFromJWT(token string) (*db.User, error) {
	if jwtCore.Database == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	parsedJWT, err := jwtCore.Parse(token)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token is expired")
		} else {
			return nil, err
		}
	}

	audience, err := parsedJWT.Claims.GetAudience()

	if err != nil {
		return nil, err
	}

	if len(audience) < 1 {
		return nil, fmt.Errorf("audience is too small")
	}

	uid, err := strconv.Atoi(audience[0])

	if err != nil {
		return nil, err
	}

	user := &db.User{}
	userRequest := jwtCore.Database.DB.Preload("Permissions").Where("id = ?", uint(uid)).Find(&user)

	if userRequest.Error != nil {
		return user, fmt.Errorf("failed to find if user exists or not: %s", userRequest.Error.Error())
	}

	userExists := userRequest.RowsAffected > 0

	if !userExists {
		return user, fmt.Errorf("user does not exist")
	}

	return user, nil
}

func (jwtCore *JWTCore) Generate(uid uint) (string, error) {
	currentJWTTime := jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtCore.TimeMultiplier)),
		IssuedAt:  currentJWTTime,
		NotBefore: currentJWTTime,
		// Convert the user ID to a string, and then set it as the audience parameters only value (there's only 1 user per key)
		Audience: []string{strconv.Itoa(int(uid))},
	})

	signedToken, err := token.SignedString(jwtCore.Key)

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (jwtCore *JWTCore) jwtKeyCallback(*jwt.Token) (any, error) {
	return jwtCore.Key, nil
}
