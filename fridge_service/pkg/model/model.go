package model

import (
	"time"
)

type Fridge struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID             string     `json:"id" gorm:"type:uuid;primaryKey"`
	Name           string     `json:"name" gorm:"unique"J`
	ExpirationTime *time.Time `json:"expiration_time"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ProductInFridge struct {
	FridgeID       string     `json:"fridge_id" gorm:"type:uuid;primaryKey"`
	ProductID      string     `json:"product_id" gorm:"type:uuid;primaryKey"`
	Quantity       float64    `json:"quantity"`
	ProductionDate *time.Time `json:"production_date"`
	ExpirationDate *time.Time `json:"expiration_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type FridgeUser struct {
	FridgeID  string    `json:"fridge_id" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;primaryKey"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID string `json:"id"`
}
