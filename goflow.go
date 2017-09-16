package goflow

import (
	"fmt"
	"sync"
)

type Flow struct {
	funcs map[string]*flowStruct
}

type flowFunc func(res map[string]interface{}) (interface{}, error)

type flowStruct struct {
	Deps []string
	Ctr  int
	Fn   flowFunc
	C    chan interface{}
	once sync.Once
}

func (fs *flowStruct) Done(r interface{}) {
	for i := 0; i < fs.Ctr; i++ {
		fs.C <- r
	}
}

func (fs *flowStruct) Close() {
	fs.once.Do(func() {
		close(fs.C)
	})
}

func (fs *flowStruct) Init() {
	fs.C = make(chan interface{}, fs.Ctr)
}

// New flow struct
func New() *Flow {
	return &Flow{
		funcs: make(map[string]*flowStruct),
	}
}

func (flw *Flow) Add(name string, d []string, fn flowFunc) *Flow {
	flw.funcs[name] = &flowStruct{
		Deps: d,
		Fn:   fn,
		Ctr:  1, // prevent deadlock
	}
	return flw
}

func (flw *Flow) Do() (map[string]interface{}, error) {
	for name, fn := range flw.funcs {
		for _, dep := range fn.Deps {
			// prevent self depends
			if dep == name {
				return nil, fmt.Errorf("Error: Function \"%s\" depends of it self!", name)
			}
			// prevent no existing dependencies
			if _, exists := flw.funcs[dep]; exists == false {
				return nil, fmt.Errorf("Error: Function \"%s\" not exists!", dep)
			}
			flw.funcs[dep].Ctr++
		}
	}
	return flw.do()
}

func (flw *Flow) do() (map[string]interface{}, error) {
	var err error
	res := make(map[string]interface{}, len(flw.funcs))

	for _, f := range flw.funcs{
		f.Init()
	}
	for name, f := range flw.funcs {
		go func(name string, fs *flowStruct) {
			defer func() { fs.Close() }()
			results := make(map[string]interface{}, len(fs.Deps))

			// drain dependency results
			for _, dep := range fs.Deps {
				results[dep] = <-flw.funcs[dep].C
			}

			r, fnErr := fs.Fn(results)
			if fnErr != nil {
				// close all channels
				for _, fn := range flw.funcs {
					fn.Close()
				}
				err = fnErr
				return
			}
			// exit if error
			if err != nil {
				return
			}
			fs.Done(r)

		}(name, f)
	}

	// wait for all
	for name, fs := range flw.funcs {
		res[name] = <-fs.C
	}

	return res, err
}
