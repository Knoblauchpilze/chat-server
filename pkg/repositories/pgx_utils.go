package repositories

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func extractForeignKeyViolation(err error) (foreignKey string, ok bool) {
	foreignKey = ""
	ok = false

	if !errors.IsErrorWithCode(err, pgx.ForeignKeyValidation) {
		return
	}

	cause := errors.Unwrap(err)
	if cause == nil {
		return
	}

	pgErr, ok := cause.(*pgconn.PgError)
	if !ok {
		return
	}

	foreignKey = pgErr.ConstraintName
	ok = true

	return
}
