package utils

import (
	"database/sql"
	"strconv"
)

func NullInt32ToString(ni sql.NullInt32) string {
	if ni.Valid {
		return strconv.FormatInt(int64(ni.Int32), 10)
	}
	return ""
}

func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
