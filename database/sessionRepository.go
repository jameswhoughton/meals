package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jameswhoughton/meals/internal/auth"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db}
}

func (sr *SessionRepository) Create(session auth.Session) (auth.Session, error) {
	result, err := sr.db.Exec("INSERT INTO sessions (session_id, user_id, created_at, updated_at) VALUES (?, ?, ?, ?)", session.SessionId, session.UserId, session.CreatedAt, session.UpdatedAt)

	if err != nil {
		return auth.Session{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return auth.Session{}, err
	}

	session.Id = int(id)

	return session, nil
}

func (sr *SessionRepository) Destroy(sessionId string) error {
	_, err := sr.db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionId)

	if err != nil {
		return fmt.Errorf("could not destroy session: %v", err)
	}

	return nil
}

func (sr *SessionRepository) Get(sessionId string, expiredTime time.Time) (auth.Session, error) {
	var session auth.Session

	if err := sr.db.QueryRow("SELECT id, session_id, user_id, created_at, updated_at FROM sessions WHERE session_id = ? AND updated_at > ?", sessionId, expiredTime).Scan(&session.Id, &session.SessionId, &session.UserId, &session.CreatedAt, &session.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return auth.Session{}, auth.ErrorSessionNotFound{}
		}

		return auth.Session{}, fmt.Errorf("error fetching session: %v", err)
	}

	return session, nil
}

func (sr *SessionRepository) DestroyExpired(expiredTime time.Time) error {
	_, err := sr.db.Exec("DELETE FROM sessions WHERE updated_at < ?", expiredTime)

	if err != nil {
		return fmt.Errorf("error removing expired sessions: %v", err)
	}

	return nil
}
