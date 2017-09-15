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
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/kr/pty"
)

type TerminalOpts struct {
	Cols int
}

type Terminal struct {
	sync.RWMutex
	Cmd  string
	Opts TerminalOpts
	pty  *os.File
}

// NewTerminal returns a new Terminal for the given id and command
func NewTerminal(cmd string, opts ...TerminalOpts) *Terminal {
	t := &Terminal{
		Cmd: cmd,
	}
	if len(opts) > 0 {
		t.Opts = opts[0]
	}

	return t
}

// Start starts reading the in chan, and writing to the out chan. the err chan
// is used to report errors.
func (t *Terminal) Start(in chan []byte, out chan []byte, err chan error) {
	if t.Opts.Cols != 0 {
		os.Setenv("COLUMNS", fmt.Sprintf("%d", t.Opts.Cols))
	}

	p, e := pty.Start(exec.Command(t.Cmd))
	if e != nil {
		err <- fmt.Errorf("failed to start: %s", e)
		return
	}

	t.Lock()
	t.pty = p
	t.Unlock()

	// pty reading
	go func() {
		for {
			buf := make([]byte, 1024)
			n, e := p.Read(buf)
			data := buf[:n]

			if e != nil {
				err <- fmt.Errorf("failed to start: %s", e)
				return
			}
			out <- data
		}
	}()

	// pty writing
	go func() {
		for b := range in {
			if _, e := p.Write(b); e != nil {
				err <- fmt.Errorf("failed to start: %s", e)
				return
			}
		}
	}()
}

// Close stops the current command and closes the Terminal
func (t *Terminal) close() {
	if err := t.pty.Close(); err != nil {
		Log.Errorf("failed to stop: %s", err)
	}
}
