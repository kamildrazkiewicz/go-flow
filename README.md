# Goflow

[![GoDoc](http://godoc.org/github.com/kamildrazkiewicz/go-flow?status.svg)](http://godoc.org/github.com/kamildrazkiewicz/go-flow) [![License](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)](http://godoc.org/github.com/kamildrazkiewicz/go-flow) [![Release](https://img.shields.io/github/release/kamildrazkiewicz/go-flow.svg?label=Release)](http://godoc.org/github.com/kamildrazkiewicz/go-flow)




Goflow is a simply package to control goroutines execution order based on dependencies. It works similar to ```async.auto``` from [node.js async package](https://github.com/caolan/async), but for Go.

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
	"time"
)

func main() {
	f1 := func(r *goflow.Results) (interface{}, error) {
		fmt.Println("function1 started")
		time.Sleep(time.Millisecond * 1000)
		return 1, nil
	}

	f2 := func(r *goflow.Results) (interface{}, error) {
		time.Sleep(time.Millisecond * 1000)
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

Output will be:
```
function1 started
function3 started 1
function2 started 1
function4 started &map[f2:some results f3:<nil>]
&map[f1:1 f2:some results f3:<nil> f4:<nil>] <nil>
```
