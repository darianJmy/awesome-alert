package database

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
)

// ExportInitializationSQL writes the schema and all initial rows from the
// database initialized by InitializeFreshDatabase. Call it only on that fresh
// build database; it is intentionally not a production backup command.
func ExportInitializationSQL(writer io.Writer) error {
	if DB == nil {
		return fmt.Errorf("database is not connected")
	}

	tables, err := listBaseTables()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(writer, "SET NAMES utf8mb4;\nSET FOREIGN_KEY_CHECKS = 0;\n\n"); err != nil {
		return err
	}
	for _, table := range tables {
		var tableName, createStatement string
		if err := DB.Raw("SHOW CREATE TABLE "+quoteIdentifier(table)).Row().Scan(&tableName, &createStatement); err != nil {
			return fmt.Errorf("read schema for %s: %w", table, err)
		}
		if _, err := fmt.Fprintf(writer, "DROP TABLE IF EXISTS %s;\n%s;\n\n", quoteIdentifier(table), createStatement); err != nil {
			return err
		}
	}
	for _, table := range tables {
		if err := exportTableRows(writer, table); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(writer, "SET FOREIGN_KEY_CHECKS = 1;")
	return err
}

func listBaseTables() ([]string, error) {
	rows, err := DB.Raw("SHOW FULL TABLES WHERE Table_type = 'BASE TABLE'").Rows()
	if err != nil {
		return nil, fmt.Errorf("list database tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table, tableType string
		if err := rows.Scan(&table, &tableType); err != nil {
			return nil, fmt.Errorf("scan database table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list database tables: %w", err)
	}
	return tables, nil
}

func exportTableRows(writer io.Writer, table string) error {
	rows, err := DB.Table(table).Rows()
	if err != nil {
		return fmt.Errorf("read initial rows from %s: %w", table, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("read columns from %s: %w", table, err)
	}
	if len(columns) == 0 {
		return nil
	}
	quotedColumns := make([]string, len(columns))
	for i, column := range columns {
		quotedColumns[i] = quoteIdentifier(column)
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		destinations := make([]interface{}, len(columns))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return fmt.Errorf("scan initial row from %s: %w", table, err)
		}
		literals := make([]string, len(values))
		for i, value := range values {
			literals[i] = mysqlLiteral(value)
		}
		if _, err := fmt.Fprintf(writer, "INSERT INTO %s (%s) VALUES (%s);\n", quoteIdentifier(table), strings.Join(quotedColumns, ", "), strings.Join(literals, ", ")); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read initial rows from %s: %w", table, err)
	}
	return nil
}

func quoteIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}

func mysqlLiteral(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return "NULL"
	case time.Time:
		return quoteMySQLString(typed.Format("2006-01-02 15:04:05.000"))
	case []byte:
		return quoteMySQLString(string(typed))
	case string:
		return quoteMySQLString(typed)
	case bool:
		if typed {
			return "1"
		}
		return "0"
	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint, float64, float32:
		return fmt.Sprint(typed)
	case sql.RawBytes:
		return quoteMySQLString(string(typed))
	default:
		return quoteMySQLString(fmt.Sprint(typed))
	}
}

func quoteMySQLString(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\x00", "\\0",
		"\n", "\\n",
		"\r", "\\r",
		"'", "\\'",
		"\x1a", "\\Z",
	)
	return "'" + replacer.Replace(value) + "'"
}
