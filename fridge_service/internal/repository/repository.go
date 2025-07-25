package repository

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) RepositoryProvider {
	return &Repository{db: db}
} 

func (r* Repository) Fridge() FridgeProvider {
	return &FridgeRepository{db: r.db}
}

func (r* Repository) Product() ProductProvider {
	return &ProductRepository{db: r.db}
}

func (r* Repository) ProductInFridge() ProductInFridgeProvider {
	return &ProductInFridgeRepository{db: r.db}
}

func (r *Repository) FridgeUser() FridgeUserProvider {
	return &FridgeUserRepository{db: r.db}
}