package web_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/memory"
	"github.com/jameswhoughton/meals/web"
)

func TestAnyExpiredSessionsAreRemovedWhenANewSessionIsCreated(t *testing.T) {
	sessionLifetime := 3600
	accountRepo := memory.AccountRepository{}

	ctx := context.Background()

	accountRepo.Create(ctx, account.User{Id: 1})

	sessionRepo := memory.SessionRepository{}

	expiredSession, err := sessionRepo.Create(ctx, web.Session{
		UserId:    2,
		CreatedAt: time.Now().Add(-time.Duration((sessionLifetime + 1) * 1000)),
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	service := web.NewSessionService(&accountRepo, &sessionRepo, sessionLifetime)

	newSession, err := service.CreateForUser(ctx, 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, err = sessionRepo.Get(ctx, expiredSession.SessionId)

	if !errors.Is(err, web.ErrSessionNotFound) {
		t.Error("expired session still exists")
	}

	check, err := sessionRepo.Get(ctx, newSession.SessionId)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if newSession.SessionId != check.SessionId {
		t.Errorf("Expected session ID %s, found %s", newSession.SessionId, check.SessionId)
	}
}
