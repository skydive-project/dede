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
	"sync"
	"time"
)

type asciinemaRecorderEntry struct {
	delay float64
	data  string
}

type asciinemaRecorder struct {
	Version   int                      `json:"version"`
	Width     int                      `json:"width"`
	Height    int                      `json:"height"`
	Duration  float64                  `json:"duration"`
	Command   string                   `json:"command"`
	Title     string                   `json:"title"`
	Env       map[string]string        `json:"env"`
	Stdout    []asciinemaRecorderEntry `json:"stdout"`
	lastEntry time.Time
	lock      sync.RWMutex
	filename  string
}

func (a *asciinemaRecorderEntry) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent([]interface{}{a.delay, a.data}, "", "  ")
}

func (a *asciinemaRecorder) addEntry(data string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	now := time.Now()
	delay := float64(now.Sub(a.lastEntry).Nanoseconds()) / float64(time.Second)
	a.Stdout = append(a.Stdout, asciinemaRecorderEntry{
		delay: delay,
		data:  data,
	})
	a.lastEntry = now
	a.Duration += delay
}

func (a *asciinemaRecorder) addInputEntry(data string) {
	a.addEntry(data)
}

func (a *asciinemaRecorder) addOutputEntry(data string) {
	a.addEntry(data)
}

func (a *asciinemaRecorder) write() error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to serialize asciinema file: %s", err)
	}

	if err := ioutil.WriteFile(a.filename, data, 0644); err != nil {
		return fmt.Errorf("Unable to write asciinema file: %s", err)
	}

	return nil
}

func newAsciinemaRecorder(filemane string) *asciinemaRecorder {
	return &asciinemaRecorder{
		Env:       make(map[string]string),
		lastEntry: time.Now(),
		filename:  filemane,
	}
}
