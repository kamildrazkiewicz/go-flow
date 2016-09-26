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
	f := &flow{}
	f.funcs = make(map[string]*FlowStruct)
	return f
}

func (f *flow) Add(name string, d []string, fn func(res *Results) (interface{}, error)) *flow {
	f.funcs[name] = &FlowStruct{
		Deps:    d,
		Fn:      fn,
		Counter: 1, // prevent deadlock
	}
	return f
}

func (f *flow) Do() (*Results, error) {
	for fname, fn := range f.funcs {
		for _, dep := range fn.Deps {
			// prevent self depends
			if dep == fname {
				return nil, errors.New(fmt.Sprintf("Error: Function \"%s\" depends of it self!", fname))
			}
			// prevent no existing dependencies
			if _, exists := f.funcs[dep]; exists == false {
				return nil, errors.New(fmt.Sprintf("Error: Function \"%s\" not exists!", dep))
			}
			f.funcs[dep].Counter++
		}
	}
	return f.do()
}

func (f *flow) do() (*Results, error) {
	var lastErr error
	res := make(Results, len(f.funcs))

	// create flow channels
	flow := make(map[string]chan interface{}, len(f.funcs))
	for name, v := range f.funcs {
		flow[name] = make(chan interface{}, v.Counter)
	}

	for name, v := range f.funcs {
		go func(name string, fs *FlowStruct) {
			defer close(flow[name])
			results := make(Results, len(fs.Deps))

			// drain dependency results
			for _, dep := range fs.Deps {
				results[dep] = <-flow[dep]
			}

			r, err := fs.Fn(&results)
			if err != nil {
				// close all channels
				for name, _ := range f.funcs {
					close(flow[name])
				}
				lastErr = err
				return
			}
			for i := 0; i < fs.Counter; i++ {
				flow[name] <- r
			}
		}(name, v)
	}

	// wait for all
	for name, _ := range f.funcs {
		res[name] = <-flow[name]
	}

	return &res, lastErr
}
