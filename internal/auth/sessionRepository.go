package auth

import "time"

type Session struct {
	Id        int
	SessionId string
	UserId    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ErrorSessionNotFound struct {
}

func (e ErrorSessionNotFound) Error() string {
	return "Session not found"
}

type SessionRepository interface {
	Create(session Session) (Session, error)
	Destroy(sessionId string) error
	Get(sessionId string, expiredTime time.Time) (Session, error)
	DestroyExpired(expiredTime time.Time) error
}
