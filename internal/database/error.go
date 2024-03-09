package database

import (
	"database/sql"

	"github.com/lib/pq"
)

const (
	UniqueViolation = "23505"
)

var ErrRecordNotFound = sql.ErrNoRows

var ErrUniqueViolation = &pq.Error{
	Code: UniqueViolation,
}
