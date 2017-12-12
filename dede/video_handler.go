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
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type videoHanlder struct {
	sync.RWMutex
	recorders map[string]*videoRecorder
}

func (v *videoHanlder) startRecord(w http.ResponseWriter, r *http.Request) {
	vp, err := createPathFromForm(r, "video.mp4")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := idFromForm(r, "video.mp4")

	v.RLock()
	ok := v.recorders[id] == nil
	v.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusConflict)
		return
	}

	recorder := newVideoRecorder(vp, 1900, 1080, 10)
	if err := recorder.start(); err != nil {
		Log.Errorf("error while starting video record: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	v.Lock()
	v.recorders[id] = recorder
	v.Unlock()

	Log.Infof("start video recording %s", vp)
	w.WriteHeader(http.StatusOK)
}

func (v *videoHanlder) stopRecord(w http.ResponseWriter, r *http.Request) {
	id := idFromForm(r, "video.mp4")

	v.RLock()
	recorder := v.recorders[id]
	v.RUnlock()
	if recorder == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	recorder.stop()

	v.Lock()
	delete(v.recorders, id)
	v.Unlock()

	Log.Infof("stop video recording")
	w.WriteHeader(http.StatusOK)
}

func RegisterVideoHandler(prefix string, router *mux.Router) *videoHanlder {
	t := &videoHanlder{
		recorders: make(map[string]*videoRecorder),
	}

	router.HandleFunc(prefix+"/video/start-record", t.startRecord)
	router.HandleFunc(prefix+"/video/stop-record", t.stopRecord)

	return t
}
