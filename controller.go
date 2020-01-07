package main

import (
	"encoding/json"
	"fmt"
	"github.com/dutchcoders/go-clamd"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
	"time"
)

type ClamAV struct {
	clamUrl string
}

func (controller ClamAV) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Clam AV Controller!\n")
}

func (controller ClamAV) Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	clam := clamd.NewClamd(controller.clamUrl)
	err := clam.Ping()
	if err == nil {
		fmt.Fprint(w, "ok")
	} else {
		http.Error(w, "Service Unavailable", 503)
	}
}

func (controller ClamAV) Version(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	clam := clamd.NewClamd(controller.clamUrl)
	v, _ := clam.Version()
	version := <-v
	response, _ := json.Marshal(Response{AV_Version: version.Raw})
	fmt.Fprint(w, string(response))
}

func (controller ClamAV) Scan(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")[0]
	fmt.Printf("File name %s\n", name)

	defer r.Body.Close()
	if name != "" {
		clam := clamd.NewClamd(controller.clamUrl)
		chanFoo := make(chan bool)

		ch, _ := clam.ScanStream(file, chanFoo)
		x := <-ch
		var detected bool
		if x.Status == "FOUND" {
			detected = true
		}
		rsp, _ := json.Marshal(Response{File: name, Detected: detected, Malware: x.Description})
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, string(rsp))
	} else {
		fmt.Fprint(w, "Empty file as an input, you must send file")
	}
}

func (controller ClamAV) Stats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clam := clamd.NewClamd(controller.clamUrl)
	stats, _ := clam.Stats()
	response, _ := json.Marshal(stats)
	fmt.Fprint(w, string(response))
}

type Response struct {
	File      *string `json:"file,omitempty"`
	Detected  *bool   `json:"detected,omitempty"`
	Malware   *string `json:"malware,omitempty"`
	AVVersion *string `json:"av_version,omitempty"`
	Metadata  *string `json:"metadata,omitempty"`
}
