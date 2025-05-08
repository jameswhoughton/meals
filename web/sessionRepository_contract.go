package web

import (
	"context"
	"errors"
	"testing"
	"time"
)

type SessionRepositoryContract struct {
	Repo func() (SessionRepository, func())
}

func (rc *SessionRepositoryContract) Test(t *testing.T) {
	t.Run("Can create get and destroy a session", func(t *testing.T) {
		repo, closeDown := rc.Repo()
		defer closeDown()

		ctx := context.Background()

		newSession := Session{
			SessionId: "NEW_SESSION_01",
			UserId:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now().Add(time.Second * -30),
		}

		createdSession, err := repo.Create(ctx, newSession)

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

		fetchedSession, err := repo.Get(ctx, newSession.SessionId)

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

		err = repo.Destroy(ctx, newSession.SessionId)

		if err != nil {
			t.Errorf("Unexpected error when destroying a session: %v", err)
		}

		_, err = repo.Get(ctx, newSession.SessionId)

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorSessionNotFound{}, err, err)
		}
	})

	t.Run("Can delete sessions that have exceeded the given lifetime", func(t *testing.T) {
		repo, closeDown := rc.Repo()
		defer closeDown()

		ctx := context.Background()

		repo.Create(ctx, Session{
			UserId:    1,
			SessionId: "AA",
			UpdatedAt: time.Date(2025, time.March, 5, 12, 30, 0, 0, time.UTC),
		})

		repo.Create(ctx, Session{
			UserId:    2,
			SessionId: "BB",
			UpdatedAt: time.Date(2025, time.March, 5, 12, 45, 0, 0, time.UTC),
		})

		repo.Create(ctx, Session{
			UserId:    3,
			SessionId: "CC",
			UpdatedAt: time.Date(2025, time.March, 5, 13, 31, 0, 0, time.UTC),
		})

		expiredTime := time.Date(2025, time.March, 5, 13, 30, 0, 0, time.UTC)

		err := repo.DestroyExpired(ctx, expiredTime)

		if err != nil {
			t.Errorf("Unexpected error when destroying user sessions: %v", err)
		}

		_, err = repo.Get(ctx, "AA")

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorSessionNotFound{}, err, err)
		}

		_, err = repo.Get(ctx, "CC")

		if err != nil {
			t.Errorf("Unexpected error when fetching session that should still exist: %v", err)
		}

		_, err = repo.Get(ctx, "BB")

		if err == nil {
			t.Error("Expected error when fetching destroyed session, got none")
		}

		if !errors.Is(err, ErrorSessionNotFound{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorSessionNotFound{}, err, err)
		}
	})
}

// Test that Refresh updates the UpdatedAt field
