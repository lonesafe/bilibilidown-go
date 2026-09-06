package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	"bilibilidown-go/internal/util"
)

// QRCodeResult holds the QR code generation result
type QRCodeResult struct {
	URL        string `json:"url"`
	QrcodeKey  string `json:"qrcodeKey"`
}

// PollResult holds the QR code polling result
type PollResult struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	RefreshToken string `json:"refreshToken"`
	CookieStr    string `json:"-"`
}

// GenQRCode generates a QR code for login (ported from INeedLogin.getAuthKey)
func (b *BilibiliAPI) GenQRCode() (*QRCodeResult, error) {
	url := "https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-web"
	body, _, err := b.http.GetWithCookies(url, util.BiliLoginAuthHeaders())
	if err != nil {
		return nil, fmt.Errorf("获取二维码失败: %w", err)
	}
	code := gjson.Get(body, "code").Int()
	if code != 0 {
		return nil, fmt.Errorf("获取二维码失败: %s", gjson.Get(body, "message").String())
	}
	return &QRCodeResult{
		URL:       gjson.Get(body, "data.url").String(),
		QrcodeKey: gjson.Get(body, "data.qrcode_key").String(),
	}, nil
}

// PollQRCode polls the QR code login status (ported from INeedLogin.getAuthStatus)
func (b *BilibiliAPI) PollQRCode(qrcodeKey string) (*PollResult, error) {
	url := fmt.Sprintf("https://passport.bilibili.com/x/passport-login/web/qrcode/poll?source=main-web&qrcode_key=%s", qrcodeKey)
	body, cookies, err := b.http.GetWithCookies(url, util.BiliLoginAuthHeaders())
	if err != nil {
		return nil, fmt.Errorf("轮询二维码状态失败: %w", err)
	}

	data := gjson.Get(body, "data")
	code := int(data.Get("code").Int())
	msg := data.Get("message").String()

	result := &PollResult{
		Code:    code,
		Message: msg,
	}

	// code 0 = success, 86038 = not scanned, 86090 = scanned but not confirmed, 86039 = expired
	if code == 0 {
		result.RefreshToken = data.Get("refresh_token").String()
		// Capture Set-Cookie headers from response
		if len(cookies) > 0 {
			cookieStr := cookiesToString(cookies)
			result.CookieStr = cookieStr
		}
	}

	return result, nil
}

// cookiesToString converts []*http.Cookie to cookie string
func cookiesToString(cookies []*http.Cookie) string {
	parts := make([]string, 0, len(cookies))
	for _, c := range cookies {
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}
