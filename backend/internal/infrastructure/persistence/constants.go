package persistence

// SQL Type Prefixes and Keywords used for Schema introspection
const (
	SQLTypeVarchar   = "VARCHAR"
	SQLTypeText      = "TEXT"
	SQLTypeChar      = "CHAR"
	SQLTypeBool      = "BOOL"
	SQLTypeBoolean   = "BOOLEAN"
	SQLTypeTinyInt   = "TINYINT"
	SQLTypeTinyInt1  = "TINYINT(1)"
	SQLTypeInt       = "INT"
	SQLTypeInteger   = "INTEGER"
	SQLTypeBigInt    = "BIGINT"
	SQLTypeSmallInt  = "SMALLINT"
	SQLTypeDecimal   = "DECIMAL"
	SQLTypeFloat     = "FLOAT"
	SQLTypeDouble    = "DOUBLE"
	SQLTypeDateTime  = "DATETIME"
	SQLTypeTimestamp = "TIMESTAMP"
	SQLTypeDate      = "DATE"
	SQLTypeJSON      = "JSON"

	// Standard SQL Type Definitions with Precision
	SQLTypeVarchar255  = "VARCHAR(255)"
	SQLTypeVarchar50   = "VARCHAR(50)"
	SQLTypeVarchar36   = "VARCHAR(36)"
	SQLTypeDecimal18_6 = "DECIMAL(18,6)"
	SQLTypeDecimal18_2 = "DECIMAL(18,2)"
	SQLTypeDecimal5_2  = "DECIMAL(5,2)"
	SQLTypeLongText    = "LONGTEXT"

	// SQL Keyword Constants
	KeywordInsertInto  = "INSERT INTO"
	KeywordValues      = "VALUES"
	KeywordOnDuplicate = "ON DUPLICATE KEY UPDATE"
	KeywordUpdate      = "UPDATE"
	KeywordSet         = "SET"
	KeywordWhere       = "WHERE"
	KeywordFrom        = "FROM"
	KeywordAnd         = "AND"
	KeywordOr          = "OR"
	KeywordSelect      = "SELECT"
	KeywordDeleteFrom  = "DELETE FROM"
	KeywordLimit       = "LIMIT"
	KeywordOrderBy     = "ORDER BY"
	KeywordAsc         = "ASC"
	KeywordDesc        = "DESC"
	KeywordTrue        = "true"
	KeywordFalse       = "false"
	KeywordNull        = "NULL"
	KeywordBegin       = "BEGIN"
	KeywordCommit      = "COMMIT"
	KeywordRollback    = "ROLLBACK"
	KeywordCreateTable = "CREATE TABLE"
	KeywordDropTable   = "DROP TABLE"
	KeywordIfExists    = "IF EXISTS"
	KeywordPrimaryKey  = "PRIMARY KEY"
	KeywordDefault     = "DEFAULT"
	KeywordLike        = "LIKE"

	// SQL Functions
	FuncNow              = "NOW()"
	FuncCurrentTimestamp = "CURRENT_TIMESTAMP"
	FuncCount            = "COUNT(*)"
)

// Query Operation Codes
const (
	OpCount   = "count"
	OpGroupBy = "group_by"
	OpSum     = "sum"
	OpAvg     = "avg"

	// Rollup Types (Uppercase)
	RollupTypeCount = "COUNT"
	RollupTypeSum   = "SUM"
	RollupTypeMin   = "MIN"
	RollupTypeMax   = "MAX"
	RollupTypeAvg   = "AVG"
)

// System Health Status
const (
	StatusHealthy      = "healthy"
	StatusUnhealthy    = "unhealthy"
	StatusOperational  = "operational"
	CreatedByBootstrap = "bootstrap"
)

// Transaction Constraints
const (
	IsoLevelSerializable = "SERIALIZABLE"
	ErrCodeDeadlock      = "1213"
	ErrCodeLockWait      = "1205"
)
