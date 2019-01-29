// +build !appengine

package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	listenFlag = flag.String("listen", "127.0.0.1:8080", "ip:port to listen on")
)

func addGroupPlayHistory(r *http.Request, gp groupPlay) {
	historyMu.Lock()
	defer historyMu.Unlock()
	memHistory = append(memHistory, gp)
}

func loadHistory(r *http.Request) (history, error) {
	historyMu.Lock()
	defer historyMu.Unlock()
	return append(history(nil), memHistory...), nil
}

func main() {
	flag.Parse()
	log.Fatal(http.ListenAndServe(*listenFlag, nil))
}
