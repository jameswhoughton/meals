package memory

import (
	"time"

	"github.com/jameswhoughton/meals/internal/auth"
)

type SessionRepository struct {
	Store []auth.Session
}

func (sr *SessionRepository) Create(session auth.Session) (auth.Session, error) {
	session.Id = len(sr.Store)

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

func (sr *SessionRepository) IsValid(sessionId string) bool {
	for _, session := range sr.Store {
		if session.SessionId == sessionId {
			return true
		}
	}

	return false
}

func (sr *SessionRepository) Get(sessionId string) (auth.Session, error) {
	for _, session := range sr.Store {
		if session.SessionId == sessionId {
			return session, nil
		}
	}

	return auth.Session{}, auth.ErrorSessionNotFound{}
}

func (sr *SessionRepository) DestroyByUserId(userId int) error {
	var sessions []auth.Session

	for _, session := range sr.Store {
		if session.UserId == userId {
			continue
		}
		sessions = append(sessions, session)
	}

	sr.Store = sessions

	return nil
}

func (sr *SessionRepository) DestroyExpired(olderThan time.Time) error {
	var sessions []auth.Session

	for _, session := range sr.Store {
		if olderThan.Compare(session.CreatedAt) < 1 {
			continue
		}

		sessions = append(sessions, session)
	}

	sr.Store = sessions

	return nil
}
