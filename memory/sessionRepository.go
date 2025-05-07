package memory

import (
	"context"
	"time"

	"github.com/jameswhoughton/meals/web"
)

type SessionRepository struct {
	Store []web.Session
}

func (sr *SessionRepository) Create(ctx context.Context, session web.Session) (web.Session, error) {
	session.Id = len(sr.Store) + 1

	sr.Store = append(sr.Store, session)

	return session, nil
}

func (sr *SessionRepository) Destroy(ctx context.Context, sessionId string) error {
	var index *int

	for i, session := range sr.Store {
		if session.SessionId == sessionId {
			index = &i
			break
		}
	}

	if index == nil {
		return web.ErrorSessionNotFound{}
	}

	existingStore := sr.Store

	sr.Store = make([]web.Session, len(existingStore)-1)

	sr.Store = append(sr.Store, existingStore[:*index]...)
	sr.Store = append(sr.Store, existingStore[*index+1:]...)

	return nil
}

func (sr *SessionRepository) Get(ctx context.Context, sessionId string) (web.Session, error) {
	for i, session := range sr.Store {
		if session.SessionId == sessionId {
			sr.Store[i] = session

			return sr.Store[i], nil
		}
	}

	return web.Session{}, web.ErrorSessionNotFound{}
}

func (sr *SessionRepository) DestroyExpired(ctx context.Context, expiredTime time.Time) error {
	var sessions []web.Session

	for _, session := range sr.Store {
		if session.UpdatedAt.Before(expiredTime) {
			continue
		}

		sessions = append(sessions, session)
	}

	sr.Store = sessions

	return nil
}

func (sr *SessionRepository) Refresh(ctx context.Context, sessionId string) error {
	for i, session := range sr.Store {
		if session.SessionId == sessionId {
			sr.Store[i].UpdatedAt = time.Now()

			return nil
		}
	}

	return web.ErrorSessionNotFound{}
}
