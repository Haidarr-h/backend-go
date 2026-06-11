package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Mock Repository
// ---------------------------------------------------------------------------

type mockRepo struct{ mock.Mock }

func (m *mockRepo) CreateUser(u User) (User, error) {
	args := m.Called(u)
	return args.Get(0).(User), args.Error(1)
}
func (m *mockRepo) UpdateUser(u User) (User, error) {
	args := m.Called(u)
	return args.Get(0).(User), args.Error(1)
}
func (m *mockRepo) DeleteUser(u User) error {
	args := m.Called(u)
	return args.Error(0)
}
func (m *mockRepo) FindByEmail(email string) (User, error) {
	args := m.Called(email)
	return args.Get(0).(User), args.Error(1)
}
func (m *mockRepo) FindByUsername(username string) (User, error) {
	args := m.Called(username)
	return args.Get(0).(User), args.Error(1)
}
func (m *mockRepo) FindByGoogleID(googleID string) (User, error) {
	args := m.Called(googleID)
	return args.Get(0).(User), args.Error(1)
}
func (m *mockRepo) FindByID(userID uint) (User, error) {
	args := m.Called(userID)
	return args.Get(0).(User), args.Error(1)
}
func (m *mockRepo) ExistByGoogleID(googleID string) (bool, error) {
	args := m.Called(googleID)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) ExistByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) ExistByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) Verify(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// GetMyData
// ---------------------------------------------------------------------------

func TestGetMyData(t *testing.T) {
	t.Run("success maps user to response", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("FindByID", uint(1)).Return(User{
			Model:      gorm.Model{ID: 1},
			Email:      "jane@example.com",
			FirstName:  "Jane",
			LastName:   "Doe",
			Username:   "janedoe",
			IsVerified: true,
		}, nil)

		svc := NewUserService(repo)
		res, err := svc.GetMyData(1)

		assert.NoError(t, err)
		assert.Equal(t, "jane@example.com", res.Email)
		assert.Equal(t, "janedoe", res.Username)
		assert.True(t, res.IsVerified)
		repo.AssertExpectations(t)
	})

	t.Run("user not found propagates", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("FindByID", uint(1)).Return(User{}, ErrUserNotFound)

		svc := NewUserService(repo)
		_, err := svc.GetMyData(1)

		assert.ErrorIs(t, err, ErrUserNotFound)
		repo.AssertExpectations(t)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		repo := &mockRepo{}
		dbErr := errors.New("db error")
		repo.On("FindByID", uint(1)).Return(User{}, dbErr)

		svc := NewUserService(repo)
		_, err := svc.GetMyData(1)

		assert.ErrorIs(t, err, dbErr)
		repo.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{}
		existing := User{Model: gorm.Model{ID: 1}, Email: "jane@example.com"}
		repo.On("FindByID", uint(1)).Return(existing, nil)
		repo.On("DeleteUser", existing).Return(nil)

		svc := NewUserService(repo)
		err := svc.Delete(1)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("user not found - delete never called", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("FindByID", uint(1)).Return(User{}, ErrUserNotFound)

		svc := NewUserService(repo)
		err := svc.Delete(1)

		assert.ErrorIs(t, err, ErrUserNotFound)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "DeleteUser", mock.Anything)
	})

	t.Run("delete error propagates", func(t *testing.T) {
		repo := &mockRepo{}
		existing := User{Model: gorm.Model{ID: 1}}
		dbErr := errors.New("fk violation")
		repo.On("FindByID", uint(1)).Return(existing, nil)
		repo.On("DeleteUser", existing).Return(dbErr)

		svc := NewUserService(repo)
		err := svc.Delete(1)

		assert.ErrorIs(t, err, dbErr)
		repo.AssertExpectations(t)
	})
}
