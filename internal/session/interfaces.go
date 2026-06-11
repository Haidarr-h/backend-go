package session

type Service interface {
	CreateSession(userID uint, req CreateSessionRequest) (SessionResponse, error)
	GetSessions(userID uint) ([]SessionResponse, error)
	GetSession(userID uint, sessionID uint) (SessionResponse, error)
	UpdateSession(id, userID uint, req UpdateSessionReq) (SessionResponse, error)
	DeleteSession(sessionID uint, userID uint) error
}

type Repo interface {
	Create(session Session) (Session, error)
	FindAll(userID uint) ([]Session, error)
	FindByID(id uint) (Session, error)
	Update(session Session) (Session, error)
	Delete(id uint, userID uint) error
}
