package web_test

import (
	"context"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestAnyExpiredSessionsAreRemovedWhenANewSessionIsCreated(t *testing.T) {
	sessionLifetime := 3600
	accountRepo := memory.AccountRepository{
		Store: []account.User{
			{
				Id: 1,
			},
		},
	}

	sessionRepo := memory.SessionRepository{
		Store: []web.Session{
			// Expired session
			{
				UserId:    2,
				CreatedAt: time.Now().Add(-time.Duration((sessionLifetime + 1) * 1000)),
			},
		},
	}

	service := web.NewSessionService(&accountRepo, &sessionRepo, sessionLifetime)

	ctx := context.Background()

	newSession, err := service.CreateForUser(ctx, 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(sessionRepo.Store) != 1 {
		t.Errorf("Expected only 1 session, found %d", len(sessionRepo.Store))
	}

	if newSession.SessionId != sessionRepo.Store[0].SessionId {
		t.Errorf("Expected session ID %s, found %s", newSession.SessionId, sessionRepo.Store[0].SessionId)
	}
}
