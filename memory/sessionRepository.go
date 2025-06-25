package memory

import (
	"context"
	"time"

	"github.com/jameswhoughton/meals/web"
)

type SessionRepository struct {
	store []web.Session
}

func (sr *SessionRepository) Create(ctx context.Context, session web.Session) (web.Session, error) {
	session.Id = len(sr.store) + 1

	sr.store = append(sr.store, session)

	return session, nil
}

func (sr *SessionRepository) Destroy(ctx context.Context, sessionId string) error {
	var index *int

	for i, session := range sr.store {
		if session.SessionId == sessionId {
			index = &i
			break
		}
	}

	if index == nil {
		return nil
	}

	existingstore := sr.store

	sr.store = make([]web.Session, len(existingstore)-1)

	sr.store = append(sr.store, existingstore[:*index]...)
	sr.store = append(sr.store, existingstore[*index+1:]...)

	return nil
}

func (sr *SessionRepository) Get(ctx context.Context, sessionId string) (web.Session, error) {
	for i, session := range sr.store {
		if session.SessionId == sessionId {
			sr.store[i] = session

			return sr.store[i], nil
		}
	}

	return web.Session{}, web.ErrSessionNotFound
}

func (sr *SessionRepository) DestroyExpired(ctx context.Context, expiredTime time.Time) error {
	var sessions []web.Session

	for _, session := range sr.store {
		if session.UpdatedAt.Before(expiredTime) {
			continue
		}

		sessions = append(sessions, session)
	}

	sr.store = sessions

	return nil
}

func (sr *SessionRepository) Refresh(ctx context.Context, sessionId string) error {
	for i, session := range sr.store {
		if session.SessionId == sessionId {
			sr.store[i].UpdatedAt = time.Now()

			return nil
		}
	}

	return nil
}
