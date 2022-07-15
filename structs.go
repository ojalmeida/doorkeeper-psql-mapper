package mapper

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"regexp"
	"strings"
)

type behavior struct {
	pathMapping pathMapping
	keyMappings []keyMapping
}

type pathMapping struct {
	path  string
	table string
}

type keyMapping struct {
	parameter  string
	column     string
	columnType string
	primary    bool
}

type response struct {
	Status int             `json:"status,omitempty"`
	Msg    string          `json:"msg,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

type PsqlMapper struct {
	conn *sqlx.DB

	behaviors  []behavior
	pathPrefix string

	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

type PsqlMapperConfig struct {
	DbConnectionString string
	PathPrefix         string
}

func (pm *PsqlMapper) SetInfoLogger(logger *log.Logger) {

	pm.info = logger
}

func (pm *PsqlMapper) SetWarnLogger(logger *log.Logger) {

	pm.warn = logger

}

func (pm *PsqlMapper) SetErrorLogger(logger *log.Logger) {

	pm.error = logger

}

func (pm PsqlMapper) Name() string {

	return "psql-mapper"
}

func (pm *PsqlMapper) MapDB() (err error) {

	err, tables := getTables(pm.conn)

	if err != nil {
		return
	}

	for _, table := range tables {

		var primaryKey string

		err, primaryKey = getPrimary(pm.conn, table)

		if err != nil {
			return
		}

		b := behavior{}

		path := strcase.ToSnake(table)

		b.pathMapping = pathMapping{
			path:  fmt.Sprintf("%s/%s", pm.pathPrefix, path),
			table: table,
		}

		b.pathMapping.path = strings.ReplaceAll(b.pathMapping.path, "//", "/")

		err2, columns := getColumns(pm.conn, table)

		if err2 != nil {
			return err2
		}

		for _, column := range columns {

			parameter := strcase.ToSnake(column)

			var columnType string

			err2, columnType = GetType(pm.conn, column)

			if err2 != nil {
				return err2
			}

			km := keyMapping{
				parameter:  parameter,
				column:     column,
				columnType: columnType,
				primary:    false,
			}

			if column == primaryKey {

				km.primary = true
			}

			b.keyMappings = append(b.keyMappings, km)

		}

		pm.behaviors = append(pm.behaviors, b)

	}

	return

}

func (pm *PsqlMapper) Configure(config PsqlMapperConfig) {

	pm.conn = sqlx.MustConnect("postgres", config.DbConnectionString)
	pm.pathPrefix = config.PathPrefix

}

func (p pathMapping) match(uri string) bool {

	return regexp.MustCompile(p.path).MatchString(uri)

}
