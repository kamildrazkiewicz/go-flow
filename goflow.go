package goflow

import (
	"fmt"
	"sync/atomic"
)

// Flow interface
type Flow interface {
	Add(name string, d []string, fn func(res map[string]interface{}) (interface{}, error)) *flow
	Do() (map[string]interface{}, error)
}

type flow struct {
	funcs map[string]*flowStruct
}

type flowStruct struct {
	Deps    []string
	Counter int
	Fn      func(map[string]interface{}) (interface{}, error)
	closed  atomic.Value
}

// New flow struct
func New() *flow {
	f := &flow{}
	f.funcs = make(map[string]*flowStruct)
	return f
}

func (f *flow) Add(name string, d []string, fn func(res map[string]interface{}) (interface{}, error)) *flow {
	f.funcs[name] = &flowStruct{
		Deps:    d,
		Fn:      fn,
		Counter: 1, // prevent deadlock
	}
	return f
}

func (f *flow) Do() (map[string]interface{}, error) {
	for fname, fn := range f.funcs {
		for _, dep := range fn.Deps {
			// prevent self depends
			if dep == fname {

				return nil, fmt.Errorf("Error: Function \"%s\" depends of it self!", fname)
			}
			// prevent no existing dependencies
			if _, exists := f.funcs[dep]; exists == false {
				return nil, fmt.Errorf("Error: Function \"%s\" not exists!", dep)
			}
			f.funcs[dep].Counter++
		}
	}
	return f.do()
}

func (f *flow) do() (map[string]interface{}, error) {
	var lastErr error
	res := make(map[string]interface{}, len(f.funcs))

	// create flow channels
	flow := make(map[string]chan interface{}, len(f.funcs))
	for name, v := range f.funcs {
		flow[name] = make(chan interface{}, v.Counter)
	}

	for name, v := range f.funcs {
		go func(name string, fs *flowStruct) {
			defer func() {
				if true == fs.closed.Load() {
					return
				}
				close(flow[name])
			}()

			results := make(map[string]interface{}, len(fs.Deps))

			// drain dependency results
			for _, dep := range fs.Deps {
				results[dep] = <-flow[dep]
			}

			r, err := fs.Fn(results)
			if err != nil {
				// close all channels
				for name, v := range f.funcs {
					if false == v.closed.Load() {
						close(flow[name])
						v.closed.Store(true)
					}

				}
				lastErr = err

				return
			}
			if true == fs.closed.Load() {
				return
			}
			for i := 0; i < fs.Counter; i++ {
				flow[name] <- r
			}
		}(name, v)
	}

	// wait for all
	for name := range f.funcs {
		res[name] = <-flow[name]
	}

	return res, lastErr
}
