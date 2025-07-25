package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"
	"gorm.io/gorm"
)

type FridgeUserRepository struct {
	db *gorm.DB
}

func NewFridgeUserRepository(db *gorm.DB) *FridgeUserRepository {
	return &FridgeUserRepository{}
}

func (f *FridgeUserRepository) AddUserToFridge(
	ctx context.Context,
	fridgeID string,
	userID string,
	role string,
) (fridgeUser *model.FridgeUser, err error) {
	tx := f.db.WithContext(ctx).Begin()

	defer func() {
		err = f.finishTransaction(err, tx)
	}()

	var existingUser model.FridgeUser

	if err = tx.Where("fridge_id = ?", fridgeID).
		Where("user_id = ?", userID).
		First(&existingUser).
		Error; err == nil {
		return &existingUser, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newFridgeUser := model.FridgeUser{
		FridgeID: fridgeID,
		UserID:   userID,
		Role:     role,
	}

	if err = tx.Create(&newFridgeUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create fridgeUser: %w", err)
	}

	return &newFridgeUser, nil
}

func (f *FridgeUserRepository) ListUserFridges(
	ctx context.Context,
	limit int,
	userID string,
	lastCreatedAt time.Time,
	lastFridgeID string,
) ([]*model.Fridge, error) {
	var fridges []*model.Fridge

	err := f.db.WithContext(ctx).
		Joins("JOIN fridge_users fu ON fu.fridge_id = fridges.id").
		Where("fu.user_id = ?", userID).
		Where("(fridges.created_at, fridges.id) > (?, ?)", lastCreatedAt, lastFridgeID).
		Order("fridges.created_at, fridges.id").
		Limit(limit).
		Find(&fridges).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to list fridges for user %s: %w", userID, err)
	}

	return fridges, nil
}

func (f *FridgeUserRepository) ListFridgeUsers(
	ctx context.Context,
	limit int,
	fridgeID string,
	lastCreatedAt time.Time,
	lastUserID string,
) ([]*model.User, error) {
	var users []*model.User

	err := f.db.WithContext(ctx).
		Joins("JOIN fridge_users fu ON fu.user_id = users.id").
		Where("fu.fridge_id = ?", fridgeID).
		Where("(users.created_at, users.id) > (?, ?)", lastCreatedAt, lastUserID).
		Order("users.created_at, users.id").
		Limit(limit).
		Find(&users).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to list users for fridge %s: %w", fridgeID, err)
	}

	return users, nil
}

func (f *FridgeUserRepository) DeleteUserFromFridge(ctx context.Context, fridgeID, userID string) (err error) {
	tx := f.db.WithContext(ctx).Begin()

	defer func() {
		err = f.finishTransaction(err, tx)
	}()

	if err = tx.Where("fridge_id = ?", fridgeID).
		Where("user_id = ?", userID).
		Delete(&model.FridgeUser{}).
		Error; err != nil {
		return fmt.Errorf("failed to delete user with id %s in fridge %s: %w", userID, fridgeID, err)
	}

	return nil
}

func (f *FridgeUserRepository) finishTransaction(origErr error, tx *gorm.DB) error {
	if origErr != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("%v; rollback error: %w", origErr, rbErr)
		}

		return origErr
	}

	if commitErr := tx.Commit().Error; commitErr != nil {
		return fmt.Errorf("commit error: %w", commitErr)
	}
	return nil
}
