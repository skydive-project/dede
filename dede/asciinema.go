/*
 * Copyright (C) 2017 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package dede

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sync"
	"time"
)

type ASCIINemaRecorderEntry struct {
	delay float64
	data  string
}

type ASCIINemaRecorder struct {
	Version   int                      `json:"version"`
	Width     int                      `json:"width"`
	Height    int                      `json:"height"`
	Duration  float64                  `json:"duration"`
	Command   string                   `json:"command"`
	Title     string                   `json:"title"`
	Env       map[string]string        `json:"env"`
	Stdout    []ASCIINemaRecorderEntry `json:"stdout"`
	lastEntry time.Time
	lock      sync.RWMutex
	path      string
	id        string
}

func (a *ASCIINemaRecorderEntry) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent([]interface{}{a.delay, a.data}, "", "  ")
}

func (a *ASCIINemaRecorder) AddEntry(data string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	now := time.Now()
	delay := float64(now.Sub(a.lastEntry).Nanoseconds()) / float64(time.Second)
	a.Stdout = append(a.Stdout, ASCIINemaRecorderEntry{
		delay: delay,
		data:  data,
	})
	a.lastEntry = now
	a.Duration += delay
}

func (a *ASCIINemaRecorder) AddInputEntry(data string) {
	a.AddEntry(data)
}

func (a *ASCIINemaRecorder) AddOutputEntry(data string) {
	a.AddEntry(data)
}

func (a *ASCIINemaRecorder) Write() error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to serialize asciinema file: %s", err)
	}

	p := path.Join(a.path, fmt.Sprintf("asciinema-%s.json", a.id))
	if err := ioutil.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("Unable to write asciinema file: %s", err)
	}

	return nil
}

func NewASCIINemaRecorder(id, path string) *ASCIINemaRecorder {
	return &ASCIINemaRecorder{
		Env:       make(map[string]string),
		lastEntry: time.Now(),
		id:        id,
		path:      path,
	}
}
