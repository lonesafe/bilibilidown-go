package api

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/tidwall/gjson"

	"bilibilidown-go/internal/model"
	"bilibilidown-go/internal/util"
)

// BilibiliAPI handles all Bilibili API interactions (ported from API.java + AbstractBaseParser.java)
type BilibiliAPI struct {
	http      *util.HTTPClient
	wbiImgKey string
	wbiSubKey string
	wbiOnce   sync.Once
}

// NewBilibiliAPI creates a new API instance
func NewBilibiliAPI() *BilibiliAPI {
	return &BilibiliAPI{
		http: util.NewHTTPClient(),
	}
}

// getWbiKeys fetches wbi signing keys from nav API
func (b *BilibiliAPI) getWbiKeys() (imgKey, subKey string, err error) {
	b.wbiOnce.Do(func() {
		url := "https://api.bilibili.com/x/web-interface/nav"
		body, e := b.http.Get(url, util.CommonHeaders())
		if e != nil {
			err = e
			return
		}
		wbiImg := gjson.Get(body, "data.wbi_img.img_url").String()
		wbiSub := gjson.Get(body, "data.wbi_img.sub_url").String()
		b.wbiImgKey = extractKey(wbiImg)
		b.wbiSubKey = extractKey(wbiSub)
	})
	if b.wbiImgKey == "" {
		err = fmt.Errorf("获取wbi keys失败")
	}
	return b.wbiImgKey, b.wbiSubKey, err
}

func extractKey(url string) string {
	lastSlash := strings.LastIndex(url, "/")
	if lastSlash < 0 {
		return url
	}
	sub := url[lastSlash+1:]
	if dot := strings.Index(sub, "."); dot > 0 {
		sub = sub[:dot]
	}
	return sub
}

// encWbi signs a URL with wbi
func (b *BilibiliAPI) encWbi(rawURL string) (string, error) {
	imgKey, subKey, err := b.getWbiKeys()
	if err != nil {
		return rawURL, err
	}
	return util.EncWbi(rawURL, imgKey, subKey), nil
}

// GetVideoDetail fetches video info including all clips (ported from getAVDetail)
func (b *BilibiliAPI) GetVideoDetail(bvid string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	video := &model.VideoInfo{
		VideoID: bvid,
		Clips:   make(map[int64]*model.ClipInfo),
	}

	// 1. Get page list
	pageListURL := fmt.Sprintf("https://api.bilibili.com/x/player/pagelist?bvid=%s&jsonp=jsonp", bvid)
	pageBody, err := b.http.Get(pageListURL, util.BiliJsonAPIHeaders(bvid))
	if err != nil {
		return nil, fmt.Errorf("获取分页列表失败: %w", err)
	}
	pages := gjson.Get(pageBody, "data")
	if !pages.Exists() || !pages.IsArray() {
		return nil, fmt.Errorf("分页列表解析失败")
	}

	firstCid := pages.Array()[0].Get("cid").Int()

	// 2. Get video detail
	detailURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/wbi/view/detail?platform=web&bvid=%s", bvid)
	signedURL, err := b.encWbi(detailURL)
	if err != nil {
		return nil, err
	}
	detailBody, err := b.http.Get(signedURL, util.BiliJsonAPIHeaders(bvid))
	if err != nil {
		return nil, fmt.Errorf("获取视频详情失败: %w", err)
	}

	code := gjson.Get(detailBody, "code").Int()
	if code == -403 {
		// fallback to fav resource info
		avid := util.Bv2Av(bvid)
		fallbackURL := fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/infos?resources=%d:2&platform=web&folder_mid=&folder_id=", avid)
		fbBody, e := b.http.Get(fallbackURL, util.BiliJsonAPIHeaders(bvid))
		if e != nil {
			return nil, e
		}
		data := gjson.Get(fbBody, "data.0")
		video.VideoName = data.Get("title").String()
		video.Brief = data.Get("intro").String()
		video.Author = data.Get("upper.name").String()
		video.AuthorID = strconv.FormatInt(data.Get("upper.mid").Int(), 10)
		video.VideoPreview = data.Get("cover").String()
	} else {
		data := gjson.Get(detailBody, "data.View")
		video.VideoName = data.Get("title").String()
		video.Brief = data.Get("desc").String()
		video.Author = data.Get("owner.name").String()
		video.AuthorID = strconv.FormatInt(data.Get("owner.mid").Int(), 10)
		video.VideoPreview = data.Get("pic").String()
	}

	// 3. Build clip list
	qnList := b.getVideoQNList(bvid, firstCid)
	pages.ForEach(func(_, page gjson.Result) bool {
		cid := page.Get("cid").Int()
		clip := &model.ClipInfo{
			AvTitle:    video.VideoName,
			AvID:       bvid,
			CID:        cid,
			Page:       int(page.Get("page").Int()),
			Title:      page.Get("part").String(),
			PicPreview: video.VideoPreview,
			UpName:     video.Author,
			UpID:       video.AuthorID,
			Links:      make(map[int]string),
		}
		for _, q := range qnList {
			if fetchLinks {
				if link, e := b.GetVideoLink(bvid, cid, q, 0); e == nil {
					clip.Links[q] = link
				}
			} else {
				clip.Links[q] = ""
			}
		}
		video.Clips[cid] = clip
		return true
	})

	return video, nil
}

