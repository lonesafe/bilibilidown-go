package util

// HTTP request headers (ported from HttpHeaders.java)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0"
)

// CommonHeaders returns common HTTP headers
func CommonHeaders() map[string]string {
	return map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "zh-CN,zh;q=0.8",
		"Connection":      "keep-alive",
		"User-Agent":      DefaultUserAgent,
	}
}

// BiliJsonAPIHeaders returns headers for JSON API requests
func BiliJsonAPIHeaders(bvid string) map[string]string {
	return map[string]string{
		"Accept":          "*/*",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "zh-CN,zh;q=0.8",
		"Connection":      "keep-alive",
		"Host":            "api.bilibili.com",
		"Referer":         "https://www.bilibili.com/video/" + bvid,
		"User-Agent":      DefaultUserAgent,
		"X-Requested-With": "ShockwaveFlash/28.0.0.137",
	}
}

// BiliWwwM4sHeaders returns headers for m4s stream download
func BiliWwwM4sHeaders(bvid string) map[string]string {
	return map[string]string{
		"Referer":    "https://www.bilibili.com/video/" + bvid,
		"User-Agent": DefaultUserAgent,
	}
}

// BiliWwwFLVHeaders returns headers for FLV stream download
func BiliWwwFLVHeaders(bvid string) map[string]string {
	return map[string]string{
		"X-Requested-With": "ShockwaveFlash/28.0.0.137",
		"Referer":          "https://www.bilibili.com/video/" + bvid,
		"User-Agent":       DefaultUserAgent,
	}
}

// ActionHeaders returns headers for like/action POST requests
func ActionHeaders(bvid string) map[string]string {
	return map[string]string{
		"User-Agent":      DefaultUserAgent,
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"Accept-Encoding": "gzip, deflate",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Origin":          "https://www.bilibili.com",
		"Referer":         "https://www.bilibili.com/video/" + bvid,
	}
}

// LogoutHeaders returns headers for logout
func LogoutHeaders() map[string]string {
	return map[string]string{
		"User-Agent":      DefaultUserAgent,
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"Accept-Encoding": "gzip, deflate",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Origin":          "https://www.bilibili.com",
		"Referer":         "https://www.bilibili.com/",
	}
}

// BiliLoginAuthHeaders returns headers for login API requests
func BiliLoginAuthHeaders() map[string]string {
	return map[string]string{
		"Accept":          "application/json, text/javascript, */*; q=0.01",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "zh-CN,zh;q=0.8",
		"Connection":      "keep-alive",
		"Host":            "passport.bilibili.com",
		"Referer":         "https://passport.bilibili.com/login",
		"X-Requested-With": "XMLHttpRequest",
		"User-Agent":      DefaultUserAgent,
	}
}
