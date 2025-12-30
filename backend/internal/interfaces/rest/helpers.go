package rest

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// GetUserFromContext extracts the authenticated user from gin.Context
func GetUserFromContext(c *gin.Context) *models.UserSession {
	userInterface, exists := c.Get(constants.ContextKeyUser)
	if !exists {
		return nil
	}

	// The middleware stores auth.UserSession, need to convert to models.UserSession
	authUser := userInterface.(auth.UserSession)
	return &models.UserSession{
		ID:            authUser.ID,
		Name:          authUser.Name,
		Email:         &authUser.Email, // Convert string to *string
		ProfileID:     authUser.ProfileId,
		RoleID:        authUser.RoleId,
		IsSystemAdmin: authUser.IsSuperUser(),
	}
}

// RespondError sends a JSON error response and logs errors for debugging
func RespondError(c *gin.Context, code int, message string) {
	// Log errors (especially 500s) for debugging
	if code >= 500 {
		log.Printf("❌ ERROR [%d] %s %s: %s", code, c.Request.Method, c.Request.URL.Path, message)
	}
	c.JSON(code, gin.H{constants.ResponseError: message})
}

// BindJSON binds JSON and returns true if successful. If failed, it sends bad request error.
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		RespondError(c, http.StatusBadRequest, errors.ErrInvalidRequest.Error()+": "+err.Error())
		return false
	}
	return true
}

// HandleGetEnvelope executes a read action and returns the result wrapped in a JSON key
// Response: { [key]: result }
func HandleGetEnvelope(c *gin.Context, key string, action func() (interface{}, error)) {
	result, err := action()
	if err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{key: result})
}

// HandleCreateEnvelope executes a create action and returns the object wrapped + message
// Response: { constants.FieldMessage: successMsg, [key]: obj } (key omitted if empty)
func HandleCreateEnvelope(c *gin.Context, key string, successMsg string, obj interface{}, action func() error) {
	if !BindJSON(c, obj) {
		return
	}
	if err := action(); err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}
	response := gin.H{constants.FieldMessage: successMsg}
	if key != "" {
		response[key] = obj
	}
	c.JSON(http.StatusCreated, response)
}

// HandleUpdateEnvelope executes an update action and returns the object wrapped + message
// Response: { constants.FieldMessage: successMsg, [key]: obj } (key omitted if empty)
func HandleUpdateEnvelope(c *gin.Context, key string, successMsg string, obj interface{}, action func() error) {
	if !BindJSON(c, obj) {
		return
	}
	if err := action(); err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}
	response := gin.H{constants.FieldMessage: successMsg}
	if key != "" {
		response[key] = obj
	}
	c.JSON(http.StatusOK, response)
}

// HandleDeleteEnvelope executes a delete action and returns a success message
// Response: { constants.FieldMessage: successMsg }
func HandleDeleteEnvelope(c *gin.Context, successMsg string, action func() error) {
	if err := action(); err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{constants.FieldMessage: successMsg})
}
