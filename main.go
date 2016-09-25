package goflow

import (
	"errors"
	"fmt"
)

type Results map[string]interface{}

type Flow interface {
	New() *flow
}

type flow struct {
	funcs map[string]*FlowStruct
}

type FlowStruct struct {
	Deps    []string
	Counter int
	Fn      func(*Results) (interface{}, error)
}

func New() *flow {
	a := &flow{}
	a.funcs = make(map[string]*FlowStruct)
	return a
}

func (a *flow) Add(name string, d []string, f func(res *Results) (interface{}, error)) *flow {
	a.funcs[name] = &FlowStruct{
		Deps:    d,
		Fn:      f,
		Counter: 1, // prevent deadlock
	}
	return a
}

func (a *flow) Do() (*Results, error) {
	for fname, fn := range a.funcs {
		for _, dep := range fn.Deps {
			// prevent self depends
			if dep == fname {
				return nil, errors.New(fmt.Sprintf("Error: Function \"%s\" depends of it self!", fname))
			}
			// prevent no existing dependencies
			if _, exists := a.funcs[dep]; exists == false {
				return nil, errors.New(fmt.Sprintf("Error: Function \"%s\" not exists!", dep))
			}
			a.funcs[dep].Counter++
		}
	}
	return a.do()
}

func (a *flow) do() (*Results, error) {
	var lastErr error
	res := make(Results, len(a.funcs))

	// create flow channels
	flow := make(map[string]chan interface{})
	for name, _ := range a.funcs {
		flow[name] = make(chan interface{})
	}

	for name, v := range a.funcs {
		go func(name string, v *FlowStruct) {
			defer close(flow[name])
			results := make(Results, len(v.Deps))

			// drain dependency results
			for _, dep := range v.Deps {
				results[dep] = <-flow[dep]
			}

			r, err := v.Fn(&results)
			if err != nil {
				// close all channels
				for name, _ := range a.funcs {
					close(flow[name])
				}
				lastErr = err
				return
			}
			for i := 0; i < v.Counter; i++ {
				flow[name] <- r
			}
		}(name, v)
	}

	// wait for all
	for name, _ := range a.funcs {
		res[name] = <-flow[name]
	}

	return &res, lastErr
}
