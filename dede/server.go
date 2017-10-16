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
	"html/template"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/skydive-project/dede/statics"
)

var (
	Log = logging.MustGetLogger("default")

	format = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} ▶ %{level:.6s}%{color:reset} %{message}`)
	router *mux.Router
	lock   sync.RWMutex

	dataDir = "/tmp"
	port    int
)

func createPathFromForm(r *http.Request, filename string) (string, error) {
	path := fmt.Sprintf("%s/%s/%s/%s", dataDir, r.FormValue("sessionID"), r.FormValue("chapterID"), r.FormValue("sectionID"))
	if err := os.MkdirAll(path, 0755); err != nil {
		Log.Errorf("unable to create data dir %s: %s", path, err)
		return "", err
	}
	return filepath.Join(dataDir, r.FormValue("sessionID"), r.FormValue("chapterID"), r.FormValue("sectionID"), filename), nil
}

func idFromForm(r *http.Request, filename string) string {
	return fmt.Sprintf("%s-%s-%s-%s", r.FormValue("sessionID"), r.FormValue("chapterID"), r.FormValue("sectionID"), filename)
}

func asset(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if strings.HasPrefix(upath, "/") {
		upath = strings.TrimPrefix(upath, "/")
	}

	content, err := statics.Asset(upath)
	if err != nil {
		Log.Errorf("unable to find the asset: %s", upath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext := filepath.Ext(upath)
	ct := mime.TypeByExtension(ext)

	w.Header().Set("Content-Type", ct+"; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func index(w http.ResponseWriter, r *http.Request) {
	asset := statics.MustAsset("statics/index.html")

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.New("index").Parse(string(asset)))
	tmpl.Execute(w, nil)
}

func ListenAndServe() {
	addr := fmt.Sprintf(":%d", port)
	Log.Info("DeDe server started on " + addr)
	Log.Fatal(http.ListenAndServe(addr, router))
}

func InitServer(dd string, pp int) {
	logging.SetFormatter(format)

	dataDir = dd
	port = pp

	router = mux.NewRouter()
	router.HandleFunc("/", index)
	router.PathPrefix("/statics").HandlerFunc(asset)

	registerTerminalHandler(router)
	registerFakeMouseHandler(router)
	registerVideoHandler(router)
	registerTextHandler(router)
}
