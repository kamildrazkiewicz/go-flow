# Goflow

[![GoDoc](http://godoc.org/github.com/kamildrazkiewicz/go-flow?status.svg)](http://godoc.org/github.com/kamildrazkiewicz/go-flow)

Goflow is a simply package to control goroutines based on dependencies. It works like ```async.auto``` from [node.js async package](https://github.com/caolan/async), but for Go.

## Install

Install the package with:

```bash
go get github.com/kamildrazkiewicz/go-flow
```

Import it with:

```go
import "github.com/kamildrazkiewicz/go-flow"
```

and use `goflow` as the package name inside the code.

## Example

```go
package main

import (
	"fmt"
	"github.com/kamildrazkiewicz/go-flow"
)

func main() {
	f1 := func(r *goflow.Results) (interface{}, error) {
		fmt.Println("function1 started")
		return 1, nil
	}

	f2 := func(r *goflow.Results) (interface{}, error) {
		fmt.Println("function2 started", (*r)["f1"])
		return "some results", nil
	}

	f3 := func(r *goflow.Results) (interface{}, error) {
		fmt.Println("function3 started", (*r)["f1"])
		return nil, nil
	}

	f4 := func(r *goflow.Results) (interface{}, error) {
		fmt.Println("function4 started", r)
		return nil, nil
	}

	res, err := goflow.New().
		Add("f1", nil, f1).
		Add("f2", []string{"f1"}, f2).
		Add("f3", []string{"f1"}, f3).
		Add("f4", []string{"f2", "f3"}, f4).
		Do()

	fmt.Println(res, err)
}

```
