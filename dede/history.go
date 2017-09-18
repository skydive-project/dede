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
	"strings"
	"sync"
)

type historyRecorderEntry struct {
	Input  string
	Output string
}

type historyRecorder struct {
	Entries       []historyRecorderEntry
	lock          sync.RWMutex
	filename      string
	current       historyRecorderEntry
	prevData      string
	prevDirection string
}

func (h *historyRecorder) addInputEntry(data string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.prevDirection == "output" {
		h.Entries = append(h.Entries, h.current)

		// generate a new entry
		h.current = historyRecorderEntry{}
	}

	t := strings.TrimSuffix(data, "\r")
	if t != "" {
		h.current.Input += t
	}
	h.prevData = data
	h.prevDirection = "input"
}

func (h *historyRecorder) addOutputEntry(data string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// do not register echo
	if data == h.prevData {
		return
	}

	h.current.Output += data
	h.prevDirection = "output"
}

func (h *historyRecorder) write() error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if h.current.Input != "" {
		h.Entries = append(h.Entries, h.current)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to serialize history file: %s", err)
	}

	if err := ioutil.WriteFile(h.filename, data, 0644); err != nil {
		return fmt.Errorf("Unable to write asciinema file: %s", err)
	}

	return nil
}

func newHistoryRecorder(filename string) *historyRecorder {
	return &historyRecorder{
		filename: filename,
	}
}
