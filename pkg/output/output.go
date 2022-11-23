package output

import (
	"fmt"
	"sync"
)

type Found struct {
	URL  string
	Hash string
	Name string
}

type Result struct {
	Map   map[string]struct{}
	Mutex *sync.RWMutex
}

func New() Result {
	return Result{
		Map:   map[string]struct{}{},
		Mutex: &sync.RWMutex{},
	}
}

func (o *Result) Printed(found string) bool {
	o.Mutex.RLock()
	if _, ok := o.Map[found]; !ok {
		o.Mutex.RUnlock()
		o.Mutex.Lock()
		o.Map[found] = struct{}{}
		o.Mutex.Unlock()

		return false
	} else {
		o.Mutex.RUnlock()
	}

	return true
}

func (f *Found) Format() string {
	return fmt.Sprintf("[%s] [%s] %s", f.Hash, f.Name, f.URL)
}
