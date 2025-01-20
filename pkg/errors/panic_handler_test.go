package errors

import (
	"fmt"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var errSample = fmt.Errorf("sample error")

func TestUnit_SafeRun_CallsProcess(t *testing.T) {
	var called int

	proc := func() {
		called++
	}

	actual := SafeRun(proc)

	assert.Nil(t, actual)
	assert.Equal(t, 1, called)
}

func TestUnit_SafeRun_NoPanic(t *testing.T) {
	proc := func() {}

	var actual error

	run := func() {
		actual = SafeRun(proc)
	}

	assert.NotPanics(t, run)
	assert.Nil(t, actual)
}

func TestUnit_SafeRun_PanicWithError(t *testing.T) {
	proc := func() {
		panic(errSample)
	}

	var actual error

	run := func() {
		actual = SafeRun(proc)
	}

	assert.NotPanics(t, run)
	assert.Equal(t, errSample, actual)
}

func TestUnit_SafeRun_PanicWithRandomDatatype(t *testing.T) {
	proc := func() {
		panic(2)
	}

	var actual error

	run := func() {
		actual = SafeRun(proc)
	}

	assert.NotPanics(t, run)
	assert.Equal(t, errors.New("2"), actual)
}
