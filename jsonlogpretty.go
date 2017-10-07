package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	AnsiReset     = "\x1b[0m"
	AnsiHighlight = "\x1b[1m"
	AnsiBlack     = "\x1b[30m"
	AnsiRed       = "\x1b[31m"
	AnsiGreen     = "\x1b[32m"
	AnsiYellow    = "\x1b[33m"
	AnsiBlue      = "\x1b[34m"
	AnsiPurple    = "\x1b[35m"
	AnsiCyan      = "\x1b[36m"
	AnsiWhite     = "\x1b[37m"
)

type message struct {
	Message string                 `json:"message"`
	Level   string                 `json:"level"`
	Time    time.Time              `json:"time"`
	Data    map[string]interface{} `json:"data"`
	Context map[string]interface{} `json:"context"`
}

func main() {
	ms := make(chan message, 0x40)
	go readMessages(os.Stdin, ms)
	prettyPrintMessages(os.Stdout, ms)
}

func readMessages(r io.Reader, o chan<- message) {
	d := json.NewDecoder(r)
	d.UseNumber()
	defer close(o)

	for true {
		m := message{}
		err := d.Decode(&m)
		if err == nil {
			o <- m
		} else {
			return
		}
	}
}

func prettyPrintMessages(w io.Writer, i <-chan message) {
	bw := bufio.NewWriter(w)
	for true {
		select {
		case m, ok := <-i:
			if ok {
				prettyPrintMessage(bw, m)
				bw.Flush()
			} else {
				bw.WriteString(AnsiReset)
				return
			}
		}
	}
}

func logLevelTag(l string) string {
	switch l {
	case "debug":
		return "DEBUG"
	case "info":
		return "INFO"
	case "warning":
		return "WARN"
	case "error":
		return "ERROR"
	default:
		return "???"
	}
}

func logLevelAnsiEscape(l string) string {
	switch l {
	case "debug":
		return AnsiWhite
	case "info":
		return AnsiReset
	case "warning":
		return AnsiYellow
	case "error":
		return AnsiRed
	default:
		return AnsiCyan
	}
}

func prettyPrintMessage(w *bufio.Writer, m message) {
	w.WriteString(fmt.Sprintf("%s[%5s] %s %s\n",
		logLevelAnsiEscape(m.Level),
		logLevelTag(m.Level),
		m.Time.Format(time.RFC3339),
		m.Message,
	))
	if m.Context != nil {
		prettyPrintData(w, "C", m.Context)
	}
	if m.Data != nil {
		prettyPrintData(w, "D", m.Data)
	}
}

func prettyPrintData(w *bufio.Writer, n string, d map[string]interface{}) {
	if len(d) == 0 {
		w.WriteString(fmt.Sprintf(" %5s\t()\n", n))
	} else {
		first := true
		for k, v := range d {
			if first {
				w.WriteString(fmt.Sprintf(" %5s\t%s: %v\n", n, k, v))
				first = false
			} else {
				w.WriteString(fmt.Sprintf("\t%s: %v\n", k, v))
			}
		}
	}
}
