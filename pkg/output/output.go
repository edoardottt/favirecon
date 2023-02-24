/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

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

// New returns a new Result struct.
func New() Result {
	return Result{
		Map:   map[string]struct{}{},
		Mutex: &sync.RWMutex{},
	}
}

// Printed checks if a string has been previously
// printed.
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

// Format returns a string ready to be printed.
func (f *Found) Format() string {
	return fmt.Sprintf("[%s] [%s] %s", f.Hash, f.Name, f.URL)
}
