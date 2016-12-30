package goflow

import (
	"fmt"
	"sync"
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
	Deps []string
	Ctr  int
	Fn   func(map[string]interface{}) (interface{}, error)
	C    chan interface{}
	once sync.Once
}

func (fs *flowStruct) Close() {
	fs.once.Do(func() {
		close(fs.C)
	})
}

func (fs *flowStruct) init() {
	fs.C = make(chan interface{}, fs.Ctr)
}

// New flow struct
func New() *flow {
	return &flow{
		funcs: make(map[string]*flowStruct),
	}
}

func (f *flow) Add(name string, d []string, fn func(res map[string]interface{}) (interface{}, error)) *flow {
	f.funcs[name] = &flowStruct{
		Deps: d,
		Fn:   fn,
		Ctr:  1, // prevent deadlock
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
			f.funcs[dep].Ctr++
		}
	}
	return f.do()
}

func (f *flow) do() (map[string]interface{}, error) {
	var lastErr error
	res := make(map[string]interface{}, len(f.funcs))

	for name, v := range f.funcs {
		v.init()
		go func(name string, fs *flowStruct) {
			defer func() { fs.Close() }()
			results := make(map[string]interface{}, len(fs.Deps))

			// drain dependency results
			for _, dep := range fs.Deps {
				results[dep] = <-f.funcs[dep].C
			}

			r, err := fs.Fn(results)
			if err != nil {
				// close all channels
				for _, v := range f.funcs {
					v.Close()
				}
				lastErr = err
				return
			}
			// exit if error
			if lastErr != nil {
				return
			}
			for i := 0; i < fs.Ctr; i++ {
				fs.C <- r
			}
		}(name, v)
	}

	// wait for all
	for name, fs := range f.funcs {
		res[name] = <-fs.C
	}

	return res, lastErr
}
