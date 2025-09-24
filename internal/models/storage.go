package models

import (
	"time"

	"github.com/oegegr/gophermart/internal/models/api"
)

type User struct {
	Login        string
	PasswordHash []byte
}

type Order struct {
	Number       string 
	Accrual      float32
	Status       string
	UploadedAt   time.Time
	Login        string
}

func (o *Order) ToApi(order Order) api.Order {
	status := statusMap[order.Status]
		return api.Order{
			Number: &order.Number,
			Status: &status,
			Accrual: &o.Accrual,
			UploadedAt: &o.UploadedAt,
		}
}

type Balance struct {
	Current float32
	Withdraw float32
}

var statusMap =  map[string]api.OrderStatus{
    "INVALID":   api.INVALID,
    "PROCESSED": api.PROCESSED,
    "PROCESSING": api.PROCESSING,
    "REGISTERED": api.REGISTERED,
}



