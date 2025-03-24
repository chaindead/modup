package tui

import (
	"sync"

	"github.com/spf13/pflag"
)

var workerCnt = pflag.UintP("parallel", "p", 10, "number of concurrent api calls")

type modules struct {
	cnt int
	mu  *sync.RWMutex

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
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return "", false
	}

	last = m.queue[len(m.queue)-1]
	m.queue = m.queue[:len(m.queue)-1]

	return last, true
}

func (m *modules) finished(name string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.scanning {
		if p == name {
			m.scanning = append(m.scanning[:i], m.scanning[i+1:]...)
			break
		}
	}

	return m.scanning
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}
