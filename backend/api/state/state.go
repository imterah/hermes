package state

import (
	"git.terah.dev/imterah/hermes/backend/api/db"
	"git.terah.dev/imterah/hermes/backend/api/jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type State struct {
	DB        *db.DB
	JWT       *jwt.JWTCore
	Engine    *gin.Engine
	Validator *validator.Validate
}

func New(db *db.DB, jwt *jwt.JWTCore, engine *gin.Engine) *State {
	return &State{
		DB:        db,
		JWT:       jwt,
		Engine:    engine,
		Validator: validator.New(),
	}
}
