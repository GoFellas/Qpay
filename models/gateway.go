package models

import (
	"gorm.io/gorm"
)

type Gateway struct {
	gorm.Model
	GatewayID     uint `gorm:"primaryKey"`
	UserID        uint
	ComID         uint
	BankAccountID uint
	User          User        `gorm:"foreignKey:UserID"`
	Commission    Commission  `gorm:"foreignKey:ComID"`
	BankAccount   BankAccount `gorm:"foreignKey:BankAccountID"`
	Transactions  []Transaction
	Name          string `gorm:"type:varchar(50)"`
	Logo          string `gorm:"type:varchar(50)"`
	Route         string `gorm:"type:varchar(50)"`
	Status        uint8
	Type          uint8
}
