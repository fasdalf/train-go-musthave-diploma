package entity

type User struct {
	ID                uint
	Login             string  `gorm:"uniqueIndex;not null;size:250"`
	PassHash          string  `gorm:"not null;size:250"`
	Amount            float64 `gorm:"not null"`
	WithdrawalsAmount float64 `gorm:"not null"`
}
