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
	ListFridges(context.Context, int, time.Time, string) ([]*model.Fridge, error)
	UpdateFridge(context.Context, string, func(*model.Fridge)(*model.Fridge, error)) (*model.Fridge, error)
	DeleteFridge(context.Context, string) error
}

type ProductProvider interface {
	UpsertProduct(context.Context, string, *time.Time) (*model.Product, error)
	GetProductByID(context.Context, string) (*model.Product, error)
	GetProductByName(context.Context, string) (*model.Product, error)
	ListProducts(context.Context, int, time.Time, string) ([]*model.Product, error)
	UpdateProduct(context.Context, string, func(*model.Product)(*model.Product, error)) (*model.Product, error)
	DeleteProduct(context.Context, string) error
}

type ProductInFridgeProvider interface {
	AddProductToFridge(context.Context, string, string, float64, *time.Time, *time.Time) (*model.ProductInFridge, error)
	GetProductInFridge(context.Context, string, string) (*model.ProductInFridge, error)
	ListProductsInFridge(context.Context, string, int, time.Time, string) ([]*model.ProductInFridge, error)
	UpdateProductInFridge(context.Context, string, string, func(*model.ProductInFridge) (*model.ProductInFridge, error)) (*model.ProductInFridge, error)
	DeleteProductFromFridge(context.Context, string, string) (error)
}

type FridgeUserProvider interface {
	AddUserToFridge(context.Context, string, string, string) (*model.FridgeUser, error)
	ListUserFridges(context.Context, int, string, time.Time, string) ([]*model.Fridge, error)
	ListFridgeUsers(context.Context, int, string, time.Time, string) ([]*model.User, error)
	DeleteUserFromFridge(context.Context, string, string) error
}
