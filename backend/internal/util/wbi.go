package util

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WBI signing (ported from Java API class)
var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
	27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
}

func getMixinKey(orig string) string {
	var sb strings.Builder
	for i := 0; i < 32; i++ {
		sb.WriteByte(orig[mixinKeyEncTab[i]])
	}
	return sb.String()
}

// EncWbi signs a URL with WBI parameters
func EncWbi(rawURL, imgKey, subKey string) string {
	mixinKey := getMixinKey(imgKey + subKey)
	wts := time.Now().Unix()

	var queryStr string
	var prefix string
	if idx := strings.Index(rawURL, "?"); idx >= 0 {
		queryStr = rawURL[idx+1:]
		prefix = rawURL[:idx+1]
	} else {
		queryStr = ""
		prefix = rawURL + "?"
	}

	queryStr += "&wts=" + strconv.FormatInt(wts, 10)

	params := strings.Split(queryStr, "&")
	encoded := make([]string, 0, len(params))
	for _, p := range params {
		if p == "" {
			continue
		}
		parts := strings.SplitN(p, "=", 2)
		k := url.QueryEscape(parts[0])
		var v string
		if len(parts) >= 2 {
			v = encodeURLValue(parts[1])
		}
		encoded = append(encoded, k+"="+v)
	}
	sort.Strings(encoded)
	sorted := strings.Join(encoded, "&")

	wbiRaw := sorted + mixinKey
	h := md5.Sum([]byte(wbiRaw))
	wRid := hex.EncodeToString(h[:])

	return prefix + sorted + "&w_rid=" + wRid
}

func encodeURLValue(s string) string {
	if strings.Contains(s, "%") {
		return strings.ReplaceAll(s, "+", "%20")
	}
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}
