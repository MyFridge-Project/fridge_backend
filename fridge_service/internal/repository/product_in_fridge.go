package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
	"gorm.io/gorm"
)

type ProductInFridgeRepository struct {
	db *gorm.DB
}

func (p *ProductInFridgeRepository) AddProductToFridge(
	ctx context.Context,
	fridgeID string,
	productID string,
	quantity float64,
	productionDate *time.Time,
	expirationDate *time.Time,
) (productInFridge *model.ProductInFridge, err error) {
	tx := p.db.WithContext(ctx).Begin()

	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	product := model.ProductInFridge{
		FridgeID:       fridgeID,
		ProductID:      productID,
		Quantity:       quantity,
		ProductionDate: productionDate,
		ExpirationDate: expirationDate,
	}

	if err = tx.Create(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to add product %s to fridge %s: %w", productID, fridgeID, err)
	}

	return &product, err
}

func (p *ProductInFridgeRepository) GetProductInFridge(
	ctx context.Context,
	fridgeID string,
	productID string,
) (*model.ProductInFridge, error) {
	var product model.ProductInFridge

	err := p.db.WithContext(ctx).
		Where("fridge_id = ?", fridgeID).
		Where("product_id = ?", productID).
		First(&product).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get product with id %s in fridge %s: %w", productID, fridgeID, err)
	}

	return &product, nil
}

func (p *ProductInFridgeRepository) ListProductsInFridge(
	ctx context.Context,
	fridgeID string,
	limit int,
	lastCreatedAt time.Time,
	lastProductID string,
) ([]*model.ProductInFridge, error) {
	var products []*model.ProductInFridge

	err := p.db.WithContext(ctx).
		Where("fridge_id = ?", fridgeID).
		Where("(created_at, id) >  (?, ?)", lastCreatedAt, lastProductID).
		Order("created_at, product_id").
		Limit(limit).
		Find(&products).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to list products in fridge %s: %w", fridgeID, err)
	}

	return products, nil
}

func (p *ProductInFridgeRepository) UpdateProductInFridge(
	ctx context.Context,
	fridgeID string,
	productID string,
	updateFn func(*model.ProductInFridge) (*model.ProductInFridge, error),
) (productInFridge *model.ProductInFridge, err error) {
	tx := p.db.WithContext(ctx).Begin()

	defer func() {
		err = p.finishTransaction(err, tx)
	}()
	
	var product model.ProductInFridge

	if err = tx.Where("fridge_id = ?", fridgeID).
		Where("product_id = ?", productID).
		First(&product).
		Error; err != nil {
		return nil, fmt.Errorf("failed to get product with id %s in fridge %s: %w", productID, fridgeID, err)
	}
	
	updatedProduct, err := updateFn(&product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product with id %s in fridge %s: %w", productID, fridgeID, err)
	}
	
	if err = tx.Save(updatedProduct).Error; err != nil {
		return nil, fmt.Errorf("failed to save changes when updating product %s in fridge %s: %w", productID, fridgeID, err)
	}

	return updatedProduct, nil
}

func (p *ProductInFridgeRepository) DeleteProductFromFridge(
	ctx context.Context,
	fridgeID string,
	productID string,
) (err error) {
	tx := p.db.WithContext(ctx).Begin()

	defer func() {
		err = p.finishTransaction(err, tx)
	}()

	err = tx.Where("fridge_id = ?", fridgeID).
		Where("product_id = ?", productID).
		Delete(&model.ProductInFridge{}).
		Error

	if err != nil {
		return fmt.Errorf("failed to delete product in fridge %s by id %s: %w", fridgeID, productID, err)
	}

	return nil
}

func (p *ProductInFridgeRepository) finishTransaction(origErr error, tx *gorm.DB) error {
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
