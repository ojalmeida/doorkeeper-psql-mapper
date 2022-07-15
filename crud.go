package mapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/exp/slices"
	"strings"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicateKey = errors.New("item identifier not unique")
var ErrIdentifierUpdate = errors.New("item identifier can not be updated")

func get(ctx context.Context, conn *sqlx.DB, b behavior, params map[string]string) (err error, obj json.RawMessage) {

	var query string

	var knownParams []string    // id, name, ...
	var knowParamTypes []string // serial, character varying, ...
	var whereQuery string       // join(whereParams, "AND")
	var whereParams []string    // ("\"key\" = cast('value' as $param_type)",...)

	for _, mapping := range b.keyMappings {

		knownParams = append(knownParams, mapping.parameter)
		knowParamTypes = append(knowParamTypes, mapping.columnType)

	}

	if len(params) != 0 {

		for i, param := range knownParams {

			if params[param] != "" {

				whereParams = append(whereParams, fmt.Sprintf("\"%s\" = cast('%s' as %s)", param, params[param], knowParamTypes[i]))
			}

		}

		whereQuery = fmt.Sprintf("WHERE %s", strings.Join(whereParams, " AND "))

	}

	query = fmt.Sprintf("SELECT %s FROM %s %s;", strings.Join(knownParams, ", "), b.pathMapping.table, whereQuery)

	rows, err := conn.QueryxContext(ctx, query)

	if err != nil {

		return

	}

	var results []map[string]any

	for rows.Next() {

		var result = map[string]any{}

		err = rows.MapScan(result)

		if err != nil {
			continue
		}

		results = append(results, result)
	}

	obj, err = json.Marshal(results)

	return
}

func getByID(ctx context.Context, conn *sqlx.DB, b behavior, id string) (err error, obj json.RawMessage) {

	var query string

	var knownParams []string    // id, name, ...
	var knowParamTypes []string // serial, character varying, ...

	for _, mapping := range b.keyMappings {

		knownParams = append(knownParams, mapping.parameter)
		knowParamTypes = append(knowParamTypes, mapping.columnType)

	}

	var idColumnName string
	var idColumnType string

	for _, km := range b.keyMappings {

		if km.primary {

			idColumnName = km.column
			break

		}

	}

	idColumnType = knowParamTypes[slices.Index(knownParams, idColumnName)]

	for _, mapping := range b.keyMappings {

		knownParams = append(knownParams, mapping.parameter)
		knowParamTypes = append(knowParamTypes, mapping.columnType)

	}

	query = fmt.Sprintf("SELECT %s FROM %s WHERE %s = cast('%s' as %s);",
		strings.Join(knownParams, ", "),
		b.pathMapping.table, idColumnName,
		id,
		idColumnType,
	)

	rows, err := conn.QueryxContext(ctx, query)

	if err != nil {

		return

	}

	var results []map[string]any

	for rows.Next() {

		var result = map[string]any{}

		err = rows.MapScan(result)

		if err != nil {
			continue
		}

		results = append(results, result)
	}

	obj, err = json.Marshal(results)

	return
}

func create(ctx context.Context, conn *sqlx.DB, b behavior, params map[string]string) (err error, obj json.RawMessage) {

	var query string

	var knownParams []string    // id, name, ...
	var knowParamTypes []string // serial, character varying, ...

	for _, mapping := range b.keyMappings {

		knownParams = append(knownParams, mapping.parameter)
		knowParamTypes = append(knowParamTypes, mapping.columnType)

	}

	var toInsertFields []string
	var toInsertFieldValues []string
	var toInsertFieldValueTypes []string

	for i := range knownParams {

		if params[knownParams[i]] != "" {

			toInsertFields = append(toInsertFields, knownParams[i])
			toInsertFieldValues = append(toInsertFieldValues, params[knownParams[i]])
			toInsertFieldValueTypes = append(toInsertFieldValueTypes, knowParamTypes[i])

		}

	}

	for i := range toInsertFieldValues {

		toInsertFieldValues[i] = fmt.Sprintf("cast('%s' as %s)", toInsertFieldValues[i], toInsertFieldValueTypes[i])

	}

	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s;",
		b.pathMapping.table,
		strings.Join(toInsertFields, ", "),
		strings.Join(toInsertFieldValues, ", "),
		strings.Join(knownParams, ", "),
	)

	rows, err := conn.QueryxContext(ctx, query)

	if err != nil {

		if pqErr, ok := err.(*pq.Error); ok {

			err = errors.New(pqErr.Code.Name())

			if err.Error() == "unique_violation" {

				err = ErrDuplicateKey

			}
		}

		return

	}

	var results []map[string]any

	for rows.Next() {

		var result = map[string]any{}

		err = rows.MapScan(result)

		if err != nil {
			continue
		}

		results = append(results, result)
	}

	obj, err = json.Marshal(results)

	return
}

