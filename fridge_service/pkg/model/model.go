package model

import (
	"time"
)

type Fridge struct {
	Id        string    `json:"id" gorm:"id"`
	Name      string    `json:"name" gorm:"name"`
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`
}

type Product struct {
	Id        string    `json:"id" gorm:"id"`
	Name      string    `json:"name" gorm:"name"`
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`
}

type ProductInFridge struct {
	FridgeId       string    `json:"fridge_id" gorm:"fridge_id"`
	ProductId      string    `json:"product_id" gorm:"product_id"`
	Quantity       float64   `json:"quantity" gorm:"quantity"`
	ProductionDate *time.Time `json:"production_date" gorm:"production_date"`
	ExpirationDate *time.Time `json:"expiration_date" gorm:"expiration_date"`
	AddedAt        time.Time `json:"id" gorm:"id"`
}
