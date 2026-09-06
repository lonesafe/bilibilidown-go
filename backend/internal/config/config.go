package config

import (
	"os"
	"path/filepath"
	"sync"
)

// Config holds application settings (ported from Global.java)
type Config struct {
	// Download
	SavePath          string `json:"savePath"`
	FFmpegPath        string `json:"ffmpegPath"`
	DownloadFormat    int    `json:"downloadFormat"`    // 0:m4s 1:flv 2:mp4
	DownloadPoolSize  int    `json:"downloadPoolSize"`  // concurrent download count
	MaxFailRetry      int    `json:"maxFailRetry"`
	MultiThreadCount  int    `json:"multiThreadCount"`
	MultiThreadMinFS  int64  `json:"multiThreadMinFileSize"`

	// Quality
	VideoCodecPriority []int `json:"videoCodecPriority"` // 7:AVC 12:HEVC 13:AV1
	AudioQualityPriority []int `json:"audioQualityPriority"`

	// Naming
	NameFormat        string `json:"nameFormat"`

	// Cookie / Login
	CookieFile        string `json:"-"`
	ServerPort        int    `json:"serverPort"`

	// UI
	PageSize          int    `json:"pageSize"`
}

var (
	global *Config
	once   sync.Once
)

// Get returns the global config instance
func Get() *Config {
	once.Do(func() {
		global = defaultConfig()
	})
	return global
}

func defaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		SavePath:           "./download/",
		FFmpegPath:         "",
		DownloadFormat:     0,
		DownloadPoolSize:   3,
		MaxFailRetry:       3,
		MultiThreadCount:   0,
		VideoCodecPriority: []int{12, 7, 13},
		AudioQualityPriority: []int{30280, 30232, 30216, -1, 30251, 30250},
		NameFormat:         "(:listName 0_listName-)UpName/avTitle-(:favTime favTime-)pAv2-qn",
		CookieFile:         filepath.Join(home, ".config", "bilibilidown", "cookies"),
		ServerPort:         8787,
		PageSize:           20,
	}
}
