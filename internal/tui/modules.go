package tui

import (
	"sync"

	"github.com/spf13/pflag"
)

var workerCnt = pflag.UintP("parallel", "p", 10, "number of concurrent api calls")

type modules struct {
	current int
	cnt     int
	mu      *sync.RWMutex

	queue    []string
	scanning []string
}

func createModules(packages []string) modules {
	return modules{
		queue: packages,
		cnt:   len(packages),
		mu:    &sync.RWMutex{},
	}
}

func (m *modules) next() (last string, ok bool) {
	if len(m.queue) == 0 {
		return "", false
	}

	last = m.queue[len(m.queue)-1]
	m.queue = m.queue[:len(m.queue)-1]

	return last, true
}

func (m modules) isFinished() bool {
	return m.current == m.cnt
}

func (m modules) progressFloat() float64 {
	return float64(m.current) / float64(m.cnt)
}

func remove(slice []string, s string) []string {
	for i, p := range slice {
		if p == s {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}

	return slice
}
