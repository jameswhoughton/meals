package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/internal/auth"
)

type SessionRepositoryContract struct {
	repo func() (auth.SessionRepository, func())
}

func (rc *SessionRepositoryContract) Init(t *testing.T) {
	t.Run("Can create get and destroy a session", func(t *testing.T) {
		repo, closeDown := rc.repo()
		defer closeDown()

		newSession := auth.Session{
			SessionId: "NEW_SESSION_01",
			UserId:    1,
			CreatedAt: time.Now(),
		}

		createdSession, err := repo.Create(newSession)

		if err != nil {
			t.Errorf("Unexpected error when creating a session: %v", err)
		}

		if createdSession.Id != 1 {
			t.Errorf("Expected Id to equal 1, got %d", createdSession.Id)
		}

		if createdSession.SessionId != newSession.SessionId {
			t.Errorf("Expected session ID %s, got %s", newSession.SessionId, createdSession.SessionId)
		}

		if createdSession.UserId != newSession.UserId {
			t.Errorf("Expected user ID %d, got %d", newSession.UserId, createdSession.Id)
		}

		fetchedSession, err := repo.Get(newSession.SessionId)

		if err != nil {
			t.Errorf("Unexpected error when fetching a session: %v", err)
		}

		if fetchedSession.Id != 1 {
			t.Errorf("Expected Id to equal 1, got %d", fetchedSession.Id)
		}

		if fetchedSession.SessionId != newSession.SessionId {
			t.Errorf("Expected session ID %s, got %s", newSession.SessionId, fetchedSession.SessionId)
		}

		if fetchedSession.UserId != newSession.UserId {
			t.Errorf("Expected user ID %d, got %d", newSession.UserId, fetchedSession.Id)
		}

		err = repo.Destroy(newSession.SessionId)

		if err != nil {
			t.Errorf("Unexpected error when destroying a session: %v", err)
		}

		_, err = repo.Get(newSession.SessionId)

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}
	})

	t.Run("Can validate a session", func(t *testing.T) {
		repo, closeDown := rc.repo()
		defer closeDown()

		newSession := auth.Session{
			SessionId: "A001",
			UserId:    1,
		}

		repo.Create(newSession)

		validSession := repo.IsValid(newSession.SessionId)

		if !validSession {
			t.Errorf("Expected session to be valid")
		}

		invalidSession := repo.IsValid("INVALID")

		if invalidSession {
			t.Errorf("Expected session to be invalid")
		}
	})

	t.Run("Can destroy individual user sessions", func(t *testing.T) {
		repo, closeDown := rc.repo()
		defer closeDown()

		repo.Create(auth.Session{
			UserId:    1,
			SessionId: "AA",
		})

		repo.Create(auth.Session{
			UserId:    2,
			SessionId: "BB",
		})

		repo.Create(auth.Session{
			UserId:    1,
			SessionId: "CC",
		})

		err := repo.DestroyByUserId(1)

		if err == nil {
			t.Error("Expected error when destroying user sessions, got none")
		}

		_, err = repo.Get("AA")

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}

		_, err = repo.Get("BB")

		if err != nil {
			t.Errorf("Unexpected error when fetching session that should still exist: %v", err)
		}

		_, err = repo.Get("CC")

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}
	})

	t.Run("Can delete sessions that have exceeded the given lifetime", func(t *testing.T) {
		repo, closeDown := rc.repo()
		defer closeDown()

		repo.Create(auth.Session{
			UserId:    1,
			SessionId: "AA",
			CreatedAt: time.Date(2025, time.March, 5, 12, 30, 0, 0, nil),
		})

		repo.Create(auth.Session{
			UserId:    2,
			SessionId: "BB",
			CreatedAt: time.Date(2025, time.March, 5, 12, 45, 0, 0, nil),
		})

		repo.Create(auth.Session{
			UserId:    3,
			SessionId: "CC",
			CreatedAt: time.Date(2025, time.March, 5, 13, 30, 0, 0, nil),
		})

		err := repo.DestroyExpired(time.Date(2025, time.March, 5, 13, 30, 0, 0, nil))

		if err == nil {
			t.Error("Expected error when destroying user sessions, got none")
		}

		_, err = repo.Get("AA")

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}

		_, err = repo.Get("BB")

		if err != nil {
			t.Errorf("Unexpected error when fetching session that should still exist: %v", err)
		}

		_, err = repo.Get("CC")

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}
	})
}
