package repository

import "github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"

type RepositoryProvider interface {
	Fridge() FridgeProvider
	Product() ProductProvider
	ProductInFridge() ProductInFridgeProvider
}

type FridgeProvider interface {
	CreateFridge() (*model.Fridge, error)
	GetFridgeById() (*model.Fridge, error)
	UpdateFridge() (*model.Fridge, error)
	DeleteFridge() error
}

type ProductProvider interface {
	CreateProduct() (*model.Product, error)
	GetProductById() (*model.Product, error)
	UpdateProduct() (*model.Product, error)
	DeleteProduct() error
}

type ProductInFridgeProvider interface {
}
