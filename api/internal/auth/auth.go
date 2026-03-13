package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/config"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/db"
)

type contextKey string

const UserKey contextKey = "user"

type UserInfo struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Source      string `json:"source"`
}

type sessionData struct {
	Username    string
	DisplayName string
	ExpiresAt   int64
}

type Handler struct {
	cfg      *config.Config
	store    db.UserStore
	mu       sync.RWMutex
	sessions map[string]sessionData
}

func NewHandler(cfg *config.Config, store db.UserStore) *Handler {
	return &Handler{
		cfg:      cfg,
		store:    store,
		sessions: make(map[string]sessionData),
	}
}

func (h *Handler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	if h.cfg.DevMode {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	authURL := fmt.Sprintf("https://%s/authorize?"+
		"response_type=code&"+
		"client_id=%s&"+
		"redirect_uri=%s&"+
		"scope=openid+profile&"+
		"connection=github",
		h.cfg.Auth0Domain,
		url.QueryEscape(h.cfg.Auth0ClientID),
		url.QueryEscape(h.cfg.Auth0Callback),
	)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

type auth0TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
}

type auth0UserInfo struct {
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Sub      string `json:"sub"`
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	tokenURL := fmt.Sprintf("https://%s/oauth/token", h.cfg.Auth0Domain)
	body := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {h.cfg.Auth0ClientID},
		"client_secret": {h.cfg.Auth0Secret},
		"code":          {code},
		"redirect_uri":  {h.cfg.Auth0Callback},
	}

	resp, err := http.PostForm(tokenURL, body)
	if err != nil {
		http.Error(w, "failed to exchange code", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResp auth0TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		http.Error(w, "failed to decode token response", http.StatusInternalServerError)
		return
	}

	userInfoURL := fmt.Sprintf("https://%s/userinfo", h.cfg.Auth0Domain)
	req, _ := http.NewRequestWithContext(r.Context(), "GET", userInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	uiResp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	defer uiResp.Body.Close()

	var userInfo auth0UserInfo
	if err := json.NewDecoder(uiResp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "failed to decode user info", http.StatusInternalServerError)
		return
	}

	if !h.isAllowedUser(userInfo.Nickname) {
		http.Error(w, "user not in allowlist", http.StatusForbidden)
		return
	}

	_, _ = h.store.UpsertUser(r.Context(), userInfo.Nickname, &userInfo.Name)

	sessionID := generateSessionID()
	h.mu.Lock()
	h.sessions[sessionID] = sessionData{
		Username:    userInfo.Nickname,
		DisplayName: userInfo.Name,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	h.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   h.cfg.CookieDomain != "localhost",
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		h.mu.Lock()
		delete(h.sessions, cookie.Value)
		h.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieDomain != "localhost",
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if h.cfg.DevMode {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserInfo{
			Username:    "dev",
			DisplayName: "Dev User",
			Source:      "dev",
		})
		return
	}
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dev mode: bypass auth entirely
		if h.cfg.DevMode {
			displayName := "Dev User"
			_, _ = h.store.UpsertUser(r.Context(), "dev", &displayName)
			ctx := context.WithValue(r.Context(), UserKey, UserInfo{
				Username:    "dev",
				DisplayName: "Dev User",
				Source:      "dev",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Try cookie auth first (SPA users)
		if cookie, err := r.Cookie("session"); err == nil {
			h.mu.RLock()
			session, ok := h.sessions[cookie.Value]
			h.mu.RUnlock()
			if ok && time.Now().Unix() < session.ExpiresAt {
				ctx := context.WithValue(r.Context(), UserKey, UserInfo{
					Username:    session.Username,
					DisplayName: session.DisplayName,
					Source:      "cookie",
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if ok {
				h.mu.Lock()
				delete(h.sessions, cookie.Value)
				h.mu.Unlock()
			}
		}

		// Try Bearer JWT (machine clients)
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			user, err := h.validateJWT(tokenStr)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserKey, *user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	})
}

func (h *Handler) validateJWT(tokenStr string) (*UserInfo, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", h.cfg.Auth0Domain)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return fetchJWKS(jwksURL, kid)
	},
		jwt.WithIssuer(fmt.Sprintf("https://%s/", h.cfg.Auth0Domain)),
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	sub, _ := claims["sub"].(string)
	clientID, _ := claims["azp"].(string)

	if h.isAllowedMachine(clientID) {
		return &UserInfo{
			Username:    sub,
			DisplayName: "machine:" + clientID,
			Source:      "bearer",
		}, nil
	}

	nickname, _ := claims["nickname"].(string)
	if nickname != "" && h.isAllowedUser(nickname) {
		return &UserInfo{
			Username:    nickname,
			DisplayName: nickname,
			Source:      "bearer",
		}, nil
	}

	return nil, fmt.Errorf("not in allowlist")
}

func (h *Handler) isAllowedUser(username string) bool {
	if len(h.cfg.AllowedUsers) == 0 {
		return true
	}
	for _, u := range h.cfg.AllowedUsers {
		if u == username {
			return true
		}
	}
	return false
}

func (h *Handler) isAllowedMachine(clientID string) bool {
	if len(h.cfg.AllowedMachines) == 0 {
		return true
	}
	for _, m := range h.cfg.AllowedMachines {
		if m == clientID {
			return true
		}
	}
	return false
}

func UserFromContext(ctx context.Context) (UserInfo, bool) {
	u, ok := ctx.Value(UserKey).(UserInfo)
	return u, ok
}

func generateSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
