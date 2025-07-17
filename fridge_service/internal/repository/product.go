package repository

import (
	"gorm.io/gorm"
	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
)

type ProductRepository struct {
	db *gorm.DB
}

func (p* ProductRepository) CreateProduct() (*model.Product, error)
func (p* ProductRepository) GetProductById() (*model.Product, error)
func (p* ProductRepository) UpdateProduct() (*model.Product, error)
func (p* ProductRepository) DeleteProduct() error