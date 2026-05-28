package user

type Repository interface {
	CreateUser(user User) (User, error)
	UpdateUser(user User) (User, error)
	FindByEmail(email string) (User, error)
	FindByUsername(username string) (User, error)
	FindByGoogleID(googleID string) (User, error)
	ExistByEmail(email string) (bool, error)
	ExistByUsername(username string) (bool, error)
	Verify(userID uint) error
}
