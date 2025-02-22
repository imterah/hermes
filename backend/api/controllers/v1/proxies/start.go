package proxies

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/backend/api/backendruntime"
	"git.terah.dev/imterah/hermes/backend/api/dbcore"
	"git.terah.dev/imterah/hermes/backend/api/jwtcore"
	"git.terah.dev/imterah/hermes/backend/api/permissions"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProxyStartRequest struct {
	Token string `validate:"required" json:"token"`
	ID    uint   `validate:"required" json:"id"`
}

func StartProxy(c *gin.Context) {
	var req ProxyStartRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to parse body: %s", err.Error()),
		})

		return
	}

	if err := validator.New().Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to validate body: %s", err.Error()),
		})

		return
	}

	user, err := jwtcore.GetUserFromJWT(req.Token)
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

	if !permissions.UserHasPermission(user, "routes.start") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	var proxy *dbcore.Proxy
	proxyRequest := dbcore.DB.Where("id = ?", req.ID).Find(&proxy)

	if proxyRequest.Error != nil {
		log.Warnf("failed to find if proxy exists or not: %s", proxyRequest.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if forward rule exists",
		})

		return
	}

	proxyExists := proxyRequest.RowsAffected > 0

	if !proxyExists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Forward rule doesn't exist",
		})

		return
	}

	backend, ok := backendruntime.RunningBackends[proxy.BackendID]

	if !ok {
		log.Warnf("Couldn't fetch backend runtime from backend ID #%d", proxy.BackendID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't fetch backend runtime",
		})

		return
	}

	backendResponse, err := backend.ProcessCommand(&commonbackend.AddProxy{
		SourceIP:   proxy.SourceIP,
		SourcePort: proxy.SourcePort,
		DestPort:   proxy.DestinationPort,
		Protocol:   proxy.Protocol,
	})

	switch responseMessage := backendResponse.(type) {
	case error:
		log.Warnf("Failed to get response for backend #%d: %s", proxy.BackendID, responseMessage.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get response from backend",
		})

		return
	case *commonbackend.ProxyStatusResponse:
		if !responseMessage.IsActive {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to start proxy",
			})

			return
		}

		break
	default:
		log.Errorf("Got illegal response type for backend #%d: %T", proxy.BackendID, responseMessage)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Got invalid response from backend. Proxy was still successfully deleted",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
