package usecase

import (
	"auth/internal/domain"
	"auth/internal/repository"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
)

type Claims struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

type AuthUseCase interface {
	Register(ctx context.Context, firstName, lastName, email, phone, password string, role domain.Role) (*domain.User, TokenPair, error)
	Login(ctx context.Context, email, password string, ua, ip string) (TokenPair, error)
	SendOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone, code, ua, ip string) (TokenPair, error)
	ParseToken(ctx context.Context, token string) (*Claims, error)
	RefreshTokens(ctx context.Context, refreshToken, ua, ip string) (TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID string) error
}

type authUseCase struct {
	users    repository.UserRepository
	sessions repository.SessionsRepository
	otp      repository.OTPRepository

	accessSecret  string
	refreshSecret string
	issuer        string
	audience      string
	accessTTL     time.Duration
	refreshTTL    time.Duration

	otpEnabled     bool
	otpTTL         time.Duration
	otpLength      int
	maxOTPAttempts int
}

type Config struct {
	AccessSecret   string
	RefreshSecret  string
	Issuer         string
	Audience       string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	OTPEnabled     bool
	OTPTTL         time.Duration
	OTPLength      int
	MaxOTPAttempts int
}

func NewAuthUseCase(
	users repository.UserRepository,
	sessions repository.SessionsRepository,
	otp repository.OTPRepository,
	cfg Config,
) AuthUseCase {
	if cfg.RefreshSecret == "" {
		cfg.RefreshSecret = cfg.AccessSecret
	}
	if cfg.MaxOTPAttempts <= 0 {
		cfg.MaxOTPAttempts = 5
	}
	return &authUseCase{
		users:    users,
		sessions: sessions,
		otp:      otp,

		accessSecret:  cfg.AccessSecret,
		refreshSecret: cfg.RefreshSecret,
		issuer:        cfg.Issuer,
		audience:      cfg.Audience,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,

		otpEnabled:     cfg.OTPEnabled,
		otpTTL:         cfg.OTPTTL,
		otpLength:      cfg.OTPLength,
		maxOTPAttempts: cfg.MaxOTPAttempts,
	}
}

func (uc *authUseCase) Register(ctx context.Context, firstName, lastName, email, phone, password string, role domain.Role) (*domain.User, TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, TokenPair{}, fmt.Errorf("hash password: %w", err)
	}
	u := &domain.User{
		ID:        uuid.NewString(),
		Email:     email,
		Phone:     phone,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}
	if err := uc.users.Create(ctx, u, string(hash)); err != nil {
		return nil, TokenPair{}, err
	}
	tp, err := uc.issueTokens(ctx, u.ID, u.Role, "", "")
	return u, tp, err
}

func (uc *authUseCase) Login(ctx context.Context, email, password, ua, ip string) (TokenPair, error) {
	u, passHash, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return TokenPair{}, ErrUserNotFound
		}
		return TokenPair{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(password)); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	return uc.issueTokens(ctx, u.ID, u.Role, ua, ip)
}

func (uc *authUseCase) SendOTP(ctx context.Context, phone string) error {
	if !uc.otpEnabled {
		return nil
	}
	code := uc.generateNumericCode(uc.otpLength)
	codeHash := uc.hashOpaque(code, uc.refreshSecret) // хеш не храним в открытом виде
	if err := uc.otp.Upsert(ctx, phone, codeHash, repository.OTPPurposeLogin, time.Now().Add(uc.otpTTL)); err != nil {
		return err
	}
	// тут отправка через SMS-провайдера (пока можно логировать/заглушка)
	_ = code
	return nil
}

func (uc *authUseCase) VerifyOTP(ctx context.Context, phone, code, ua, ip string) (TokenPair, error) {
	ch, _, attempts, err := uc.otp.GetActive(ctx, phone, repository.OTPPurposeLogin)
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	if attempts >= uc.maxOTPAttempts {
		return TokenPair{}, ErrInvalidCredentials
	}
	codeHash := uc.hashOpaque(code, uc.refreshSecret)
	if codeHash != ch {
		_ = uc.otp.IncAttempt(ctx, phone, repository.OTPPurposeLogin)
		return TokenPair{}, ErrInvalidCredentials
	}
	_ = uc.otp.Consume(ctx, phone, repository.OTPPurposeLogin)

	// если пользователя нет — создаём пустого (минимум)
	u, err := uc.users.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			u = &domain.User{
				ID:    uuid.NewString(),
				Phone: phone,
				Email: "",
				Role:  domain.StudentRole,
			}
			// создаём с пустым email/паролем (пароль пуст, вход только по OTP)
			if err := uc.users.Create(ctx, u, ""); err != nil {
				return TokenPair{}, err
			}
		} else {
			return TokenPair{}, err
		}
	}
	return uc.issueTokens(ctx, u.ID, u.Role, ua, ip)
}

func (uc *authUseCase) ParseToken(ctx context.Context, token string) (*Claims, error) {
	return uc.parseWithSecret(token, uc.accessSecret)
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken, ua, ip string) (TokenPair, error) {
	claims, err := uc.parseWithSecret(refreshToken, uc.refreshSecret)
	if err != nil {
		return TokenPair{}, ErrInvalidRefresh
	}
	// проверяем refresh по сессии
	hash := uc.hashOpaque(refreshToken, uc.refreshSecret)
	sessID, userID, exp, revokedAt, err := uc.sessions.GetByHash(ctx, hash)
	if err != nil || revokedAt != nil || time.Now().After(exp) || userID != claims.UserID {
		return TokenPair{}, ErrInvalidRefresh
	}
	// ревокация старого refresh
	_ = uc.sessions.Revoke(ctx, sessID)
	// выпуск новой пары
	return uc.issueTokens(ctx, claims.UserID, claims.Role, ua, ip)
}

func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	hash := uc.hashOpaque(refreshToken, uc.refreshSecret)
	// найдём сессию и пометим revoked
	sessID, _, _, _, err := uc.sessions.GetByHash(ctx, hash)
	if err != nil {
		return ErrInvalidRefresh
	}
	return uc.sessions.Revoke(ctx, sessID)
}

func (uc *authUseCase) LogoutAll(ctx context.Context, userID string) error {
	return uc.sessions.RevokeAllByUser(ctx, userID)
}

// --- helpers

func (uc *authUseCase) issueTokens(ctx context.Context, userID string, role domain.Role, ua, ip string) (TokenPair, error) {
	access, err := uc.sign(userID, role, uc.accessSecret, uc.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := uc.sign(userID, role, uc.refreshSecret, uc.refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	// сохраняем refresh-сессию (по хешу)
	hash := uc.hashOpaque(refresh, uc.refreshSecret)
	if err := uc.sessions.Create(ctx, uuid.NewString(), userID, hash, time.Now().Add(uc.refreshTTL), ua, ip); err != nil {
		return TokenPair{}, fmt.Errorf("create session: %w", err)
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		AccessTTL:    uc.accessTTL,
		RefreshTTL:   uc.refreshTTL,
	}, nil
}

func (uc *authUseCase) sign(userID string, role domain.Role, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    uc.issuer,
			Audience:  []string{uc.audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func (uc *authUseCase) parseWithSecret(token, secret string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := t.Claims.(*Claims); ok && t.Valid {
		return c, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func (uc *authUseCase) generateNumericCode(n int) string {
	// простая заглушка: в реале используйте криптонадёжный генератор
	if n <= 0 {
		n = 6
	}
	return "123456"
}

func (uc *authUseCase) hashOpaque(s, pepper string) string {
	h := sha256.Sum256([]byte(pepper + ":" + s))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
