package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ranji/clothing-erp/internal/apperr"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo            *Repository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(repo *Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: []byte(jwtSecret), accessTokenTTL: accessTTL, refreshTokenTTL: refreshTTL}
}

// NeedsBootstrap reports whether no users exist yet, in which case the
// first /auth/register call is allowed without authentication (and is
// forced to ADMIN) to break the chicken-and-egg problem of needing an admin
// to create the first admin.
func (s *Service) NeedsBootstrap(ctx context.Context) (bool, error) {
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return false, apperr.Internal("count users", err)
	}
	return count == 0, nil
}

// normalizeEmail trims and lowercases so login isn't sensitive to
// capitalization or accidental whitespace from copy/paste or a mobile
// keyboard's auto-capitalize-first-letter behavior -- verified live: a
// production deploy's admin login failed with 422 purely from this, even
// though the stored email and the user's typed email were "the same" to a
// human.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) Register(ctx context.Context, in RegisterInput, forceAdmin bool) (User, error) {
	in.Email = normalizeEmail(in.Email)
	if in.Email == "" || in.Password == "" || in.Name == "" {
		return User{}, apperr.Validation("email, password, and name are required")
	}
	if len(in.Password) < 8 {
		return User{}, apperr.Validation("password must be at least 8 characters")
	}

	role := in.Role
	if forceAdmin {
		role = RoleAdmin
	}
	if !role.Valid() {
		return User{}, apperr.Validation("role must be one of ADMIN, SALES, PRODUCTION, FINANCE, ACCOUNTING")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, apperr.Internal("hash password", err)
	}

	user, err := s.repo.CreateUser(ctx, in.Email, string(hash), in.Name, role)
	if err != nil {
		return User{}, apperr.Wrapf(err, "create user (email may already be registered)")
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (TokenPair, User, error) {
	user, err := s.repo.GetByEmail(ctx, normalizeEmail(in.Email))
	if errors.Is(err, pgx.ErrNoRows) {
		return TokenPair{}, User{}, apperr.Validation("invalid email or password")
	}
	if err != nil {
		return TokenPair{}, User{}, apperr.Internal("load user", err)
	}
	if !user.IsActive {
		return TokenPair{}, User{}, apperr.Forbidden("account is deactivated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return TokenPair{}, User{}, apperr.Validation("invalid email or password")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return TokenPair{}, User{}, err
	}
	return pair, user, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, User, error) {
	hash := hashToken(refreshToken)
	stored, err := s.repo.GetActiveRefreshToken(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return TokenPair{}, User{}, apperr.Forbidden("refresh token is invalid, expired, or already used")
	}
	if err != nil {
		return TokenPair{}, User{}, apperr.Internal("load refresh token", err)
	}

	user, err := s.repo.GetByID(ctx, stored.UserID)
	if err != nil {
		return TokenPair{}, User{}, apperr.Internal("load user", err)
	}
	if !user.IsActive {
		return TokenPair{}, User{}, apperr.Forbidden("account is deactivated")
	}

	// Rotate: the presented token is single-use, so a stolen-and-replayed
	// refresh token is invalidated the moment the legitimate client uses it.
	if err := s.repo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return TokenPair{}, User{}, apperr.Internal("revoke used refresh token", err)
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return TokenPair{}, User{}, err
	}
	return pair, user, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, apperr.Internal("list users", err)
	}
	return users, nil
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, apperr.NotFound("user not found")
	}
	return user, err
}

func (s *Service) issueTokenPair(ctx context.Context, user User) (TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTokenTTL)

	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  string(user.Role),
		"iat":   now.Unix(),
		"exp":   accessExp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, apperr.Internal("sign access token", err)
	}

	refreshPlain, err := randomToken()
	if err != nil {
		return TokenPair{}, apperr.Internal("generate refresh token", err)
	}
	refreshExp := now.Add(s.refreshTokenTTL)
	if err := s.repo.InsertRefreshToken(ctx, user.ID, hashToken(refreshPlain), refreshExp); err != nil {
		return TokenPair{}, apperr.Internal("store refresh token", err)
	}

	return TokenPair{
		AccessToken: accessToken, RefreshToken: refreshPlain,
		AccessExpiresAt: accessExp, RefreshExpiresAt: refreshExp,
	}, nil
}

// ValidateAccessToken verifies signature and expiry and extracts claims. It
// does not hit the database -- role/active-state changes take effect on the
// next login or token refresh, not mid-session, which is the standard JWT
// tradeoff for statelessness.
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, apperr.Forbidden("invalid or expired token")
	}

	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperr.Forbidden("invalid token claims")
	}

	sub, _ := claimsMap["sub"].(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, apperr.Forbidden("invalid token subject")
	}
	email, _ := claimsMap["email"].(string)
	role, _ := claimsMap["role"].(string)

	return &Claims{UserID: userID, Email: email, Role: Role(role)}, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
