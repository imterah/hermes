package users

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/backend/api/db"
	"git.terah.dev/imterah/hermes/backend/api/permissions"
	"git.terah.dev/imterah/hermes/backend/api/state"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

type UserRemovalRequest struct {
	Token string `validate:"required"`
	UID   *uint  `json:"uid"`
}

func SetupRemoveUser(state *state.State) {
	state.Engine.POST("/api/v1/users/remove", func(c *gin.Context) {
		var req UserRemovalRequest

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

		user, err := state.JWT.GetUserFromJWT(req.Token)

		if err != nil {
			if err.Error() == "token is expired" || err.Error() == "user does not exist" {
				c.JSON(http.StatusForbidden, gin.H{
					"error": err.Error(),
				})

				return
			} else {
				log.Warnf("Failed to get user from the provided JWT token: %s", err.Error())

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to parse token",
				})

				return
			}
		}

		uid := user.ID

		if req.UID != nil {
			uid = *req.UID

			if uid != user.ID && !permissions.UserHasPermission(user, "users.remove") {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Missing permissions",
				})

				return
			}
		}

		// Make sure the user exists first if we have a custom UserID

		if uid != user.ID {
			var customUser *db.User
			userRequest := state.DB.DB.Where("id = ?", uid).Find(customUser)

			if userRequest.Error != nil {
				log.Warnf("failed to find if user exists or not: %s", userRequest.Error.Error())

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to find if user exists",
				})

				return
			}

			userExists := userRequest.RowsAffected > 0

			if !userExists {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "User doesn't exist",
				})

				return
			}
		}

		state.DB.DB.Select("Tokens", "Permissions", "Proxys", "Backends").Where("id = ?", uid).Delete(user)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})
}
