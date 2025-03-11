package memory

import (
	"time"

	"github.com/jameswhoughton/meals/internal/auth"
)

type SessionRepository struct {
	Store []auth.Session
}

func (sr *SessionRepository) Create(session auth.Session) (auth.Session, error) {
	session.Id = len(sr.Store) + 1

	sr.Store = append(sr.Store, session)

	return session, nil
}

func (sr *SessionRepository) Destroy(sessionId string) error {
	var index *int

	for i, session := range sr.Store {
		if session.SessionId == sessionId {
			index = &i
			break
		}
	}

	if index == nil {
		return auth.ErrorSessionNotFound{}
	}

	existingStore := sr.Store

	sr.Store = make([]auth.Session, 0)

	sr.Store = append(sr.Store, existingStore[:*index]...)
	sr.Store = append(sr.Store, existingStore[*index+1:]...)

	return nil
}

func (sr *SessionRepository) Get(sessionId string, expiredTime time.Time) (auth.Session, error) {
	for i, session := range sr.Store {
		if session.SessionId == sessionId && session.UpdatedAt.After(expiredTime) {
			session.UpdatedAt = time.Now()
			sr.Store[i] = session

			return sr.Store[i], nil
		}
	}

	return auth.Session{}, auth.ErrorSessionNotFound{}
}

func (sr *SessionRepository) DestroyExpired(expiredTime time.Time) error {
	var sessions []auth.Session

	for _, session := range sr.Store {
		if session.UpdatedAt.Before(expiredTime) {
			continue
		}

		sessions = append(sessions, session)
	}

	sr.Store = sessions

	return nil
}
