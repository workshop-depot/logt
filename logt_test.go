package logt

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ Output = &StdLogget{}

func TestHere(t *testing.T) {
	funcName, fileName, fileLine, err := here(1)
	var name string
	if err != nil {
		name = ""
	} else {
		name = fmt.Sprintf("%s:%02d %s()", fileName, fileLine, funcName)
	}
	assert.NotEqual(t, "", name)
	assert.Equal(t, "logt/logt_test.go", fileName)
	assert.Condition(t, func() bool {
		if fileLine <= 0 {
			return false
		}
		return true
	})
	assert.Equal(t, "TestHere", funcName)
}

type warn []interface{}

func (w warn) Warn() string { return fmt.Sprintf("%v", w) }

func TestSmoke01(t *testing.T) {
	t.SkipNow()
	// write a mock Output for this test (and other tests)
	lg := New(NewStdLogget(), "", 0)
	lg.Println("data is ok")
	lg.Println("some more data ", 1, 2, 3)
	lg.Println("there are some errors here ", 14, errors.New("BOOM"))
	lg.Printf("%v %v\n", 10, errors.New("BOOM"))
	lg.Printf("%v %v\n", 10, "BOOM")
	var w warn
	w = append(w, "WARN")
	lg.Println(w, "BOOM")
}

func TestSmoke02(t *testing.T) {
	t.SkipNow()
	Println("data is ok")
	Println("some more data ", 1, 2, 3)
	Println("there are some errors here ", 14, errors.New("BOOM"))
	Printf("%v %v\n", 10, errors.New("BOOM"))
	Printf("%v %v\n", 10, "BOOM")
}
