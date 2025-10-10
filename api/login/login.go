package login

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xredis"
	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"

	projectconf "github.com/revenkroz/vite-ssr-golang/conf"
	"github.com/revenkroz/vite-ssr-golang/dao"
)

const (
	httpRequestTimeout  = 10 * time.Second
	emailCodeTTL        = 10 * time.Minute
	emailSendCooldown   = 60 * time.Second
	defaultEmailFrom    = "onboarding@resend.dev"
	defaultEmailSubject = "Your verification code"
)

var (
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type ReqAuthLoginGoogle struct {
	Code        string `json:"code"`
	AccessToken string `json:"access_token"`
}

type ReqAuthEmail struct {
	Email string `json:"email" binding:"required,email"`
}

type ReqAuthVerify struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type RespAuthEmail struct {
	Message string `json:"message"`
}

type RespAuthLogin struct {
	SessionToken string      `json:"session_token"`
	User         AuthUserDTO `json:"user"`
}

type AuthUserDTO struct {
	ID            string `json:"id"`
	ExternalID    string `json:"external_id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	AvatarURL     string `json:"avatar_url"`
	Channel       string `json:"channel"`
	AppID         string `json:"appid"`
	ProjectUserID int    `json:"project_user_id"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IdToken     string `json:"id_token"`
}

type projectUserInput struct {
	AppID      string
	Email      string
	Name       string
	AvatarURL  string
	Channel    string
	InviteCode string
}

type ReqExample struct{}

type RespExample struct{}

func AuthLoginGoogle(ctx *gin.Context, req ReqAuthLoginGoogle) (*RespAuthLogin, error) {
	requestCtx := ctx.Request.Context()
	accessToken := strings.TrimSpace(req.AccessToken)

	cfg := projectconf.Get()
	if cfg == nil {
		return nil, errors.New("config not initialized")
	}

	appID := strings.TrimSpace(cfg.Project.AppID)
	inviteCode, _ := ctx.Cookie("invite_code")

	if accessToken == "" && strings.TrimSpace(req.Code) != "" {
		token, err := exchangeCodeForAccessToken(requestCtx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("exchange-code: %w", err)
		}
		accessToken = token
	}

	if accessToken == "" {
		return nil, errors.New("missing access token")
	}

	info, err := fetchGoogleUser(requestCtx, accessToken)
	if err != nil {
		return nil, err
	}

	projectUser, err := ensureProjectUser(projectUserInput{
		AppID:      appID,
		Email:      info.Email,
		Name:       info.Name,
		AvatarURL:  info.Picture,
		Channel:    "google",
		InviteCode: inviteCode,
	})
	if err != nil {
		return nil, err
	}

	user := AuthUserDTO{
		ID:            fmt.Sprintf("%d", projectUser.Id),
		ExternalID:    info.Sub,
		Email:         projectUser.Email,
		Name:          projectUser.UserName,
		Provider:      "google",
		AvatarURL:     projectUser.AvatarURL,
		Channel:       projectUser.Channel,
		AppID:         projectUser.AppID,
		ProjectUserID: projectUser.Id,
	}

	token := buildMockToken(user)
	setSessionCookie(ctx, token)

	return &RespAuthLogin{
		SessionToken: token,
		User:         user,
	}, nil
}

func AuthRequestEmailCode(ctx *gin.Context, req ReqAuthEmail) (*RespAuthEmail, error) {
	requestCtx := ctx.Request.Context()
	email := normalizeEmail(req.Email)
	if email == "" {
		return nil, errors.New("invalid email")
	}

	if err := enforceEmailCooldown(requestCtx, email); err != nil {
		return nil, err
	}

	code, err := generateVerificationCode(requestCtx, email)
	if err != nil {
		return nil, err
	}

	if err := sendVerificationEmail(requestCtx, email, code); err != nil {
		clearVerificationState(requestCtx, email)
		return nil, err
	}

	return &RespAuthEmail{Message: "verification code sent"}, nil
}

func AuthVerifyEmailCode(ctx *gin.Context, req ReqAuthVerify) (*RespAuthLogin, error) {
	requestCtx := ctx.Request.Context()
	email := normalizeEmail(req.Email)
	if email == "" {
		return nil, errors.New("invalid email")
	}

	if err := verifyEmailCode(requestCtx, email, strings.TrimSpace(req.Code)); err != nil {
		return nil, err
	}

	cfg := projectconf.Get()
	if cfg == nil {
		return nil, errors.New("config not initialized")
	}

	appID := strings.TrimSpace(cfg.Project.AppID)
	inviteCode, _ := ctx.Cookie("invite_code")
	displayName := deriveNameFromEmail(email)

	projectUser, err := ensureProjectUser(projectUserInput{
		AppID:      appID,
		Email:      email,
		Name:       displayName,
		Channel:    "email",
		InviteCode: inviteCode,
	})
	if err != nil {
		return nil, err
	}

	user := AuthUserDTO{
		ID:            fmt.Sprintf("%d", projectUser.Id),
		ExternalID:    "",
		Email:         projectUser.Email,
		Name:          projectUser.UserName,
		Provider:      "email",
		AvatarURL:     projectUser.AvatarURL,
		Channel:       projectUser.Channel,
		AppID:         projectUser.AppID,
		ProjectUserID: projectUser.Id,
	}

	token := buildMockToken(user)
	setSessionCookie(ctx, token)

	return &RespAuthLogin{
		SessionToken: token,
		User:         user,
	}, nil
}

func AuthLogout(ctx *gin.Context, _ ReqExample) (*RespAuthEmail, error) {
	ctx.SetCookie("session_token", "", -1, "/", "", false, true)
	return &RespAuthEmail{Message: "logged-out"}, nil
}

func AuthSession(ctx *gin.Context, _ ReqExample) (*RespAuthLogin, error) {
	token, err := ctx.Cookie("session_token")
	if err != nil || token == "" {
		return &RespAuthLogin{}, nil
	}

	user, err := parseMockToken(token)
	if err != nil {
		return &RespAuthLogin{}, nil
	}

	return &RespAuthLogin{
		SessionToken: token,
		User:         *user,
	}, nil
}

func fetchGoogleUser(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	resp, err := xrequest.New().
		WithContext(ctx).
		SetTimeout(httpRequestTimeout).
		SetHeader("Authorization", "Bearer "+accessToken).
		Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google-userinfo: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("google-userinfo: unexpected status %d", resp.StatusCode())
	}

	var info googleUserInfo
	if err := resp.Scan(&info); err != nil {
		return nil, fmt.Errorf("google-userinfo: decode: %w", err)
	}

	if info.Email == "" || info.Sub == "" {
		return nil, errors.New("google-userinfo: missing required fields")
	}

	return &info, nil
}

