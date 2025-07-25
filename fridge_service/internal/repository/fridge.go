package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
	"gorm.io/gorm"
)

type FridgeRepository struct {
	db *gorm.DB
}

func NewFridgeRepository(db *gorm.DB) *FridgeRepository {
	return &FridgeRepository{db: db}
}

func (f *FridgeRepository) CreateFridge(ctx context.Context, name string, userID string) (fridge *model.Fridge, err error) {
	tx := f.db.WithContext(ctx).Begin()
	
	defer func() {
		err = f.finishTransaction(err, tx)
	}()
	
	newFridge := model.Fridge{
		Name: name,
	}
	
	if err = tx.Create(&newFridge).Error; err != nil {
		return nil, fmt.Errorf("failed to create fridge: %w", err)
	}
	
	fridgeUser := model.FridgeUser{
		FridgeID: newFridge.ID,
		UserID: userID,
		Role: "admin",
	}
	
	if err = tx.Create(&fridgeUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create fridge-user link: %w", err)
	}
	
	return &newFridge, nil
}

func (f *FridgeRepository) GetFridgeByID(ctx context.Context, fridgeID string) (*model.Fridge, error) {
	var fridge model.Fridge
	
	if err := f.db.WithContext(ctx).First(&fridge, "id = ?", fridgeID).Error; err != nil {
		return nil, fmt.Errorf("failed to get fridge by id %s: %w", fridgeID, err)
	}
	
	return &fridge, nil
}

func (f *FridgeRepository) ListFridges(ctx context.Context, limit int, lastCreatedAt time.Time, lastFridgeID string) ([]*model.Fridge, error) {
	var fridges []*model.Fridge
	
	err := f.db.WithContext(ctx).
		Where("(created_at, id) > (?, ?)", lastCreatedAt, lastFridgeID).
		Order("created_at, id").
		Limit(limit).
		Find(&fridges).
		Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to list fridges: %w", err)
	}
	
	return fridges, nil
}

func (f *FridgeRepository) UpdateFridge(
	ctx context.Context, 
	fridgeID string, 
	updateFn func(*model.Fridge)(*model.Fridge, error),
) (fridge *model.Fridge, err error) {
	tx := f.db.WithContext(ctx).Begin()
	
	defer func(){
		err = f.finishTransaction(err, tx)
	}()
	
	var oldFridge model.Fridge
	if err = tx.First(&oldFridge, "id = ?", fridgeID).Error; err != nil {
		return nil, fmt.Errorf("failed to get fridge %s to update it: %w", fridgeID, err)
	}
	
	updatedFridge, err := updateFn(&oldFridge)
	if err != nil {
		return nil, fmt.Errorf("failed to update fridge %s: %w", fridgeID, err)
	}
	
	if err = tx.Save(updatedFridge).Error; err != nil {
		return nil, fmt.Errorf("failed to save updated fridge %s: %w", fridgeID, err)
	}
	
	return updatedFridge, nil
} 

func (f *FridgeRepository) DeleteFridge(ctx context.Context, fridgeID string) (err error) {
	tx := f.db.WithContext(ctx).Begin()
	
	defer func() {
		err = f.finishTransaction(err, tx)
	}()
	
	if err = tx.Where("fridge_id = ?", fridgeID).Delete(&model.ProductInFridge{}).Error; err != nil {
		return fmt.Errorf("failed to delete all products in fridge with id %s: %w", fridgeID, err)
	}
	
	if err = tx.Where("fridge_id = ?", fridgeID).Delete(&model.FridgeUser{}).Error; err != nil {
		return fmt.Errorf("failed to delete fridge-users links: %w", err)
	}
	
	if err = tx.Where("id = ?", fridgeID).Delete(&model.Fridge{}).Error; err != nil {
		return fmt.Errorf("failed to delete fridge by id %s: %w", fridgeID, err)
	}
	
	return nil
} 

func (f *FridgeRepository) finishTransaction(origErr error, tx *gorm.DB) error {
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

