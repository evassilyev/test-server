package models

type Data struct {
	State         string  `json:"state" db:"operation"`
	Amount        float64 `json:"amount" db:"amount"`
	TransactionId string  `json:"transactionId" db:"tid"`
}
