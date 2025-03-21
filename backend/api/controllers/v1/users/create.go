package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"git.terah.dev/imterah/hermes/backend/api/db"
	permissionHelper "git.terah.dev/imterah/hermes/backend/api/permissions"
	"git.terah.dev/imterah/hermes/backend/api/state"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserCreationRequest struct {
	Name     string `validate:"required"`
	Email    string `validate:"required"`
	Password string `validate:"required"`
	Username string `validate:"required"`
	IsBot    bool
}

func SetupCreateUser(state *state.State) {
	state.Engine.POST("/api/v1/users/create", func(c *gin.Context) {
		if !signupEnabled && !unsafeSignup {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Signing up is not enabled at this time.",
			})

			return
		}

		var req UserCreationRequest

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to parse body: %s", err.Error()),
			})

			return
		}

		if err := state.Validator.Struct(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to validate body: %s", err.Error()),
			})

			return
		}

		var user *db.User
		userRequest := state.DB.DB.Where("email = ? OR username = ?", req.Email, req.Username).Find(&user)

		if userRequest.Error != nil {
			log.Warnf("failed to find if user exists or not: %s", userRequest.Error.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to find if user exists",
			})

			return
		}

		userExists := userRequest.RowsAffected > 0

		if userExists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User already exists",
			})

			return
		}

		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

		if err != nil {
			log.Warnf("Failed to generate password for client upon signup: %s", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate password hash",
			})

			return
		}

		permissions := []db.Permission{}

		for _, permission := range permissionHelper.DefaultPermissionNodes {
			permissionEnabledState := false

			if unsafeSignup || strings.HasPrefix(permission, "routes.") || permission == "permissions.see" {
				permissionEnabledState = true
			}

			permissions = append(permissions, db.Permission{
				PermissionNode: permission,
				HasPermission:  permissionEnabledState,
			})
		}

		tokenRandomData := make([]byte, 80)

		if _, err := rand.Read(tokenRandomData); err != nil {
			log.Warnf("Failed to read random data to use as token: %s", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate refresh token",
			})

			return
		}

		user = &db.User{
			Email:       req.Email,
			Username:    req.Username,
			Name:        req.Name,
			IsBot:       &req.IsBot,
			Password:    base64.StdEncoding.EncodeToString(passwordHashed),
			Permissions: permissions,
			Tokens: []db.Token{
				{
					Token:          base64.StdEncoding.EncodeToString(tokenRandomData),
					DisableExpiry:  forceNoExpiryTokens,
					CreationIPAddr: c.ClientIP(),
				},
			},
		}

		if result := state.DB.DB.Create(&user); result.Error != nil {
			log.Warnf("Failed to create user: %s", result.Error.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to add user into database",
			})

			return
		}

		jwt, err := state.JWT.Generate(user.ID)

		if err != nil {
			log.Warnf("Failed to generate JWT: %s", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate refresh token",
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"token":        jwt,
			"refreshToken": base64.StdEncoding.EncodeToString(tokenRandomData),
		})
	})
}
