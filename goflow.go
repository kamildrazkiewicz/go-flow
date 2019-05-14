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
	Deps    []string
	Ctr     int
	Fn      flowFunc
	C       chan interface{}
	once    sync.Once
	ctrLock sync.RWMutex
	cLock   sync.RWMutex
}

func (fs *flowStruct) Done(r interface{}) {
	fs.ctrLock.RLock()
	defer fs.ctrLock.RUnlock()
	for i := 0; i < fs.Ctr; i++ {
		fs.C <- r
	}
}

func (fs *flowStruct) Close() {
	fs.once.Do(func() {
		fs.cLock.Lock()
		defer fs.cLock.Unlock()
		close(fs.C)
	})
}

func (fs *flowStruct) Init() {
	fs.ctrLock.RLock()
	fs.cLock.Lock()
	defer fs.ctrLock.RUnlock()
	defer fs.cLock.Unlock()
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
			flw.funcs[dep].ctrLock.Lock()
			flw.funcs[dep].Ctr++
			flw.funcs[dep].ctrLock.Unlock()
		}
	}
	return flw.do()
}

func (flw *Flow) do() (map[string]interface{}, error) {
	var err error
	var errLock sync.RWMutex
	res := make(map[string]interface{}, len(flw.funcs))

	for _, f := range flw.funcs {
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
				errLock.Lock()
				defer errLock.Unlock()
				err = fnErr
				return
			}
			// exit if error
			errLock.RLock()
			if err != nil {
				errLock.RUnlock()
				return
			}
			errLock.RUnlock()
			fs.Done(r)

		}(name, f)
	}

	// wait for all
	for name, fs := range flw.funcs {
		fs.cLock.RLock()
		res[name] = <-fs.C
		fs.cLock.RUnlock()
	}

	errLock.RLock()
	defer errLock.RUnlock()
	return res, err
}
