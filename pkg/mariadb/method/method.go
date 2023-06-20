package method

import (
	"database/sql"
	"recorder/pkg/mariadb"
)

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	return mariadb.DB.Query(query, args...)

}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	return mariadb.DB.Exec(query, args...)
}
