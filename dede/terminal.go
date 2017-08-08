package dede

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/kr/pty"
)

type TerminalOpt struct {
}

type Terminal struct {
	sync.RWMutex
	ID     string
	cmd    string
	pty    *os.File
	record ASCIINemaRecord
}

func NewTerminal(id string, cmd string) *Terminal {
	t := &Terminal{
		ID:  id,
		cmd: cmd,
	}
	t.record.Env = make(map[string]string)
	t.record.lastEntry = time.Now()

	return t
}

func (t *Terminal) Start(in chan []byte, out chan []byte) {
	os.Setenv("COLUMNS", "-1")
	p, err := pty.Start(exec.Command(t.cmd))
	if err != nil {
		Log.Fatalf("Failed to start: %s\n", err)
	}

	t.Lock()
	t.pty = p
	t.Unlock()

	// pty reading
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := p.Read(buf)
			data := buf[:n]

			t.record.AddEntry(string(data))

			if err != nil {
				Log.Errorf("Failed to read pty: %s", err)
				return
			}
			out <- data
		}
	}()

	// pty writing
	go func() {
		for b := range in {
			if _, err := p.Write(b); err != nil {
				Log.Errorf("Failed to write pty: %s", err)
				return
			}
		}
	}()
}

func (t *Terminal) close() {
	if err := t.pty.Close(); err != nil {
		Log.Errorf("Failed to stop: %s\n", err)
	}

	if err := t.record.Write(fmt.Sprintf("%s/%s.json", ASCIINEMA_DATA_DIR, t.ID)); err != nil {
		Log.Error(err)
	}
}
