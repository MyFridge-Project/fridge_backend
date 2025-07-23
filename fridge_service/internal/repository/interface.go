package repository

import (
	"context"
	"time"

	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
)

type RepositoryProvider interface {
	Fridge() FridgeProvider
	Product() ProductProvider
	ProductInFridge() ProductInFridgeProvider
}

type FridgeProvider interface {
	CreateFridge(context.Context, string, string) (*model.Fridge, error)
	GetFridgeByID(context.Context, string) (*model.Fridge, error)
	// AddUserToFridge() (, error)
	ListUserFridges(context.Context, int, string, time.Time, string) ([]*model.Fridge, error)
	UpdateFridge(context.Context, string, func(*model.Fridge)(*model.Fridge, error)) (*model.Fridge, error)
	DeleteFridge(context.Context, string) error
}

type ProductProvider interface {
	CreateProduct() (*model.Product, error)
	GetProductById() (*model.Product, error)
	UpdateProduct() (*model.Product, error)
	DeleteProduct() error
}

type ProductInFridgeProvider interface {
}
