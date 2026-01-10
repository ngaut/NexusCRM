package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

// RespondAppError sends a standardised JSON error response using pkg/errors
func RespondAppError(c *gin.Context, err error) {
	code := errors.GetHTTPStatus(err)
	errorCode := errors.GetErrorCode(err)
	message := err.Error()

	if code >= 500 {
		log.Printf("❌ ERROR [%d] %s %s: %s", code, c.Request.Method, c.Request.URL.Path, message)
	}

	c.JSON(code, gin.H{
		"message": message,
		"code":    errorCode,
		"data":    nil,
	})
}

// BindJSON binds JSON and returns true if successful. If failed, it sends bad request error.
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		RespondAppError(c, errors.NewValidationError("body", err.Error()))
		return false
	}
	return true
}

// BindJSONStrict binds JSON and enforces strict field validation (no unknown fields).
func BindJSONStrict(c *gin.Context, obj interface{}) bool {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(obj); err != nil {
		RespondAppError(c, errors.NewValidationError("body", err.Error()))
		return false
	}
	if err := binding.Validator.ValidateStruct(obj); err != nil {
		RespondAppError(c, errors.NewValidationError("body", err.Error()))
		return false
	}
	return true
}

// HandleGetEnvelope executes a read action and returns the result wrapped in a JSON key
// Response: { [key]: result }
func HandleGetEnvelope(c *gin.Context, key string, action func() (interface{}, error)) {
	result, err := action()
	if err != nil {
		RespondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{key: result})
}

// HandleCreateEnvelope executes a create action and returns the object wrapped + message
// Response: { constants.FieldMessage: successMsg, "data": obj }
// If key is empty, defaults to "data" for consistent API response format.
func HandleCreateEnvelope(c *gin.Context, key string, successMsg string, obj interface{}, action func() error) {
	if !BindJSON(c, obj) {
		return
	}
	if err := action(); err != nil {
		RespondAppError(c, err)
		return
	}
	// Default to "data" for consistent response format
	if key == "" {
		key = "data"
	}
	c.JSON(http.StatusCreated, gin.H{
		constants.FieldMessage: successMsg,
		key:                    obj,
	})
}

// HandleUpdateEnvelope executes an update action and returns the object wrapped + message
// Response: { constants.FieldMessage: successMsg, "data": obj }
// If key is empty, defaults to "data" for consistent API response format.
func HandleUpdateEnvelope(c *gin.Context, key string, successMsg string, obj interface{}, action func() error) {
	if !BindJSON(c, obj) {
		return
	}
	if err := action(); err != nil {
		RespondAppError(c, err)
		return
	}
	// Default to "data" for consistent response format
	if key == "" {
		key = "data"
	}
	c.JSON(http.StatusOK, gin.H{
		constants.FieldMessage: successMsg,
		key:                    obj,
	})
}

// HandleDeleteEnvelope executes a delete action and returns a success message
// Response: { constants.FieldMessage: successMsg }
func HandleDeleteEnvelope(c *gin.Context, successMsg string, action func() error) {
	if err := action(); err != nil {
		RespondAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{constants.FieldMessage: successMsg})
}
