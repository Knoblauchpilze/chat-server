package repositories

import (
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ExtractForeignKeyViolation(t *testing.T) {
	testCases := []struct {
		err           error
		expectedKey   string
		expectedFound bool
	}{
		{},
		{err: fmt.Errorf("some error")},
		{err: &db.DatabaseError{Code: db.ErrUniqueConstraintViolation, Constraint: "other-constraint"}},
		{
			err:           &db.DatabaseError{Code: db.ErrForeignKeyValidation, Constraint: "my-constraint"},
			expectedKey:   "my-constraint",
			expectedFound: true,
		},
		{
			err: fmt.Errorf("wrapped: %w", &db.DatabaseError{
				Code: db.ErrForeignKeyValidation, Constraint: "my-constraint",
			}),
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
