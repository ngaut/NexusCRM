package constants

// HTTP and API constants
const (
	// Content types
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html"

	// HTTP Headers
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderXRequestID    = "X-Request-ID"

	// Auth
	BearerPrefix = "Bearer "

	// Response Keys
	ResponseError     = "error"
	ResponseSuccess   = "success"
	ResponseItems     = "items"
	ResponseLayout    = "layout"
	ResponseDashboard = "dashboard"
)

// Query parameter constants
const (
	ParamLimit    = "limit"
	ParamOffset   = "offset"
	ParamOrderBy  = "orderBy"
	ParamOrderDir = "orderDir"
	ParamSearch   = "search"
	ParamFilters  = "filters"
	ParamFields   = "fields"
	ParamInclude  = "include"

	// Pagination defaults
	DefaultLimit    = 25
	DefaultMaxLimit = 1000
	DefaultOrderDir = "DESC"
)

// Sort directions
const (
	SortASC  = "ASC"
	SortDESC = "DESC"
)

// Record states
const (
	IsDeletedTrue  = 1
	IsDeletedFalse = 0
)

// Default values (see defaults.go for additional defaults)
const (
	DefaultUserID  = "system"
	DefaultProfile = "Standard User"
)

// Log levels
const (
	LogLevelDebug   = "DEBUG"
	LogLevelInfo    = "INFO"
	LogLevelWarning = "WARNING"
	LogLevelError   = "ERROR"
)

// Object categories
const (
	CategoryStandard = "Standard"
	CategoryCustom   = "Custom"
	CategorySystem   = "System"
)

// Permission operations
const (
// Permissions defined in z_generated.go
// PermissionRead, PermissionCreate, PermissionEdit, PermissionDelete
// PermissionViewAll, PermissionModifyAll
)
