package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/evassilyev/test-server/db"
	"github.com/evassilyev/test-server/models"
	"github.com/gorilla/mux"
)

type app struct {
	conf           Configuration
	db             *db.DB
	allowedSources map[string]bool
}

func NewApp(c Configuration) *app {
	a := new(app)
	a.conf = c
	a.allowedSources = map[string]bool{
		"client":  true,
		"server":  true,
		"payment": true,
	}
	a.db = db.NewDB(a.conf.Pgdb)
	return a
}

func (a *app) Run() error {
	router := mux.NewRouter()

	router.HandleFunc(a.conf.Endpoint, a.requestHandler).Methods(http.MethodPost)

	go func() {
		log.Println(fmt.Sprintf("Post processor started at %s with interval %d minutes", time.Now().Format(time.RFC822), a.conf.Ppinterval))
		for {
			time.Sleep(time.Duration(a.conf.Ppinterval) * time.Minute)
			a.db.PostProcess()
		}
	}()

	log.Println("Started at port " + a.conf.Port)
	return http.ListenAndServe(":"+a.conf.Port, router)
}

func (a *app) processRequest(r *http.Request) (data models.Data, err error) {
	source := r.Header.Get("Source-Type")
	if _, ok := a.allowedSources[source]; !ok {
		err = errors.New(fmt.Sprintf("source type %s is not allowed", source))
		return
	}
	if r.Proto != "HTTP/1.1" {
		err = errors.New(fmt.Sprintf("protocol %s is not allowed", r.Proto))
		return
	}
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		err = errors.New(fmt.Sprintf("wrong content type: %s", ct))
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err = decoder.Decode(&data)
	if err != nil {
		return
	}

	log.Println("Extracted data: ", data)
	if data.State != "win" && data.State != "lost" {
		err = errors.New(fmt.Sprintf("wrong operation: %s", data.State))
		return
	}
	if data.Amount < 0 {
		err = errors.New(fmt.Sprintf("negative amount: %f", data.Amount))
		return
	}
	return
}

func (a *app) requestHandler(w http.ResponseWriter, r *http.Request) {
	data, err := a.processRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = a.db.StoreData(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
