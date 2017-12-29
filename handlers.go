// Copyright (c) 2017 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"os/user"
	"strconv"
)

const homeTmplPath = "./templates/index.html"

type homeHandler struct {
	tmpl *template.Template
}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmplData, err := assembleTmplData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.tmpl.Execute(w, tmplData)
}

func (h *homeHandler) templateInit() {
	h.tmpl = template.Must(template.ParseFiles(homeTmplPath))
}

func handleKey(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		handleAdd(w, r)
		return
	}
	// PATCH is used to delete a row
	// as with DELETE r.ParseForm()
	// returns nothing.
	if r.Method == "PATCH" {
		handleDelete(w, r)
		return
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	keyID := r.FormValue("keyID")
	if keyID == "" {
		http.Error(w, "provide an ID for the key", http.StatusInternalServerError)
		return
	}
	useGithub := r.FormValue("useGithub")
	pubKey, cipher, err := extractKey(r.FormValue("pubKey"))
	if useGithub == "true" {
		pubKey, cipher, err = retriveFromGH(keyID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = appendKey(keyID, cipher, pubKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	rowID := r.FormValue("rowID")
	if rowID == "" {
		http.Error(w, "specify a row to delete", http.StatusInternalServerError)
		return
	}
	rowIndex, err := strconv.Atoi(rowID)
	checkError := func() {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	checkError()
	user, err := user.Current()
	checkError()
	filePath := user.HomeDir + keyFilePath
	dd, err := ioutil.ReadFile(filePath)
	checkError()
	lines := bytes.Split(dd, []byte("\n"))
	lines = append(lines[:rowIndex], lines[rowIndex+1:]...)
	dd = bytes.Join(lines, []byte("\n"))
	err = ioutil.WriteFile(filePath, dd, 0644)
	checkError()
}
