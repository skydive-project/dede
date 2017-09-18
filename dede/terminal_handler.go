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
	recorders []terminalRecorder
	recording bool
}

type terminalHanlder struct {
	sync.RWMutex
	terminalIndexes map[string]*sessions
}

func (t *terminalHanlder) terminalStartRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := idFromVars(vars, "term")

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
	ap, err := pathFromVars(vars, "asciinema.json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.recorders = append(s.recorders, newAsciinemaRecorder(ap))

	tp, err := pathFromVars(vars, "terminal.json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.recorders = append(s.recorders, newHistoryRecorder(tp))
	s.recording = true
	s.Unlock()

	Log.Infof("start recording terminal session %s", id)
	w.WriteHeader(http.StatusOK)
}

func (t *terminalHanlder) terminalStopRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := idFromVars(vars, "term")

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
		if err := recorder.write(); err != nil {
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

func (t *terminalHanlder) terminalIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := idFromVars(vars, "term")

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
		ID       string
		Cols     string
		Rows     string
		Width    string
		Height   string
		Delay    string
		Controls string
	}{
		ID:       id,
		Cols:     r.FormValue("cols"),
		Rows:     r.FormValue("rows"),
		Width:    width,
		Height:   height,
		Delay:    r.FormValue("delay"),
		Controls: r.FormValue("controls"),
	}

	tmpl := template.Must(template.New("terminal").Parse(string(asset)))
	if err := tmpl.Execute(w, data); err != nil {
		Log.Errorf("Unable to execute terminal template: %s", err)
	}

	t.Lock()
	t.terminalIndexes[id] = &sessions{}
	t.Unlock()
}

func (t *terminalHanlder) terminalWebsocket(w http.ResponseWriter, r *http.Request) {
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
	term := newTerminal("/bin/bash", terminalOpts{cols: cols})
	term.start(in, out, nil)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer conn.Close()

		for msg := range out {
			s.RLock()
			for _, recorder := range s.recorders {
				recorder.addOutputEntry(string(msg))
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
				recorder.addInputEntry(string(msg))
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

func registerTerminalHandler(router *mux.Router) *terminalHanlder {
	t := &terminalHanlder{
		terminalIndexes: make(map[string]*sessions),
	}

	router.HandleFunc(baseURL+"/terminal/ws", t.terminalWebsocket)
	router.HandleFunc(baseURL+"/terminal/start-record", t.terminalStartRecord)
	router.HandleFunc(baseURL+"/terminal/stop-record", t.terminalStopRecord)
	router.HandleFunc(baseURL+"/terminal", t.terminalIndex)

	return t
}
