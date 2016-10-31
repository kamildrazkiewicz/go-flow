package goflow

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	gf := New()
	if gf == nil {
		t.Error("New() error")
	}
}

func TestAdd1(t *testing.T) {
	gf := New()
	gf.Add("test", []string{"dep1"}, func(res map[string]interface{}) (interface{}, error) {
		return "test result", nil
	})
	_, err := gf.Do()

	if err.Error() != "Error: Function \"dep1\" not exists!" {
		t.Error("Not existing function error")
	}
}

func TestAdd2(t *testing.T) {
	gf := New()
	gf.Add("test", []string{"test"}, func(res map[string]interface{}) (interface{}, error) {
		return "test result", nil
	})
	_, err := gf.Do()

	if err.Error() != "Error: Function \"test\" depends of it self!" {
		t.Error("Self denepdency error")
	}
}

func TestDo1(t *testing.T) {
	gf := New()
	gf.Add("test", []string{}, func(res map[string]interface{}) (interface{}, error) {
		return "test result", nil
	})
	res, err := gf.Do()

	if err != nil || res["test"] != "test result" {
		t.Error("Incorrect result")
	}
}

func TestDo2(t *testing.T) {
	var shouldBeFalse bool = false
	gf := New()
	gf.Add("first", []string{}, func(res map[string]interface{}) (interface{}, error) {
		time.Sleep(time.Second * 1)
		shouldBeFalse = true
		return "first result", nil
	})
	gf.Add("second", []string{"first"}, func(res map[string]interface{}) (interface{}, error) {
		shouldBeFalse = false
		return "second result", nil
	})
	_, err := gf.Do()

	if err != nil || shouldBeFalse == true {
		t.Error("Incorrect goroutines execution order")
	}
}

func TestDo3(t *testing.T) {
	gf := New()
	gf.Add("first", []string{}, func(res map[string]interface{}) (interface{}, error) {
		return "first result", nil
	})
	gf.Add("second", []string{"first"}, func(res map[string]interface{}) (interface{}, error) {
		return "second result", nil
	})
	res, err := gf.Do()

	firstResult := res["first"]
	secondResult := res["second"]

	if err != nil || firstResult != "first result" || secondResult != "second result" {
		t.Error("Incorrect results")
	}
}

func TestDo4(t *testing.T) {
	gf := New()
	gf.Add("first", []string{}, func(res map[string]interface{}) (interface{}, error) {
		return "first result", errors.New("some error")
	})
	gf.Add("second", []string{"first"}, func(res map[string]interface{}) (interface{}, error) {
		return "second result", nil
	})
	_, err := gf.Do()

	if err.Error() != "some error" {
		t.Error("Incorrect error value")
	}
}
