package user

import "errors"

type userService struct {
	repo Repository
}

func NewUserService(repo Repository) *userService {
	return &userService{repo: repo}
}

// GET MY USER DATA
func (s *userService) GetMyData(userID uint) (UserResponse, error) {

	user, err := s.repo.FindByID(userID)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return UserResponse{}, err
		}

		return UserResponse{}, err
	}

	return mapToUserResponse(user), nil
}

// DELETE USER DATA
func (s *userService) Delete(userID uint) error {

	// 1. check if user exists
	userData, err := s.repo.FindByID(userID)

	if err != nil {
		return err
	}

	if userData.ID == 0 {
		return ErrUserNotFound
	}

	// 2. perform deletion
	if err := s.repo.DeleteUser(userData); err != nil {
		return err
	}

	return nil
}


func mapToUserResponse(u User) UserResponse {
	var userData UserResponse

	userData.FirstName = u.FirstName
	userData.LastName = u.LastName
	userData.Email = u.Email
	userData.Username = u.Username
	userData.IsVerified = u.IsVerified
	userData.Picture = u.Picture

	return userData
}

