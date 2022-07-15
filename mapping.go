package mapper

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func getTables(conn *sqlx.DB) (err error, tables []string) {

	err = conn.Select(&tables, `SELECT tablename
									FROM pg_catalog.pg_tables
									WHERE schemaname != 'pg_catalog' AND 
    								schemaname != 'information_schema'`)

	return

}

func getColumns(conn *sqlx.DB, table string) (err error, columns []string) {

	err = conn.Select(&columns, `SELECT column_name
										FROM information_schema.columns
										WHERE table_name = $1`, table)

	return

}

func GetType(conn *sqlx.DB, column string) (err error, columnType string) {

	err = conn.Get(&columnType, `SELECT data_type
										FROM information_schema.columns
										WHERE column_name = $1`, column)

	return

}

func getPrimary(conn *sqlx.DB, table string) (err error, primary string) {

	err = conn.Get(&primary, `SELECT a.attname
									FROM   pg_index i
									JOIN   pg_attribute a ON a.attrelid = i.indrelid
									AND a.attnum = ANY(i.indkey)
									WHERE  i.indrelid = $1::regclass
									AND    i.indisprimary;`, table)

	if errors.Is(err, sql.ErrNoRows) {

		err = nil

	}

	return

}
