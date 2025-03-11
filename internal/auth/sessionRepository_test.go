package auth_test

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/auth"
	"github.com/jameswhoughton/meals/memory"
)

type SessionRepositoryContract struct {
	repo func() (auth.SessionRepository, func())
}

func (rc *SessionRepositoryContract) Test(t *testing.T) {
	t.Run("Can create get and destroy a session", func(t *testing.T) {
		repo, closeDown := rc.repo()
		defer closeDown()

		newSession := auth.Session{
			SessionId: "NEW_SESSION_01",
			UserId:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now().Add(time.Second * -30),
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

		fetchedSession, err := repo.Get(newSession.SessionId, time.Now().Add(-time.Second*100))

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

		if fetchedSession.UpdatedAt == createdSession.UpdatedAt {
			t.Error("updated_at should update when using Get, dates match")
		}

		err = repo.Destroy(newSession.SessionId)

		if err != nil {
			t.Errorf("Unexpected error when destroying a session: %v", err)
		}

		_, err = repo.Get(newSession.SessionId, time.Now().Add(-time.Second*100))

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}
	})

	t.Run("Cannot get a session with a valid ID that has expired", func(t *testing.T) {
		repo, closeDown := rc.repo()
		defer closeDown()

		repo.Create(auth.Session{
			UserId:    1,
			SessionId: "AA",
			UpdatedAt: time.Date(2025, time.March, 5, 12, 30, 0, 0, time.UTC),
		})

		_, err := repo.Get("AA", time.Date(2025, time.March, 5, 13, 0, 0, 0, time.UTC))

		if err == nil {
			t.Error("Expected error, got nil")
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
			UpdatedAt: time.Date(2025, time.March, 5, 12, 30, 0, 0, time.UTC),
		})

		repo.Create(auth.Session{
			UserId:    2,
			SessionId: "BB",
			UpdatedAt: time.Date(2025, time.March, 5, 12, 45, 0, 0, time.UTC),
		})

		repo.Create(auth.Session{
			UserId:    3,
			SessionId: "CC",
			UpdatedAt: time.Date(2025, time.March, 5, 13, 31, 0, 0, time.UTC),
		})

		expiredTime := time.Date(2025, time.March, 5, 13, 30, 0, 0, time.UTC)

		err := repo.DestroyExpired(expiredTime)

		if err != nil {
			t.Errorf("Unexpected error when destroying user sessions: %v", err)
		}

		_, err = repo.Get("AA", expiredTime)

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}

		_, err = repo.Get("CC", expiredTime)

		if err != nil {
			t.Errorf("Unexpected error when fetching session that should still exist: %v", err)
		}

		_, err = repo.Get("BB", expiredTime)

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, auth.ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", auth.ErrorSessionNotFound{}, err, err)
		}
	})
}

func TestDatabaseSessionService(t *testing.T) {
	init := func() (auth.SessionRepository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		closeDown := func() {
			os.Remove("meals.db")
		}
		return database.NewSessionRepository(conn), closeDown
	}

	contract := SessionRepositoryContract{
		init,
	}

	contract.Test(t)

}

func TestMemorySessionService(t *testing.T) {
	init := func() (auth.SessionRepository, func()) {
		return &memory.SessionRepository{
			Store: []auth.Session{},
		}, func() {}
	}

	contract := SessionRepositoryContract{
		repo: init,
	}

	contract.Test(t)

}
