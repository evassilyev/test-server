package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/evassilyev/test-server/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"log"
	"strings"
	"sync"
)

type DB struct {
	*sqlx.DB
	sync.Mutex
}

// TODO add proper concurrency

func NewDB(url string) *DB {
	db := sqlx.MustOpen("postgres", url)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(0)
	err := db.Ping()
	if err != nil {
		log.Println("can't initialize database connection")
		panic(err)
	}
	return &DB{DB: db}
}

func (d *DB) StoreData(data models.Data) (err error) {
	var balance float64
	d.Lock()
	defer d.Unlock()

	// Calculate account balance in SQL
	err = d.Get(&balance, "select balance from calculated_balance_view")
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if data.State == "lost" && balance-data.Amount < 0 {
		err = errors.New(fmt.Sprintf("attempt to set negative balance: %f actual balance: %f", balance-data.Amount, balance))
		return
	}

	_, err = d.NamedExec(`insert into balance_history(operation, amount, tid) values (:operation, :amount, :tid)`, data)
	if err != nil {
		if (err.(*pq.Error)).Code.Name() == "unique_violation" {
			err = errors.New(fmt.Sprintf("transaction with id:%s has been processed already", data.TransactionId))
		}
	}
	return
}

func (d *DB) PostProcess() {
	d.Lock()
	defer d.Unlock()
	log.Println("Post processing started")
	var (
		tids []string
		odds []string
		err  error
	)
	defer func() {
		if err != nil {
			log.Println(err)
			log.Println("Post processing failed")
		} else {
			log.Println(fmt.Sprintf("%d records with transaction IDs (%s) cancelled", len(odds), strings.Join(odds, ",")))
			log.Println("Post processing completed")
		}
	}()

	err = d.Select(&tids,
		`select tid 
				from balance_history 
				where deleted = false 
				order by date_time desc limit 19`)
	if err != nil {
		return
	}

	for i, v := range tids {
		if i%2 == 0 {
			odds = append(odds, v)
		}
	}

	updq := fmt.Sprintf("update balance_history set deleted = true where tid in ('%s')", strings.Join(odds, "','"))
	_, err = d.Exec(updq)
}
