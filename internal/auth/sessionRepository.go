package auth

import "time"

type Session struct {
	Id        int
	SessionId string
	UserId    int
	CreatedAt time.Time
}

type ErrorSessionNotFound struct {
}

func (e ErrorSessionNotFound) Error() string {
	return "Session not found"
}

type SessionRepository interface {
	Create(session Session) (Session, error)
	Destroy(sessionId string) error
	IsValid(sessionId string) bool
	Get(sessionId string) (Session, error)
	DestroyByUserId(userId int) error
	DestroyExpired(olderThan time.Time) error
}