// getVideoQNList returns available quality numbers
func (b *BilibiliAPI) getVideoQNList(bvid string, cid int64) []int {
	defaultQN := []int{120, 116, 112, 80, 74, 64, 32, 16}
	url := fmt.Sprintf("https://api.bilibili.com/x/player/playurl?cid=%d&bvid=%s&qn=32&type=&otype=json&fnver=0&fnval=4048&fourk=1", cid, bvid)
	body, err := b.http.Get(url, util.BiliJsonAPIHeaders(bvid))
	if err != nil {
		return defaultQN
	}
	arr := gjson.Get(body, "data.accept_quality")
	if !arr.Exists() {
		return defaultQN
	}
	result := make([]int, 0, len(arr.Array()))
	arr.ForEach(func(_, v gjson.Result) bool {
		result = append(result, int(v.Int()))
		return true
	})
	return result
}

// GetVideoLink gets the download URL for a specific quality (ported from getVideoLinkByFormat)
func (b *BilibiliAPI) GetVideoLink(bvid string, cid int64, qn int, fnval int) (string, error) {
	fnv := 4048
	if fnval != 0 {
		fnv = fnval
	}

	// Try regular playurl endpoint first (works without login)
	url := fmt.Sprintf("https://api.bilibili.com/x/player/playurl?cid=%d&bvid=%s&qn=%d&type=&otype=json&fnver=0&fnval=%d&fourk=1", cid, bvid, qn, fnv)
	body, err := b.http.Get(url, util.BiliJsonAPIHeaders(bvid))
	if err != nil {
		// Fallback to wbi endpoint
		wbiURL := fmt.Sprintf("https://api.bilibili.com/x/player/wbi/playurl?cid=%d&bvid=%s&qn=%d&type=&otype=json&fnver=0&fnval=%d&fourk=1&gaia_source=&from_client=BROWSER&is_main_page=true&need_fragment=false&isGaiaAvoided=false&client_attr=0&session=&voice_balance=1&web_location=1315873", cid, bvid, qn, fnv)
		signedURL, signErr := b.encWbi(wbiURL)
		if signErr != nil {
			return "", err
		}
		body, err = b.http.Get(signedURL, util.BiliJsonAPIHeaders(bvid))
		if err != nil {
			return "", err
		}
	}

	data := gjson.Get(body, "data")
	if !data.Exists() {
		// Try PGC (bangumi) format
		avid := util.Bv2Av(bvid)
		pgcURL := fmt.Sprintf("https://api.bilibili.com/pgc/player/web/playurl?fnver=0&fourk=1&otype=json&avid=%d&cid=%d&qn=%d&fnval=%d", avid, cid, qn, fnv)
		pgcBody, pgcErr := b.http.Get(pgcURL, util.BiliJsonAPIHeaders("av"+strconv.FormatInt(avid, 10)))
		if pgcErr != nil {
			return "", fmt.Errorf("获取播放链接失败")
		}
		data = gjson.Get(pgcBody, "result")
		if !data.Exists() {
			return "", fmt.Errorf("获取播放链接失败: %s", body)
		}
	}

	// Try DASH (parseType1)
	dash := data.Get("dash")
	if dash.Exists() {
		return b.parseDash(dash, qn, bvid)
	}

	// Fallback to durl (parseType2)
	durl := data.Get("durl")
	if durl.Exists() && durl.IsArray() && len(durl.Array()) > 0 {
		return durl.Array()[0].Get("url").String(), nil
	}

	return "", fmt.Errorf("无法解析下载链接")
}

