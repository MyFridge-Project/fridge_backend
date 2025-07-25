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
)

func DBMock(t *testing.T) (*sql.DB, *gorm.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqldb,
	}), &gorm.Config{})
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
	var fridgeID int64 = 52
	var role = "admin"

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `fridges`")).
		WithArgs(sqlmock.AnyArg(), name, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(fridgeID, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `fridge_users`")).
		WithArgs(fmt.Sprintf("%d", fridgeID), userID, role, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	fridge, err := repo.CreateFridge(ctx, name, userID)

	require.NoError(t, err)
	require.NotNil(t, fridge)
	require.Equal(t, name, fridge.Name)
	require.Equal(t, fmt.Sprintf("%d", fridgeID), fridge.ID)

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

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fridges` WHERE id = ?")).
		WithArgs(fridgeID).
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

	mock.ExpectQuery("SELECT * FROM `fridges` WHERE id = ?").
		WithArgs(fridgeID).
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
	
}

func TestUpdateFridge_Fail(t *testing.T) {

}

func TestDeleteFridge_Success(t *testing.T) {

}

func TestDeleteFridge_Fail(t *testing.T) {

}
