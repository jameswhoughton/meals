package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jameswhoughton/meals/internal/account"
)

type Session struct {
	Id        int
	SessionId string
	UserId    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ErrorSessionNotFound struct {
}

func (e ErrorSessionNotFound) Error() string {
	return "Session not found"
}

type SessionRepository interface {
	Create(ctx context.Context, session Session) (Session, error)
	Destroy(ctx context.Context, sessionId string) error
	Get(ctx context.Context, sessionId string) (Session, error)
	DestroyExpired(ctx context.Context, expiredTime time.Time) error
	Refresh(ctx context.Context, sessionId string) error
}

func NewSessionService(accountRepo account.Repository, sessionRepo SessionRepository, sessionLifetime int) *SessionService {
	return &SessionService{
		SessionRepo:     sessionRepo,
		AccountRepo:     accountRepo,
		SessionLifetime: sessionLifetime,
	}
}

type SessionService struct {
	SessionRepo     SessionRepository
	AccountRepo     account.Repository
	SessionLifetime int
}

func (ss *SessionService) CreateForUser(ctx context.Context, userId int) (Session, error) {
	createdAt := time.Now()

	session := Session{
		SessionId: GenerateSessionId(),
		UserId:    userId,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	// Remove any expired sessions
	ss.SessionRepo.DestroyExpired(ctx, time.Now().Add(-time.Duration(ss.SessionLifetime)*time.Second))

	return ss.SessionRepo.Create(ctx, session)
}

func (ss *SessionService) UserFromSession(ctx context.Context, sessionId string) (account.User, error) {
	session, err := ss.SessionRepo.Get(ctx, sessionId)

	if err != nil {
		return account.User{}, fmt.Errorf("error fetching session: %w", err)
	}

	expiredTime := time.Now().Add(-time.Duration(ss.SessionLifetime) * time.Second).UTC()

	if session.UpdatedAt.UTC().Before(expiredTime) {
		ss.SessionRepo.Destroy(ctx, sessionId)

		return account.User{}, fmt.Errorf("Session has expired")
	}

	user, err := ss.AccountRepo.Get(ctx, account.GetForm{Id: &session.UserId})

	if err != nil {
		return account.User{}, fmt.Errorf("error fetching user: %w", err)
	}

	ss.SessionRepo.Refresh(ctx, sessionId)

	return user, nil
}

func GenerateSessionId() string {
	key := make([]byte, 32)
	rand.Read(key)

	return base64.StdEncoding.EncodeToString(key)
}