// parseDash parses DASH format response and returns video#audio URL
func (b *BilibiliAPI) parseDash(dash gjson.Result, qn int, bvid string) (string, error) {
	var sb strings.Builder

	// Video stream
	videoArr := dash.Get("video").Array()
	var videoStream gjson.Result
	found := false
	for _, v := range videoArr {
		if int(v.Get("id").Int()) == qn {
			videoStream = v
			found = true
			break
		}
	}
	if !found && len(videoArr) > 0 {
		videoStream = videoArr[0]
	}
	if videoStream.Exists() {
		sb.WriteString(videoStream.Get("base_url").String())
	}
	sb.WriteString("#")

	// Audio stream
	audioArr := dash.Get("audio").Array()
	if len(audioArr) > 0 {
		// Pick highest quality audio (30280 = 192K)
		bestAudio := audioArr[0]
		for _, a := range audioArr {
			id := int(a.Get("id").Int())
			if id == 30280 {
				bestAudio = a
				break
			}
		}
		sb.WriteString(bestAudio.Get("base_url").String())
	}

	return sb.String(), nil
}

// GetSubtitleLink gets subtitle URL
func (b *BilibiliAPI) GetSubtitleLink(bvid string, cid int64) (string, error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/player/wbi/v2?bvid=%s&cid=%d&isGaiaAvoided=false", bvid, cid)
	signedURL, err := b.encWbi(url)
	if err != nil {
		return "", err
	}
	body, err := b.http.Get(signedURL, util.BiliJsonAPIHeaders(bvid))
	if err != nil {
		return "", err
	}
	subtitles := gjson.Get(body, "data.subtitle.subtitles")
	if !subtitles.Exists() || len(subtitles.Array()) == 0 {
		return "", fmt.Errorf("没有字幕")
	}
	return "https:" + subtitles.Array()[0].Get("subtitle_url").String(), nil
}

// GetDanmakuLink gets danmaku XML URL
func (b *BilibiliAPI) GetDanmakuLink(cid int64) string {
	return fmt.Sprintf("https://api.bilibili.com/x/v1/dm/list.so?oid=%d", cid)
}

// LikeVideo likes a video (ported from API.like)
func (b *BilibiliAPI) LikeVideo(bvid string) (bool, error) {
	avid := util.Bv2Av(bvid)
	relURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/archive/relation?aid=%d&bvid=%s", avid, bvid)
	body, err := b.http.Get(relURL, util.ActionHeaders(bvid))
	if err != nil {
		return false, err
	}
	liked := gjson.Get(body, "data.like").Bool()
	if liked {
		return true, nil
	}
	likeURL := "https://api.bilibili.com/x/web-interface/archive/like"
	csrf := util.GetCSRF()
	payload := fmt.Sprintf("aid=%d&like=1&from_spmid=333.1387.homepage.video_card.click&spmid=333.788.0.0&statistics=%%7B%%22appId%%22%%3A100%%2C%%22platform%%22%%3A5%%7D&eab_x=1&ramval=20&source=web_normal&ga=1&csrf=%s", avid, csrf)
	resp, err := b.http.Post(likeURL, util.ActionHeaders(bvid), payload)
	if err != nil {
		return false, err
	}
	return strings.HasPrefix(resp, `{"code":0`), nil
}

// IsLoggedIn checks if the user is logged in
func (b *BilibiliAPI) IsLoggedIn() bool {
	url := "https://api.bilibili.com/x/web-interface/nav"
	body, err := b.http.Get(url, util.CommonHeaders())
	if err != nil {
		return false
	}
	return gjson.Get(body, "data.isLogin").Bool()
}

// GetUserInfo gets current user info
func (b *BilibiliAPI) GetUserInfo() (name string, mid int64, avatar string, err error) {
	url := "https://api.bilibili.com/x/web-interface/nav"
	body, err := b.http.Get(url, util.CommonHeaders())
	if err != nil {
		return "", 0, "", err
	}
	name = gjson.Get(body, "data.uname").String()
	mid = gjson.Get(body, "data.mid").Int()
	avatar = gjson.Get(body, "data.face").String()
	return
}