func exchangeCodeForAccessToken(ctx context.Context, code string) (string, error) {
	cfg := projectconf.Get()
	if cfg == nil {
		return "", errors.New("config not initialized")
	}

	clientID := strings.TrimSpace(cfg.GoogleOAuth.ClientID)
	clientSecret := strings.TrimSpace(cfg.GoogleOAuth.ClientSecret)
	redirectURI := strings.TrimSpace(cfg.GoogleOAuth.RedirectURI)

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		return "", errors.New("google oauth environment not configured")
	}

	resp, err := xrequest.New().
		WithContext(ctx).
		SetTimeout(httpRequestTimeout).
		SetFormUrlEncode(map[string]string{
			"code":          code,
			"client_id":     clientID,
			"client_secret": clientSecret,
			"redirect_uri":  redirectURI,
			"grant_type":    "authorization_code",
		}).
		Post("https://oauth2.googleapis.com/token")
	if err != nil {
		return "", fmt.Errorf("google-token: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("google-token: unexpected status %d", resp.StatusCode())
	}

	var data googleTokenResponse
	if err := resp.Scan(&data); err != nil {
		return "", fmt.Errorf("google-token: decode: %w", err)
	}

	if data.AccessToken == "" {
		return "", errors.New("google-token: empty access_token")
	}

	return data.AccessToken, nil
}

func ensureProjectUser(input projectUserInput) (*dao.ProjectUserRecord, error) {
	email := normalizeEmail(input.Email)
	if email == "" {
		return nil, errors.New("missing email")
	}

	appID := strings.TrimSpace(input.AppID)
	displayName := strings.TrimSpace(input.Name)
	if displayName == "" {
		displayName = deriveNameFromEmail(email)
	}
	inviteCode := strings.TrimSpace(input.InviteCode)
	refUID := 0

	if refUID == 0 && inviteCode != "" {
		if inviter, err := dao.GetProjectUserByInviteCode(appID, inviteCode); err == nil {
			refUID = inviter.Id
		} else if !errors.Is(err, xdb.ErrNotFound) {
			return nil, err
		}
	}

	record, err := dao.GetProjectUserByAppAndEmail(appID, email)
	if err != nil {
		if !errors.Is(err, xdb.ErrNotFound) {
			return nil, err
		}

		newRecord := &dao.ProjectUserRecord{
			AppID:      appID,
			RefUID:     refUID,
			InviteCode: inviteCode,
			Email:      email,
			Password:   "",
			UserName:   displayName,
			AvatarURL:  input.AvatarURL,
			Channel:    input.Channel,
		}
		lastID, err := dao.CreateProjectUser(newRecord)
		if err != nil {
			return nil, err
		}
		newRecord.Id = int(lastID)
		return newRecord, nil
	}

	updates := xdb.Record{}
	if displayName != "" && displayName != record.UserName {
		updates["user_name"] = displayName
		record.UserName = displayName
	}
	if input.AvatarURL != "" && input.AvatarURL != record.AvatarURL {
		updates["avatar_url"] = input.AvatarURL
		record.AvatarURL = input.AvatarURL
	}
	if appID != "" && record.AppID != appID {
		updates["appid"] = appID
		record.AppID = appID
	}
	if input.Channel != "" && record.Channel != input.Channel {
		updates["channel"] = input.Channel
		record.Channel = input.Channel
	}
	if inviteCode != "" && record.InviteCode == "" && record.RefUID == 0 {
		record.InviteCode = inviteCode
	}
	if refUID > 0 && record.RefUID == 0 {
		updates["ref_uid"] = refUID
		record.RefUID = refUID
	}

	if len(updates) > 0 {
		if err := dao.UpdateProjectUserByID(record.Id, updates); err != nil {
			return nil, err
		}
	}

	return record, nil
}

