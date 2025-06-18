package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
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

var ErrSessionNotFound = errors.New("session not found")

type SessionRepository interface {

	// Create a new session
	Create(ctx context.Context, session Session) (Session, error)

	// Delete an existing sessionId
	//
	// If the sessionId does not exist, does nothing
	Destroy(ctx context.Context, sessionId string) error

	// Gets the session that matches the sessionId
	//
	// Returns ErrSessionNotFound if the sessionId does not exist.
	Get(ctx context.Context, sessionId string) (Session, error)

	// Deletes any sessions that have an UpdatedAt time older than expiredTime
	DestroyExpired(ctx context.Context, expiredTime time.Time) error

	// Updates the session UpdatedAt time.
	//
	// If the session does not exist does nothing
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

	user, err := ss.AccountRepo.GetById(ctx, session.UserId)

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
