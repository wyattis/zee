package isql

import (
	"database/sql"
	"fmt"
)

func Begin(db IBegin, fn func(tx *sql.Tx) (err error)) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	hasCommitted := false
	defer func() {
		if !hasCommitted {
			fmt.Println("rolling back")
			err2 := tx.Rollback()
			if err2 != nil {
				fmt.Println("rollback err", err2)
			}
		}
	}()
	if err = fn(tx); err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		return
	}
	hasCommitted = true
	return
}
