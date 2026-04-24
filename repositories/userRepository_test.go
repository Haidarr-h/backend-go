package repositories_test

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}

var dbError = errors.New("db error")

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return gormDB, mock, sqlDB
}

func userColumns() []string {
	return []string{
		"id", "created_at", "updated_at", "deleted_at",
		"email", "first_name", "last_name", "username", "google_id", "picture", "password",
	}
}

// ===== FindByEmail =====

func TestFindByEmail_Success(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows(userColumns()).
		AddRow(1, now, now, nil, "test@example.com", "John", "Doe", "johndoe", nil, nil, "hashed")

	mock.ExpectQuery(`SELECT \* FROM users WHERE email`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.FindByEmail("test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "johndoe", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_NotFound(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	rows := sqlmock.NewRows(userColumns())

	mock.ExpectQuery(`SELECT \* FROM users WHERE email`).
		WithArgs("notfound@example.com").
		WillReturnRows(rows)

	_, err := repo.FindByEmail("notfound@example.com")

	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_DBError(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM users WHERE email`).
		WithArgs("test@example.com").
		WillReturnError(dbError)

	_, err := repo.FindByEmail("test@example.com")

	assert.ErrorIs(t, err, dbError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== ExistByEmail =====

func TestExistByEmail_True(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(1))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	exists, err := repo.ExistByEmail("test@example.com")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExistByEmail_False(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WithArgs("nobody@example.com").
		WillReturnRows(rows)

	exists, err := repo.ExistByEmail("nobody@example.com")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExistByEmail_DBError(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WithArgs("test@example.com").
		WillReturnError(dbError)

	_, err := repo.ExistByEmail("test@example.com")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== ExistByUsername =====

func TestExistByUsername_True(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(1))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WithArgs("johndoe").
		WillReturnRows(rows)

	exists, err := repo.ExistByUsername("johndoe")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExistByUsername_False(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WithArgs("nobody").
		WillReturnRows(rows)

	exists, err := repo.ExistByUsername("nobody")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExistByUsername_DBError(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WithArgs("johndoe").
		WillReturnError(dbError)

	_, err := repo.ExistByUsername("johndoe")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== FindByUsername =====

func TestFindByUsername_Success(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows(userColumns()).
		AddRow(1, now, now, nil, "test@example.com", "John", "Doe", "johndoe", nil, nil, "hashed")

	mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("johndoe").
		WillReturnRows(rows)

	user, err := repo.FindByUsername("johndoe")

	assert.NoError(t, err)
	assert.Equal(t, "johndoe", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByUsername_NotFound(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	rows := sqlmock.NewRows(userColumns())

	mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("nobody").
		WillReturnRows(rows)

	_, err := repo.FindByUsername("nobody")

	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByUsername_DBError(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("johndoe").
		WillReturnError(dbError)

	_, err := repo.FindByUsername("johndoe")

	assert.ErrorIs(t, err, dbError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== CreateUser =====

func TestCreateUser_Success(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)
	now := time.Now()
	pass := "hashedpass"

	input := models.User{
		Email:     "new@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
		Password:  &pass,
	}

	rows := sqlmock.NewRows(userColumns()).
		AddRow(2, now, now, nil, "new@example.com", "Jane", "Doe", "janedoe", nil, nil, "hashedpass")

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("new@example.com", "Jane", "Doe", "janedoe", sqlmock.AnyArg()).
		WillReturnRows(rows)

	created, err := repo.CreateUser(input)

	assert.NoError(t, err)
	assert.Equal(t, "janedoe", created.Username)
	assert.Equal(t, uint(2), created.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_NoRowsAffected(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)
	pass := "hashedpass"

	input := models.User{
		Email:     "new@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
		Password:  &pass,
	}

	rows := sqlmock.NewRows(userColumns())

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("new@example.com", "Jane", "Doe", "janedoe", sqlmock.AnyArg()).
		WillReturnRows(rows)

	_, err := repo.CreateUser(input)

	assert.ErrorIs(t, err, repositories.ErrUserCreation)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_DBError(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repositories.NewUserRepository(gormDB)
	pass := "hashedpass"

	input := models.User{
		Email:     "new@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
		Password:  &pass,
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("new@example.com", "Jane", "Doe", "janedoe", sqlmock.AnyArg()).
		WillReturnError(dbError)

	_, err := repo.CreateUser(input)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
