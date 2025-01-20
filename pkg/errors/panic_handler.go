package errors

import (
	"fmt"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
)

type Process func()

func SafeRun(proc Process) (err error) {
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				if asErr, ok := recovered.(error); ok {
					err = asErr
				} else {
					err = errors.New(fmt.Sprintf("%v", recovered))
				}
			}
		}()

		proc()

	}()
	return
}
