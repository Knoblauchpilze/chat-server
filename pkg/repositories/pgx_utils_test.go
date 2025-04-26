package repositories

import (
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ExtractForeignKeyViolation(t *testing.T) {
	type testCase struct {
		err           error
		expectedKey   string
		expectedFound bool
	}

	pgErr := &pgconn.PgError{
		ConstraintName: "my-constraint",
	}

	testCases := []testCase{
		{
			err:           fmt.Errorf("some error"),
			expectedKey:   "",
			expectedFound: false,
		},
		{
			err:           errors.NewCode(pgx.ForeignKeyValidation),
			expectedKey:   "",
			expectedFound: false,
		},
		{
			err:           errors.WrapCode(fmt.Errorf("some error"), pgx.ForeignKeyValidation),
			expectedKey:   "",
			expectedFound: false,
		},
		{
			err:           errors.WrapCode(pgErr, pgx.ForeignKeyValidation),
			expectedKey:   "my-constraint",
			expectedFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			actual, ok := extractForeignKeyViolation(tc.err)

			assert.Equal(t, tc.expectedKey, actual)
			assert.Equal(t, tc.expectedFound, ok)
		})
	}
}
