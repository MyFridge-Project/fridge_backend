package repository

import "gorm.io/gorm"

type ProductInFridgeRepository struct {
	db *gorm.DB
}