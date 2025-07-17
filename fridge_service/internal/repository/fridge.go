package repository

import (
	"gorm.io/gorm"
	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
)

type FridgeRepository struct {
	db *gorm.DB
}

func (f *FridgeRepository) CreateFridge() (*model.Fridge, error)  
func (f *FridgeRepository) GetFridgeById() (*model.Fridge, error) 
func (f *FridgeRepository) UpdateFridge() (*model.Fridge, error) 
func (f *FridgeRepository) DeleteFridge() error 
