package entity

import "time"

type Withdrawal struct {
	ID          uint
	UserID      uint      `gorm:"not null;index:idx_withdrawal_user_id"`
	User        *User     `gorm:"not null"`
	OrderNumber string    `gorm:"not null;size:250;index:idx_withdrawal_user_id_number,unique"`
	Amount      float64   `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
}
