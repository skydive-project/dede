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

$('#dede-terminal-record').change(function() {
  DedeTerminal.toggleRecord();
});

_DedeTerminal = function(endpoint, id, cols, rows, delay, initCmd) {
  var self = this;

  this.endpoint = endpoint;
  this.id = id;
  this.cols = parseInt(cols) || 80;
  this.rows = parseInt(rows) || 40;
  this.delay = parseInt(delay) || 100;
  this.recording = false;

  var container = document.getElementById('dede-terminal');
  var term = new Terminal({
    geometry: [this.cols, this.rows],
    //debug: true
  });
  term.open(container);
  term.fit();

  this.term = term;
  this.onMessageCallback = null;

  var websocket;
  function wsConnect() {
    var protocol = "ws://";
    if (location.protocol === "https:") {
      protocol = "wss://";
    }

    websocket = new WebSocket(protocol + endpoint + "/terminal/" + id + "/ws?cols=" + cols);
    websocket.onopen = function(evt) {
      term.attach(websocket);

      if (initCmd) {
        self.typeCmd(initCmd);
      }
    };
    websocket.addEventListener('message', function(msg) {
      if (self.onMessageCallback) self.onMessageCallback(msg);
    });
    websocket.onclose = wsDisconnected;
    websocket.onerror = wsOnError;
  }
  wsConnect();

  function wsOnError() {
    $.notify("Writing error", "error");
  }

  function wsDisconnected(evt) {
    $.notify("Connection error", "error");

    setTimeout(wsConnect, 5000);
  }
};

_DedeTerminal.prototype.type = function(str, callback) {
  var self = this;

  var el = str.split('');
  var i = 0;

  var fnc = function() {
    self.term.send(el[i]);
    if (++i < el.length) {
      setTimeout(fnc, self.delay);
    } else if (callback) callback();
  };
  fnc();
};

_DedeTerminal.prototype.typeCmd = function(str, callback) {
  var self = this;

  this.type(str, function() {
    self.term.send("\n");
    if (callback) callback();
  });
};

_DedeTerminal.prototype.typeCmdWait = function(str, regex, callback) {
  var self = this;

  var re = new RegExp(regex, "m");
  this.onMessageCallback = function(msg) {
    if (re.test(msg.data)) {
      self.onMessageCallback = null;
      if (callback) callback();
    }
  };

  this.typeCmd(str);
};

_DedeTerminal.prototype.startRecord = function(sessionID, chapterID, sectionID) {
  $.ajax({
     url : location.protocol + '//' + this.endpoint + '/terminal/' + this.id + '/start-record?sessionID=' + sessionID +
           "&chapterID=" + chapterID + "&sectionID=" + sectionID
  });
  this.recording = true;
};

_DedeTerminal.prototype.stopRecord = function() {
  $.ajax({
     url : location.protocol + '//' + this.endpoint + '/terminal/' + this.id + '/stop-record'
  });
  this.recording = false;
};

_DedeTerminal.prototype.toggleRecord = function() {
  if (this.recording)
    this.stopRecord();
  else
    this.startRecord();
};
