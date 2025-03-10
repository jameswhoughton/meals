package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jameswhoughton/meals/internal/auth"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return SessionRepository{db}
}

func (sr *SessionRepository) Add(session auth.Session) (auth.Session, error) {
	result, err := sr.db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", session.SessionId, session.UserId)

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

func (sr *SessionRepository) IsValid(sessionId string) bool {
	row := sr.db.QueryRow("SELECT id FROM sessions WHERE session_id = ?", sessionId)

	if err := row.Scan(); errors.Is(err, sql.ErrNoRows) {
		return false
	}

	return true
}

func (sr *SessionRepository) GetUser(sessionId string) (auth.User, error) {
	var user auth.User

	if err := sr.db.QueryRow("SELECT u.id, u.email, u.name FROM sessions s LEFT JOIN users u ON s.user_id = u.id WHERE session_id = ?", sessionId).Scan(&user.Id, &user.Email, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return auth.User{}, fmt.Errorf("session ID invalid")
		}

		return auth.User{}, fmt.Errorf("error fetching user: %v", err)
	}

	return user, nil
}

func (sr *SessionRepository) DestroyByUserId(userId int) error {
	_, err := sr.db.Exec("DELETE FROM sessions WHERE user_id = ?", userId)

	if err != nil {
		return fmt.Errorf("error removing old user sessions (user %d): %v", userId, err)
	}

	return nil
}

func (sr *SessionRepository) DestroyExpired(lifetime int) error {
	_, err := sr.db.Exec("DELETE FROM sessions WHERE created_at - ? < NOW()", lifetime)

	if err != nil {
		return fmt.Errorf("error removing expired sessions: %v", err)
	}

	return nil
}
