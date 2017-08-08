package dede

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

type ASCIINemaRecordEntry struct {
	delay float64
	data  string
}

type ASCIINemaRecord struct {
	Version   int                    `json:"version"`
	Width     int                    `json:"width"`
	Height    int                    `json:"height"`
	Duration  float64                `json:"duration"`
	Command   string                 `json:"command"`
	Title     string                 `json:"title"`
	Env       map[string]string      `json:"env"`
	Stdout    []ASCIINemaRecordEntry `json:"stdout"`
	lastEntry time.Time
	lock      sync.RWMutex
}

func (a *ASCIINemaRecordEntry) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent([]interface{}{a.delay, a.data}, "", "  ")
}

func (a *ASCIINemaRecord) AddEntry(data string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	now := time.Now()
	delay := float64(now.Sub(a.lastEntry).Nanoseconds()) / float64(time.Second)
	a.Stdout = append(a.Stdout, ASCIINemaRecordEntry{
		delay: delay,
		data:  data,
	})
	a.lastEntry = now
	a.Duration += delay
}

func (a *ASCIINemaRecord) Write(path string) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to serialize asciinema file: %s", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("Unable to write asciinema file: %s", err)
	}

	return nil
}
