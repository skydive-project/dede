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
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/mux"
	"github.com/skydive-project/dede/statics"
)

type sessions struct {
	sync.RWMutex
	id        string
	recorders []TerminalRecorder
	recording bool
}

type TerminalHanlder struct {
	sync.RWMutex
	terminalIndexes map[string]*sessions
}

func (t *TerminalHanlder) terminalStartRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	t.RLock()
	s, ok := t.terminalIndexes[id]
	t.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if s.recording {
		w.WriteHeader(http.StatusConflict)
		return
	}

	s.Lock()
	s.recorders = append(s.recorders, NewASCIINemaRecorder(id, ASCIINEMA_DATA_DIR))
	s.recorders = append(s.recorders, NewHistoryRecorder(id, ASCIINEMA_DATA_DIR))
	s.recording = true
	s.Unlock()

	Log.Infof("start recording terminal session %s", id)

	w.WriteHeader(http.StatusOK)
}

func (t *TerminalHanlder) terminalStopRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	t.RLock()
	s, ok := t.terminalIndexes[id]
	t.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	t.RLock()
	recorders := s.recorders
	t.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusPreconditionFailed)
		return
	}

	for _, recorder := range recorders {
		if err := recorder.Write(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	s.Lock()
	s.recorders = s.recorders[:0]
	s.recording = false
	s.Unlock()

	Log.Infof("stop recording terminal session %s", id)

	w.WriteHeader(http.StatusOK)
}

func (t *TerminalHanlder) terminalIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	asset := statics.MustAsset("statics/terminal.html")

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	width := r.FormValue("width")
	if width == "" {
		width = "1200"
	}
	height := r.FormValue("height")
	if height == "" {
		height = "600"
	}

	data := struct {
		ID     string
		Cols   string
		Rows   string
		Width  string
		Height string
		Delay  string
	}{
		ID:     id,
		Cols:   r.FormValue("cols"),
		Rows:   r.FormValue("rows"),
		Width:  width,
		Height: height,
		Delay:  r.FormValue("delay"),
	}

	tmpl := template.Must(template.New("terminal").Parse(string(asset)))
	if err := tmpl.Execute(w, data); err != nil {
		Log.Errorf("Unable to execute terminal template: %s", err)
	}

	t.Lock()
	t.terminalIndexes[id] = &sessions{}
	t.Unlock()
}

func (t *TerminalHanlder) terminalWebsocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	t.RLock()
	s, ok := t.terminalIndexes[id]
	t.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
	if err != nil {
		Log.Errorf("websocket upgrade error: %s", err)
		return
	}
	Log.Infof("websocket new client from: %s", r.RemoteAddr)

	in := make(chan []byte, 50)
	out := make(chan []byte, 50)

	var cols int
	if value := r.FormValue("cols"); value != "" {
		cols, err = strconv.Atoi(value)
		if err != nil {
			Log.Errorf("unable to parse cols value %s: %s", value, err)
		}
	}

	// start a new terminal for this connection
	term := NewTerminal(id, "/bin/bash", TerminalOpts{Cols: cols})
	term.Start(in, out, nil)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer conn.Close()

		for msg := range out {
			s.RLock()
			for _, recorder := range s.recorders {
				recorder.AddOutputEntry(string(msg))
			}
			s.RUnlock()

			err = wsutil.WriteServerMessage(conn, ws.OpText, msg)
			if err != nil {
				Log.Errorf("websocket error while writing message: %s", err)
				break
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				Log.Errorf("websocket error while reading message: %s", err)
				break
			}

			s.RLock()
			for _, recorder := range s.recorders {
				recorder.AddInputEntry(string(msg))
			}
			s.RUnlock()
			in <- msg
		}
		term.close()

		close(out)
	}()

	go func() {
		wg.Wait()
		Log.Infof("websocket client left: %s", r.RemoteAddr)
	}()
}

func NewTerminalHandler(router *mux.Router) *TerminalHanlder {
	t := &TerminalHanlder{
		terminalIndexes: make(map[string]*sessions),
	}

	router.HandleFunc("/terminal/{id}/ws", t.terminalWebsocket)
	router.HandleFunc("/terminal/{id}/start-record", t.terminalStartRecord)
	router.HandleFunc("/terminal/{id}/stop-record", t.terminalStopRecord)
	router.HandleFunc("/terminal/{id}", t.terminalIndex)

	return t
}