func generateVerificationCode(ctx context.Context, email string) (string, error) {
	code := fmt.Sprintf("%06d", rng.Intn(1000000))
	client := xredis.Get()
	if client == nil {
		return "", errors.New("redis not configured")
	}

	if err := client.Set(ctx, emailCodeRedisKey(email), code, emailCodeTTL).Err(); err != nil {
		return "", err
	}

	return code, nil
}

func verifyEmailCode(ctx context.Context, email string, code string) error {
	if code == "" {
		return errors.New("verification code required")
	}
	client := xredis.Get()
	if client == nil {
		return errors.New("redis not configured")
	}

	stored, err := client.Get(ctx, emailCodeRedisKey(email)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("verification code expired or invalid")
		}
		return err
	}

	if stored != code {
		return errors.New("verification code mismatch")
	}

	client.Del(ctx, emailCodeRedisKey(email))
	client.Del(ctx, emailCooldownRedisKey(email))
	return nil
}

func enforceEmailCooldown(ctx context.Context, email string) error {
	client := xredis.Get()
	if client == nil {
		return errors.New("redis not configured")
	}

	key := emailCooldownRedisKey(email)
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if ttl > 0 {
		return fmt.Errorf("please wait %d seconds before requesting another code", int(ttl.Seconds())+1)
	}

	return client.Set(ctx, key, "1", emailSendCooldown).Err()
}

func clearVerificationState(ctx context.Context, email string) {
	client := xredis.Get()
	if client == nil {
		return
	}
	client.Del(ctx, emailCodeRedisKey(email))
	client.Del(ctx, emailCooldownRedisKey(email))
}

func sendVerificationEmail(ctx context.Context, email string, code string) error {
	sender, err := getEmailSender()
	if err != nil {
		return err
	}

	msg := EmailMessage{
		From:    defaultEmailFrom,
		To:      []string{email},
		Subject: defaultEmailSubject,
		HTML:    formatVerificationHTML(code),
		Text:    formatVerificationText(code),
	}

	cfg := projectconf.Get()
	if cfg != nil {
		if from := strings.TrimSpace(cfg.Email.FromAddress); from != "" {
			msg.From = from
		}
		if subject := strings.TrimSpace(cfg.Email.Subject); subject != "" {
			msg.Subject = subject
		}
	}

	return sender.Send(ctx, msg)
}

func emailCodeRedisKey(email string) string {
	return fmt.Sprintf("login:email:code:%s", email)
}

func emailCooldownRedisKey(email string) string {
	return fmt.Sprintf("login:email:cooldown:%s", email)
}

func buildMockToken(user AuthUserDTO) string {
	buf, _ := json.Marshal(mockTokenPayload{
		AuthUserDTO: user,
		IssuedAt:    time.Now().Unix(),
	})
	return base64.StdEncoding.EncodeToString(buf)
}

func parseMockToken(token string) (*AuthUserDTO, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	var payload mockTokenPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}

	if payload.Email == "" {
		return nil, fmt.Errorf("invalid-token")
	}

	return &payload.AuthUserDTO, nil
}

func setSessionCookie(ctx *gin.Context, token string) {
	ctx.SetCookie("session_token", token, 3600, "/", "", false, true)
}

type mockTokenPayload struct {
	AuthUserDTO
	IssuedAt int64 `json:"iat"`
}

func deriveNameFromEmail(email string) string {
	for i := range email {
		if email[i] == '@' {
			return email[:i]
		}
	}
	return email
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
