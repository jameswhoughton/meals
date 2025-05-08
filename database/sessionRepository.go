package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jameswhoughton/meals/web"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db}
}

func (sr *SessionRepository) Create(ctx context.Context, session web.Session) (web.Session, error) {
	result, err := sr.db.ExecContext(ctx, "INSERT INTO sessions (session_id, user_id, created_at, updated_at) VALUES (?, ?, ?, ?)", session.SessionId, session.UserId, session.CreatedAt, session.UpdatedAt)

	if err != nil {
		return web.Session{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return web.Session{}, err
	}

	session.Id = int(id)

	return session, nil
}

func (sr *SessionRepository) Destroy(ctx context.Context, sessionId string) error {
	_, err := sr.db.ExecContext(ctx, "DELETE FROM sessions WHERE session_id = ?", sessionId)

	if err != nil {
		return fmt.Errorf("could not destroy session: %v", err)
	}

	return nil
}

func (sr *SessionRepository) Get(ctx context.Context, sessionId string) (web.Session, error) {
	var session web.Session

	if err := sr.db.QueryRow("SELECT id, session_id, user_id, created_at, updated_at FROM sessions WHERE session_id = ?", sessionId).Scan(&session.Id, &session.SessionId, &session.UserId, &session.CreatedAt, &session.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return web.Session{}, web.ErrorSessionNotFound{}
		}

		return web.Session{}, fmt.Errorf("error fetching session: %v", err)
	}

	return session, nil
}

func (sr *SessionRepository) DestroyExpired(ctx context.Context, expiredTime time.Time) error {
	_, err := sr.db.ExecContext(ctx, "DELETE FROM sessions WHERE updated_at < ?", expiredTime)

	if err != nil {
		return fmt.Errorf("error removing expired sessions: %v", err)
	}

	return nil
}

func (sr *SessionRepository) Refresh(ctx context.Context, sessionId string) error {
	_, err := sr.db.ExecContext(ctx, "UPDATE sessions SET updated_at = DATETIME('now') WHERE session_id = ?", sessionId)

	if err != nil {
		return fmt.Errorf("error refreshing session: %v", err)
	}

	return nil
}