func updateByID(ctx context.Context, conn *sqlx.DB, b behavior, id string, params map[string]string) (err error, obj json.RawMessage) {

	var query string

	var knownParams []string    // id, name, ...
	var knowParamTypes []string // serial, character varying, ...

	for _, mapping := range b.keyMappings {

		knownParams = append(knownParams, mapping.parameter)
		knowParamTypes = append(knowParamTypes, mapping.columnType)

	}

	var idColumnName string
	var idColumnType string

	for _, km := range b.keyMappings {

		if km.primary {

			idColumnName = km.column
			break

		}

	}

	idColumnType = knowParamTypes[slices.Index(knownParams, idColumnName)]

	var toUpdateQuery []string

	for i := range knownParams {

		if params[knownParams[i]] == idColumnName {

			err = ErrIdentifierUpdate
			return

		}

		if params[knownParams[i]] != "" {

			toUpdateQuery = append(toUpdateQuery, fmt.Sprintf("\"%s\" = cast('%s' as %s)",

				knownParams[i],
				params[knownParams[i]],
				knowParamTypes[i],
			))

		}

	}

	query = fmt.Sprintf("UPDATE %s SET %s WHERE \"%s\" = cast('%s' as %s) RETURNING %s;",
		b.pathMapping.table,
		strings.Join(toUpdateQuery, ", "),
		idColumnName,
		id,
		idColumnType,
		strings.Join(knownParams, ", "),
	)

	rows, err := conn.QueryxContext(ctx, query)

	if err != nil {

		if pqErr, ok := err.(*pq.Error); ok {

			err = errors.New(pqErr.Code.Name())

			if err.Error() == "unique_violation" {

				err = ErrDuplicateKey

			}
		}

		return

	}

	var results []map[string]any

	for rows.Next() {

		var result = map[string]any{}

		err = rows.MapScan(result)

		if err != nil {
			continue
		}

		results = append(results, result)
	}

	obj, err = json.Marshal(results)

	return
}

func deleteByID(ctx context.Context, conn *sqlx.DB, b behavior, id string) (err error) {

	var query string

	var knownParams []string    // id, name, ...
	var knowParamTypes []string // serial, character varying, ...

	for _, mapping := range b.keyMappings {

		knownParams = append(knownParams, mapping.parameter)
		knowParamTypes = append(knowParamTypes, mapping.columnType)

	}

	var idColumnName string
	var idColumnType string

	for _, km := range b.keyMappings {

		if km.primary {

			idColumnName = km.column
			break

		}

	}

	idColumnType = knowParamTypes[slices.Index(knownParams, idColumnName)]

	query = fmt.Sprintf("DELETE FROM %s WHERE \"%s\" = cast('%s' as %s);",
		b.pathMapping.table,
		idColumnName,
		id,
		idColumnType,
	)

	result, err := conn.ExecContext(ctx, query)

	if err != nil {

		if pqErr, ok := err.(*pq.Error); ok {

			err = errors.New(pqErr.Code.Name())
		}

		return

	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {

		if pqErr, ok := err.(*pq.Error); ok {

			err = errors.New(pqErr.Code.Name())
		}

		return

	}

	if rowsAffected < 1 {

		err = ErrNotFound

	}

	return
}
