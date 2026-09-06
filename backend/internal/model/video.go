package model

// VideoInfo represents a video's metadata (ported from VideoInfo.java)
type VideoInfo struct {
	VideoName    string                `json:"videoName"`
	VideoID      string                `json:"videoId"`
	Author       string                `json:"author"`
	AuthorID     string                `json:"authorId"`
	VideoPreview string                `json:"videoPreview"`
	Brief        string                `json:"brief"`
	VideoLink    string                `json:"videoLink"`
	Clips        map[int64]*ClipInfo   `json:"clips"`
}

// ClipInfo represents a single clip/page of a video (ported from ClipInfo.java)
type ClipInfo struct {
	AvTitle       string            `json:"avTitle"`
	CID           int64             `json:"cId"`
	AvID          string            `json:"avId"`
	Page          int               `json:"page"`
	Title         string            `json:"title"`
	PicPreview    string            `json:"picPreview"`
	ListName      string            `json:"listName"`
	ListOwnerName string            `json:"listOwnerName"`
	ListOwnerID   string            `json:"listOwnerId"`
	FavTime       int64             `json:"favTime"`
	CTime         int64             `json:"cTime"`
	UpName        string            `json:"upName"`
	UpID          string            `json:"upId"`
	Links         map[int]string    `json:"links"` // qn -> url
	Remark        int               `json:"remark"`
}

// DownloadMode 0=audio+video 1=video only 2=audio only
type DownloadMode int

const (
	ModeAll DownloadMode = iota
	ModeVideoOnly
	ModeAudioOnly
)

// DownloadModeEnum string mapping
var DownloadModeNames = map[DownloadMode]string{
	ModeAll:       "音视频",
	ModeVideoOnly: "仅视频",
	ModeAudioOnly: "仅音频",
}

// TaskInfo represents a download task status
type TaskInfo struct {
	ID            string `json:"id"`
	AvID          string `json:"avId"`
	CID           int64  `json:"cId"`
	Title         string `json:"title"`
	Status        string `json:"status"` // waiting, downloading, success, fail, stopped, merging
	TotalSize     int64  `json:"totalSize"`
	Downloaded    int64  `json:"downloaded"`
	Progress      int    `json:"progress"`
	Speed         int64  `json:"speed"`
	FileName      string `json:"fileName"`
	ErrMsg        string `json:"errMsg"`
}
