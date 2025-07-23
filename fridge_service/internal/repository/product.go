package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

func (p *ProductRepository) UpsertProduct(ctx context.Context, name string, expirationTime *time.Time) (product *model.Product, err error) {
	tx := p.db.WithContext(ctx).Begin()

	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	newProduct := model.Product{
		Name:           name,
		ExpirationTime: expirationTime,
	}

	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"expiration_time": gorm.Expr(
				"IF(VALUES(expiration_time) IS NULL, expiration_time, VALUES(expiration_time))",
			),
		}),
	}).Create(&newProduct).Error

	if err != nil {
		return nil, fmt.Errorf("failed to create a new product %s: %w", name, err)
	}

	return &newProduct, nil
}

func (p *ProductRepository) GetProductByID(ctx context.Context, productID string) (*model.Product, error) {
	var product model.Product

	err := p.db.WithContext(ctx).Where("id = ?", productID).First(&product).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id %s: %w", productID, err)
	}

	return &product, nil
}

func (p *ProductRepository) GetProductByName(ctx context.Context, productName string) (*model.Product, error) {
	var product model.Product
	
	err := p.db.WithContext(ctx).Where("name = ?", productName).First(&product).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get product by name %s: %w", productName, err)
	}
	
	return &product, nil
}

func (p *ProductRepository) ListProducts(
	ctx context.Context, 
	limit int, 
	lastCreatedAt time.Time, 
	lastProductID string,
) ([]*model.Product, error) {
	var products []*model.Product
	
	err := p.db.WithContext(ctx).
		Where("(created_at, id) > (?,?)", lastCreatedAt, lastProductID).
		Order("created_at, id").
		Limit(limit).
		Find(&products).
		Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	
	return products, nil
}


func (p *ProductRepository) UpdateProduct(
	ctx context.Context,
	productID string,
	updateFn func(*model.Product) (*model.Product, error),
) (updatedProduct *model.Product, err error) {
	tx := p.db.WithContext(ctx).Begin()

	defer func() {
		err = p.finishTransaction(err, tx)
	}()
	
	var oldProduct model.Product
	if err = tx.Where("id = ?", productID).First(&oldProduct).Error; err != nil {
		return nil, fmt.Errorf("failed to get product by id %s to update it: %w", productID, err)
	}
	
	updatedProduct, err = updateFn(&oldProduct)
	if err != nil {
		return nil, fmt.Errorf("failed to update product with id %s: %w", productID, err)
	}
	
	if err = tx.Save(updatedProduct).Error; err != nil {
		return nil, fmt.Errorf("failed to save changes when updatind product %s: %w", productID, err)
	}

	return updatedProduct, nil
}

func (p *ProductRepository) DeleteProduct(ctx context.Context, productID string) (err error) {
	tx := p.db.WithContext(ctx).Begin()

	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	if err = tx.Where("product_id = ?", productID).
		Delete(&model.ProductInFridge{}).Error; err != nil {
		return fmt.Errorf("failed to delete all products in fridges with product_id %s: %w", productID, err)
	}

	if err = tx.Where("id = ?", productID).
		Delete(&model.Product{}).Error; err != nil {
		return fmt.Errorf("failed to delete product with id %s: %w", productID, err)
	}

	return nil
}

func (p *ProductRepository) finishTransaction(origErr error, tx *gorm.DB) error {
	if origErr != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("%v; rollback error: %w", origErr, rbErr)
		}
		return origErr
	}

	if commitErr := tx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}
