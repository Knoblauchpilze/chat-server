package repositories

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
)

func extractForeignKeyViolation(err error) (string, bool) {
	dbErr, ok := db.AsDatabaseError(err)
	if !ok || dbErr.Code != db.ErrForeignKeyValidation {
		return "", false
	}

	return dbErr.Constraint, true
}
