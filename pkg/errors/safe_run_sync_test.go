package errors

import (
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var errSample = fmt.Errorf("sample error")

func TestUnit_SafeRunSync_CallsProcess(t *testing.T) {
	var called int

	proc := func() {
		called++
	}

	actual := SafeRunSync(proc)

	assert.Nil(t, actual, "Actual err: %v", actual)
	assert.Equal(t, 1, called)
}

func TestUnit_SafeRunSync_NoPanic(t *testing.T) {
	proc := func() {}

	var actual error

	run := func() {
		actual = SafeRunSync(proc)
	}

	assert.NotPanics(t, run)
	assert.Nil(t, actual, "Actual err: %v", actual)
}

func TestUnit_SafeRunSync_PanicWithError(t *testing.T) {
	proc := func() {
		panic(errSample)
	}

	var actual error

	run := func() {
		actual = SafeRunSync(proc)
	}

	assert.NotPanics(t, run)
	assert.Equal(t, errSample, actual)
}

func TestUnit_SafeRunSync_PanicWithRandomDatatype(t *testing.T) {
	proc := func() {
		panic(2)
	}

	var actual error

	run := func() {
		actual = SafeRunSync(proc)
	}

	assert.NotPanics(t, run)
	assert.Equal(t, errors.New("2"), actual)
}
