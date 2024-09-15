package entity

import "time"

type Order struct {
	ID          uint
	UserID      uint      `gorm:"not null;index:idx_order_user_id"`
	User        *User     `gorm:"not null"`
	FetchStatus string    `gorm:"not null;size:250"`
	OrderNumber string    `gorm:"not null;size:250;index:idx_order_order_number,unique"`
	OrderStatus string    `gorm:"size:250"`
	Amount      float64   `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

const (
	FetchStatusWaiting    = "waiting"
	FetchStatusInProgress = "in_progress"
	FetchStatusFinished   = "finished"
	FetchStatusFailure    = "failure"
)
