package dede

import (
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/skydive-project/dede/statics"
)

const (
	ASCIINEMA_DATA_DIR = "/tmp"
)

var (
	Log = logging.MustGetLogger("default")

	format = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} ▶ %{level:.6s}%{color:reset} %{message}`)
	router *mux.Router
	lock   sync.RWMutex
)

func asset(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if strings.HasPrefix(upath, "/") {
		upath = strings.TrimPrefix(upath, "/")
	}

	content, err := statics.Asset(upath)
	if err != nil {
		Log.Errorf("Unable to find the asset: %s", upath)
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
	asset := statics.MustAsset("statics/server.html")

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.New("index").Parse(string(asset)))
	tmpl.Execute(w, nil)
}

func terminalHanlder(w http.ResponseWriter, r *http.Request) {
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
}

func terminalWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
	if err != nil {
		Log.Errorf("Websocket error: %s", err.Error())
		return
	}
	Log.Infof("Websocket new client from: %s", r.RemoteAddr)

	in := make(chan []byte, 50)
	out := make(chan []byte, 50)

	// start a new terminal for this connection
	term := NewTerminal(id, "/bin/bash")
	term.Start(in, out)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer conn.Close()

		for msg := range out {
			err = wsutil.WriteServerMessage(conn, ws.OpText, msg)
			if err != nil {
				Log.Errorf("Websocket error while writing message: %s", err)
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
				Log.Errorf("Websocket error while reading message: %s", err)
				break
			}
			in <- msg
		}
		term.close()

		close(out)
	}()

	go func() {
		wg.Wait()
		Log.Infof("Websocket client left: %s", r.RemoteAddr)
	}()
}

func ListenAndServe() {
	Log.Info("Dede server started")
	Log.Fatal(http.ListenAndServe(":12345", router))
}

func InitServer() {
	logging.SetFormatter(format)

	router = mux.NewRouter()
	router.HandleFunc("/", index)
	router.PathPrefix("/statics").HandlerFunc(asset)
	router.HandleFunc("/terminal/{id}/ws", terminalWebsocketHandler)
	router.HandleFunc("/terminal/{id}", terminalHanlder)
}
