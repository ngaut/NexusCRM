package persistence

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestCheckUserExistsByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	email := "test@example.com"
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", constants.TableUser, constants.FieldEmail)

	// Test Case 1: User exists
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.CheckUserExistsByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test Case 2: User does not exist
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("nonexistent@example.com").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = repo.CheckUserExistsByEmail(context.Background(), "nonexistent@example.com")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCheckUserExistsByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	id := "user-123"
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", constants.TableUser, constants.FieldID)

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.CheckUserExistsByID(context.Background(), id)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCheckEmailConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	email := "test@example.com"
	excludeID := "user-123"
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ? AND %s != ?)", constants.TableUser, constants.FieldEmail, constants.FieldID)

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(email, excludeID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.CheckEmailConflict(context.Background(), email, excludeID)
	assert.NoError(t, err)
	assert.True(t, exists)
}
