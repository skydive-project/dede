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
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skydive-project/skydive/common"
)

type textHandler struct {
}

func (v *textHandler) addText(w http.ResponseWriter, r *http.Request) {
	tp, err := createPathFromForm(r, "text.json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	text := struct {
		Type string
		Text string
	}{}

	if err = common.JSONDecode(r.Body, &text); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.MarshalIndent(text, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := ioutil.WriteFile(tp, data, 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	Log.Infof("Text recorded %s", tp)
	w.WriteHeader(http.StatusOK)
}

func registerTextHandler(router *mux.Router) *textHandler {
	t := &textHandler{}

	router.HandleFunc("/text", t.addText)
	return t
}
