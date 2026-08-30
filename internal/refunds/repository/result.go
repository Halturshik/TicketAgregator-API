package repository

import (
	"database/sql"
	"fmt"
)

func requireAffectedRows(result sql.Result, expected int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != expected {
		return fmt.Errorf("affected rows = %d, want %d", affected, expected)
	}
	return nil
}
