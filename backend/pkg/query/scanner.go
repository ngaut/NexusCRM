package query

import (
	"database/sql"

	"github.com/nexuscrm/shared/pkg/models"
)

// ScanRowsToSObjects scans SQL rows into a slice of SObject maps
func ScanRowsToSObjects(rows *sql.Rows) ([]models.SObject, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]models.SObject, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		record := make(models.SObject)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}

		results = append(results, record)
	}

	return results, nil
}
