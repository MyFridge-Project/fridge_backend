package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MyFridge-Project/fridge_backend/fridge_service/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBMock(t *testing.T) (*sql.DB, *gorm.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	
	mock.ExpectQuery(regexp.QuoteMeta("SELECT VERSION()")).
        WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.30"))
	
	gormdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqldb,
	}), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)

	return sqldb, gormdb, mock
}

func TestCreateFridge(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	var name = "TestFridge"
	var userID = "user-123"
	var role = "admin"
	//fridgeID := uuid.New().String()
	
	//mock.ExpectQuery(regexp.QuoteMeta("SELECT VERSION()"))
	
	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `fridges`")).
		WithArgs(sqlmock.AnyArg(), name, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `fridge_users`")).
		WithArgs(sqlmock.AnyArg(), userID, role, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	fridge, err := repo.CreateFridge(ctx, name, userID)

	require.NoError(t, err)
	require.NotNil(t, fridge)
	require.Equal(t, name, fridge.Name)
	//require.Equal(t, fridgeID, fridge.ID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFridgeByID_Success(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()

	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	var fridgeID = "123"
	var fridgeName = "TestFridge"

	rows := mock.NewRows([]string{
		"id",
		"name",
		"created_at",
		"updated_at",
	}).AddRow(
		fridgeID,
		fridgeName,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fridges` WHERE id = ? ORDER BY `fridges`.`id` LIMIT ?")).
		WithArgs(fridgeID, 1).
		WillReturnRows(rows)

	fridge, err := repo.GetFridgeByID(ctx, fridgeID)

	require.NoError(t, err)
	assert.NotNil(t, fridge)
	assert.Equal(t, fridgeID, fridge.ID)
	assert.Equal(t, fridgeName, fridge.Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFridgeByID_Fail(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()

	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	var fridge *model.Fridge
	var fridgeID = "nonexistent"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fridges` WHERE id = ? ORDER BY `fridges`.`id` LIMIT ?")).
		WithArgs(fridgeID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	fridge, err := repo.GetFridgeByID(ctx, fridgeID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	assert.Nil(t, fridge)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListFridges_First_Success(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()

	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	now := time.Now()
	fridge1ID := "1"
	fridge2ID := "2"
	fridge1Name := "fridge1"
	fridge2Name := "fridge2"
	limit := 2

	rows := mock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRows(
			[]driver.Value{
				fridge1ID,
				fridge1Name,
				now,
				now,
			},
			[]driver.Value{
				fridge2ID,
				fridge2Name,
				now.Add(time.Minute),
				now.Add(time.Minute),
			},
		)

	expectQuery := "SELECT * FROM `fridges` WHERE (created_at, id) > (?, ?) ORDER BY created_at, id LIMIT ?"
	mock.ExpectQuery(regexp.QuoteMeta(expectQuery)).
		WithArgs(time.Time{}, "", limit).
		WillReturnRows(rows)

	fridges, err := repo.ListFridges(ctx, limit, time.Time{}, "")

	require.NoError(t, err)
	require.NotNil(t, fridges)
	require.Len(t, fridges, 2)
	
	assert.Equal(t, fridge1ID, fridges[0].ID)
	assert.Equal(t, fridge1Name, fridges[0].Name)
	assert.Equal(t, fridge2ID, fridges[1].ID)
	assert.Equal(t, fridge2Name, fridges[1].Name)	
	
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListFridges_Middle_Success(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	limit := 2
	
	fridge2ID := "2"
	fridge3ID := "3"
	fridge3Name := "fridge3"
	fridge4ID := "4"
	fridge4Name := "fridge4"
	
	now := time.Now()
	times := make([]time.Time, 4)
	for i := range(4) {
		times[i] = now.Add(time.Duration(i) * time.Minute)
	}
	
	rowsData := [][]driver.Value{
		{"1", "fridge1", times[0], times[0]},
		{"2", "fridge2", times[1], times[1]},
		{fridge3ID, fridge3Name, times[2], times[2]},
		{"4", fridge4Name, times[3], times[3]},
	}
	
	returnRows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).AddRows(rowsData[2:]...)
	
	expectQuery := "SELECT * FROM `fridges` WHERE (created_at, id) > (?, ?) ORDER BY created_at, id LIMIT ?"
	mock.ExpectQuery(regexp.QuoteMeta(expectQuery)).
		WithArgs(times[1], fridge2ID, limit).
		WillReturnRows(returnRows)
	
	fridges, err := repo.ListFridges(ctx, limit, times[1], fridge2ID)
	
	require.NoError(t, err)
	require.NotNil(t, fridges)
	require.Len(t, fridges, 2)
	
	assert.Equal(t, fridge3ID, fridges[0].ID)
	assert.Equal(t, fridge3Name, fridges[0].Name)	
	assert.Equal(t, fridge4ID, fridges[1].ID)
	assert.Equal(t, fridge4Name, fridges[1].Name)
	
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListFridges_Fail(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	limit := 2
	
	expectError := errors.New("test error")
	expectQuery := "SELECT * FROM `fridges` WHERE (created_at, id) > (?, ?) ORDER BY created_at, id LIMIT ?"
	mock.ExpectQuery(regexp.QuoteMeta(expectQuery)).
		WithArgs(time.Time{}, "", limit).
		WillReturnError(expectError)
	
	fridges, err := repo.ListFridges(ctx, limit, time.Time{}, "")
	
	require.Error(t, err)
	require.Nil(t, fridges)
	assert.Contains(t, err.Error(), expectError.Error())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateFridge_Success(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	fridgeID := "1"
	oldName := "SomeName"
	now := time.Now()
	
	newName := "SomeNewName"
	updateFn := func(fridge *model.Fridge) (*model.Fridge, error) {
		if fridge == nil {
			return nil, fmt.Errorf("fridge to update cannot be nil")
		}
		
		fridge.Name = newName
		return fridge, nil
	}
	
	selectRows := mock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRow(fridgeID, oldName, now, now)
	
	mock.ExpectBegin()
	
	expectQuery1 := "SELECT * FROM `fridges` WHERE id = ? ORDER BY `fridges`.`id` LIMIT ?"
	mock.ExpectQuery(regexp.QuoteMeta(expectQuery1)).
		WithArgs(fridgeID, 1).
		WillReturnRows(selectRows)
	
	expectQuery2 := "UPDATE `fridges` SET `name`=?,`created_at`=?,`updated_at`=? WHERE `id` = ?"
	mock.ExpectExec(regexp.QuoteMeta(expectQuery2)).
		WithArgs(newName, sqlmock.AnyArg(), sqlmock.AnyArg(), fridgeID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	mock.ExpectCommit()
	
	updatedFridge, err := repo.UpdateFridge(ctx, fridgeID, updateFn)
	
	require.NoError(t, err)
	require.NotNil(t, updatedFridge)
	
	assert.Equal(t, fridgeID, updatedFridge.ID)
	assert.Equal(t, newName, updatedFridge.Name)
	
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateFridge_Fail(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	fridgeID := "nonexistent"
	
	updateFn := func(fridge *model.Fridge) (*model.Fridge, error) {return fridge, nil}
	
	mock.ExpectBegin()
	
	expectedQuery := "SELECT * FROM `fridges` WHERE id = ? ORDER BY `fridges`.`id` LIMIT ?"
	mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
		WithArgs(fridgeID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	
	mock.ExpectRollback()
	
	fridge, err := repo.UpdateFridge(ctx, fridgeID, updateFn)
	
	require.Error(t, err)
	require.Nil(t, fridge)
	
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteFridge_Success(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo := NewFridgeRepository(gormdb)
	ctx := context.Background()
	fridgeID := "1"
	
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	
	mock.ExpectBegin()
	
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `fridges` WHERE id = ?")).
		WithArgs(fridgeID).
		WillReturnRows(countRows)
	
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `product_in_fridges` WHERE fridge_id =  ?")).
		WithArgs(fridgeID).WillReturnResult(sqlmock.NewResult(0, 1))
	
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `fridge_users` WHERE fridge_id = ?")).
		WithArgs(fridgeID).WillReturnResult(sqlmock.NewResult(0, 1))
	
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `fridges` WHERE id = ?")).
		WithArgs(fridgeID).WillReturnResult(sqlmock.NewResult(0, 1))
	
	mock.ExpectCommit()
	
	err := repo.DeleteFridge(ctx, fridgeID)
	
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteFridge_Fail(t *testing.T) {
	sqldb, gormdb, mock := DBMock(t)
	defer sqldb.Close()
	
	repo :=  NewFridgeRepository(gormdb)
	ctx := context.Background()
	fridgeID := "nonexistent"
	
	mock.ExpectBegin()
	
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `fridges` WHERE id = ?")).
		WithArgs(fridgeID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	
	mock.ExpectRollback()
	
	err := repo.DeleteFridge(ctx, fridgeID)
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	require.NoError(t, mock.ExpectationsWereMet())
}
