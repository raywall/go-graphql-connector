package adapters

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type RDSAdapter struct {
	db         *sql.DB
	resultMode string
}

func NewRDSAdapter(driverName, dsn, resultMode string) (Adapter, error) {
	if driverName == "" {
		return nil, fmt.Errorf("rds driverName is required")
	}
	if dsn == "" {
		return nil, fmt.Errorf("rds dsn is required")
	}
	if resultMode == "" {
		resultMode = "one"
	}
	if resultMode != "one" && resultMode != "many" {
		return nil, fmt.Errorf("rds resultMode must be one or many")
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open rds connection: %v", err)
	}

	return &RDSAdapter{
		db:         db,
		resultMode: resultMode,
	}, nil
}

func (r *RDSAdapter) GetData(ctx context.Context, query string) (map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rds query: %v", err)
	}
	defer rows.Close()

	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rds rows: %v", err)
	}

	if r.resultMode == "many" {
		return map[string]interface{}{"items": items}, nil
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return items[0], nil
}

func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to read rds columns: %v", err)
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan rds row: %v", err)
		}

		item := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			item[column] = normalizeSQLValue(values[i])
		}
		result = append(result, item)
	}
	return result, nil
}

func normalizeSQLValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		var decoded interface{}
		if err := json.Unmarshal(typed, &decoded); err == nil {
			return decoded
		}
		return string(typed)
	default:
		return typed
	}
}
