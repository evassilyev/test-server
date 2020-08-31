package main

import (
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type app struct {
	conf           Configuration
	db             *sqlx.DB
	allowedSources map[string]bool
}

func (a *app) initDB() {
	a.db = sqlx.MustOpen("postgres", a.conf.Pgdb)
	a.db.SetMaxIdleConns(2)
	a.db.SetMaxOpenConns(0)
	err := a.db.Ping()
	if err != nil {
		log.Println("can't initialize database connection")
		panic(err)
	}
}

func NewApp(c Configuration) *app {
	a := new(app)
	a.conf = c
	a.allowedSources = map[string]bool{
		"client":  true,
		"server":  true,
		"payment": true,
	}
	a.initDB()
	return a
}

func (a *app) Run() error {
	router := mux.NewRouter()

	router.HandleFunc(a.conf.Endpoint, a.requestHandler).Methods(http.MethodPost)

	log.Println("Started at port " + a.conf.Port)
	err := http.ListenAndServe(":"+a.conf.Port, router)
	return err
}

type request struct {
	State         string  `json:"state" db:"operation"`
	Amount        float32 `json:"amount" db:"amount"`
	TransactionId string  `json:"transactionId" db:"tid"`
}

// TODO decompose
func (a *app) requestHandler(w http.ResponseWriter, r *http.Request) {
	source := r.Header.Get("Source-Type")
	if _, ok := a.allowedSources[source]; !ok {
		http.Error(w, fmt.Sprintf("source type %s is not allowed", source), http.StatusForbidden)
		return
	}
	if r.Proto != "HTTP/1.1" {
		http.Error(w, fmt.Sprintf("protocol %s is not allowed", r.Proto), http.StatusForbidden)
		return
	}
	log.Println(r.Proto)
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		http.Error(w, fmt.Sprintf("wrong content type: %s", ct), http.StatusForbidden)
		return
	}
	log.Println(ct)

	var data request
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(data)
	if data.State != "win" || data.State != "lost" {
		http.Error(w, fmt.Sprintf("wrong operation: %s", data.State), http.StatusForbidden)
		return
	}
	_, err = a.db.NamedExec(`insert into balance_history(operation, amount, tid) values (:operation, :amount, :tid)`, data)
	if err != nil {
		if (err.(*pq.Error)).Code.Name() == "unique_violation" {
			http.Error(w, fmt.Sprintf("transaction with id:%s has been processed already", data.TransactionId), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var balance float64
	err = a.db.Get(&balance, "select balance from calculated_balance_view")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(balance)
	w.Write([]byte(fmt.Sprintf("Data stored. Calculated balance: %f", balance)))
}

func (a *app) postProcessor() {
	for {
		time.Sleep(10 * time.Minute)
		var (
			tids []string
			odds []string
		)
		// TODO add transactions
		err := a.db.Select(&tids,
			`select tid 
				from balance_history 
				where deleted = false 
				order by date_time desc limit 19`)
		if err != nil {
			log.Println(err)
		}
		for i, v := range tids {
			if i%2 == 0 {
				odds = append(odds, v)
			}
		}
		fmt.Println(len(odds))
		updq := fmt.Sprintf("update balance_history set deleted = true where tid in (%s)", strings.Join(odds, ","))
		_, err = a.db.Exec(updq)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(fmt.Sprintf("%n records with transaction IDs (%s) cancelled", len(odds), strings.Join(odds, ",")))
		}
	}
}
