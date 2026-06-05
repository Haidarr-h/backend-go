package user

type Service interface {
	GetMyData(userID uint) (UserResponse, error)
	Delete(userID uint) error
}

type Repository interface {
	CreateUser(user User) (User, error)
	UpdateUser(user User) (User, error)
	DeleteUser(user User) error
	FindByEmail(email string) (User, error)
	FindByUsername(username string) (User, error)
	FindByGoogleID(googleID string) (User, error)
	FindByID(userID uint) (User, error)
	ExistByGoogleID(googleID string) (bool, error)
	ExistByEmail(email string) (bool, error)
	ExistByUsername(username string) (bool, error)
	Verify(userID uint) error
}
