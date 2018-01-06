// Copyright (c) 2017 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package main

import (
	"context"
	"fmt"
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
	hasher := createHasher(os.Getenv("HTTP_PASSWORD"))
	infoLogger := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	//
	// Routes
	//
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/key", middle.Auth(hasher, middle.Log(infoLogger, http.HandlerFunc(handleKey))))
	http.Handle("/", middle.Log(infoLogger, middle.Auth(hasher, homeHndlr)))
	addr := fmt.Sprintf("%s:%s", os.Getenv("LISTENING_ADDRESS"), os.Getenv("LISTENING_PORT"))
	httpServer := &http.Server{
		Addr: addr,
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
