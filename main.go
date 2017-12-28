package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/eraclitux/middle"
)

func main() {
	homeHndlr := &homeHandler{}
	homeHndlr.templateInit()
	hasher := createHasher("secret")
	//
	// Routes
	//
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/key", middle.MustAuth(hasher, http.HandlerFunc(handleKey)))
	http.Handle("/", middle.MustAuth(hasher, homeHndlr))
	// For security reasons this server,
	// will only listen on localhost.
	// FIXME parametrize port
	httpServer := &http.Server{
		Addr: "127.0.0.1:8080",
	}
	sigCh := make(chan os.Signal, 1)
	shutDownCh := make(chan struct{})
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		httpServer.Shutdown(context.Background())
		shutDownCh <- struct{}{}
	}()
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Println(err)
		// In case error has not been
		// caused by signals (es. address already in use)
		// avoid to block on shutDownCh.
		close(sigCh)
	}
	<-shutDownCh
}
