package util

import (
	"net/http"
	"strings"
	"sync"
)

// CookieManager manages global cookies (ported from HttpCookies.java)
type CookieManager struct {
	mu         sync.RWMutex
	cookies    map[string]string
	fingerprint string
	csrf       string
}

var globalCookies = &CookieManager{cookies: make(map[string]string)}

// SetGlobalCookies sets cookies from a cookie string
func SetGlobalCookies(cookieStr string) {
	globalCookies.mu.Lock()
	defer globalCookies.mu.Unlock()
	globalCookies.cookies = make(map[string]string)
	globalCookies.csrf = ""
	if cookieStr == "" {
		return
	}
	for _, part := range strings.Split(cookieStr, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			globalCookies.cookies[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
}

// GetGlobalCookies returns the cookie string
func GetGlobalCookies() string {
	globalCookies.mu.RLock()
	defer globalCookies.mu.RUnlock()
	return cookieMapToString(globalCookies.cookies)
}

// SetFingerprint sets the fingerprint cookie string
func SetFingerprint(fp string) {
	globalCookies.mu.Lock()
	defer globalCookies.mu.Unlock()
	globalCookies.fingerprint = fp
}

// GetFingerprint returns the fingerprint
func GetFingerprint() string {
	globalCookies.mu.RLock()
	defer globalCookies.mu.RUnlock()
	return globalCookies.fingerprint
}

// CookiesWithFingerprint returns cookies merged with fingerprint
func CookiesWithFingerprint() string {
	globalCookies.mu.RLock()
	defer globalCookies.mu.RUnlock()
	merged := make(map[string]string)
	if globalCookies.fingerprint != "" {
		for _, part := range strings.Split(globalCookies.fingerprint, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				merged[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	for k, v := range globalCookies.cookies {
		merged[k] = v
	}
	return cookieMapToString(merged)
}

// GetCSRF returns the bili_jct csrf token
func GetCSRF() string {
	globalCookies.mu.RLock()
	defer globalCookies.mu.RUnlock()
	if globalCookies.csrf != "" {
		return globalCookies.csrf
	}
	return globalCookies.cookies["bili_jct"]
}

// GetCookie returns a specific cookie value
func GetCookie(name string) string {
	globalCookies.mu.RLock()
	defer globalCookies.mu.RUnlock()
	if v, ok := globalCookies.cookies[name]; ok {
		return v
	}
	if globalCookies.fingerprint != "" {
		for _, part := range strings.Split(globalCookies.fingerprint, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, name+"=") {
				return strings.TrimPrefix(part, name+"=")
			}
		}
	}
	return ""
}

// SetCookie sets a single cookie value
func SetCookie(name, value string) {
	globalCookies.mu.Lock()
	defer globalCookies.mu.Unlock()
	globalCookies.cookies[name] = value
}

func cookieMapToString(m map[string]string) string {
	var sb strings.Builder
	first := true
	for k, v := range m {
		if !first {
			sb.WriteString("; ")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
		first = false
	}
	return sb.String()
}

// ApplyCookiesToRequest adds cookies to an http.Request
func ApplyCookiesToRequest(req *http.Request) {
	cookieStr := CookiesWithFingerprint()
	if cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}
}
