package db

import (
	"database/sql"
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

	// Calculate account balance in SQL view
	err = d.Get(&balance, "select balance from calculated_balance_view")
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if data.State == "lost" && balance-data.Amount < 0 {
		err = fmt.Errorf("attempt to set negative balance: %f actual balance: %f", balance-data.Amount, balance)
		return
	}

	_, err = d.NamedExec(`insert into balance_history(operation, amount, tid) values (:operation, :amount, :tid)`, data)
	if err != nil {
		if (err.(*pq.Error)).Code.Name() == "unique_violation" {
			err = fmt.Errorf("transaction with id:%s has been processed already", data.TransactionId)
		}
	}
	return
}

func (d *DB) PostProcess() {
	d.Lock()
	defer d.Unlock()
	log.Println("Post processing started")
	var (
		tids      []string
		err       error
		cancelled []string
		odds      []interface{}
	)
	tx, err := d.Beginx()
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err != nil {
			terr := tx.Rollback()
			if terr != nil {
				log.Println("SERIOUS DATABASE PROBLEM:" + terr.Error())
			}
			log.Println("Post processing failed")
			log.Println(err)
		} else {
			terr := tx.Commit()
			if terr != nil {
				log.Println("SERIOUS DATABASE PROBLEM:" + terr.Error())
			}
			log.Printf("%d records with transaction IDs (%s) cancelled", len(cancelled), strings.Join(cancelled, ","))
			log.Println("Post processing completed")
		}
	}()

	err = tx.Select(&tids,
		`select tid 
				from balance_history 
				where deleted = false 
				order by date_time desc limit 19`) // 19 might be moved to the configuration
	if err != nil {
		return
	}

	if len(tids) == 0 {
		log.Println("nothing to cancel")
		return
	}

	var placeholders []string
	n := 1
	for i, v := range tids {
		if i%2 == 0 {
			odds = append(odds, v)
			cancelled = append(cancelled, v)
			// Prevention of SQL injections. Building string '$1,$2,$3,$4...'.
			placeholders = append(placeholders, fmt.Sprintf("$%d", n))
			n++
		}
	}

	updq := fmt.Sprintf("update balance_history set deleted = true where tid in (%s)", strings.Join(placeholders, ","))
	_, err = tx.Exec(updq, odds...)

	var balance float64
	err = tx.Get(&balance, "select balance from calculated_balance_view")
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if balance < 0 {
		err = fmt.Errorf("attempt to set negative during post processing. Balance: %f ", balance)
	}
}
