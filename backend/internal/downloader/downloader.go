package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bilibilidown-go/internal/config"
	"bilibilidown-go/internal/util"
)

// Status represents download task status
type Status string

const (
	StatusWaiting    Status = "waiting"
	StatusDownloading Status = "downloading"
	StatusMerging    Status = "merging"
	StatusSuccess    Status = "success"
	StatusFail       Status = "fail"
	StatusStopped    Status = "stopped"
)

// Task represents a single download task
type Task struct {
	mu         sync.Mutex
	ID         string
	AvID       string
	CID        int64
	Title      string
	URL        string // video#audio
	Status     Status
	TotalSize  int64
	Downloaded int64
	Speed      int64
	FileName   string
	ErrMsg     string
	stopCh     chan struct{}
}

// Downloader manages download tasks
type Downloader struct {
	mu      sync.RWMutex
	tasks   map[string]*Task
	http    *util.HTTPClient
	pool    chan struct{} // semaphore for concurrency
}

// NewDownloader creates a download manager
func NewDownloader() *Downloader {
	cfg := config.Get()
	return &Downloader{
		tasks: make(map[string]*Task),
		http:  util.NewHTTPClient(),
		pool:  make(chan struct{}, cfg.DownloadPoolSize),
	}
}

// AddTask creates and starts a download task
func (d *Downloader) AddTask(avid string, cid int64, title, url string) *Task {
	id := fmt.Sprintf("%s-%d", avid, cid)
	task := &Task{
		ID:     id,
		AvID:   avid,
		CID:    cid,
		Title:  title,
		URL:    url,
		Status: StatusWaiting,
		stopCh: make(chan struct{}),
	}
	d.mu.Lock()
	d.tasks[id] = task
	d.mu.Unlock()

	go d.execute(task)
	return task
}

// GetTask returns a task by ID
func (d *Downloader) GetTask(id string) *Task {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.tasks[id]
}

// GetAllTasks returns all tasks
func (d *Downloader) GetAllTasks() []*Task {
	d.mu.RLock()
	defer d.mu.RUnlock()
	tasks := make([]*Task, 0, len(d.tasks))
	for _, t := range d.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// StopTask stops a running task
func (d *Downloader) StopTask(id string) {
	d.mu.RLock()
	task, ok := d.tasks[id]
	d.mu.RUnlock()
	if ok {
		select {
		case <-task.stopCh:
		default:
			close(task.stopCh)
		}
		task.mu.Lock()
		task.Status = StatusStopped
		task.mu.Unlock()
	}
}

// TaskSnapshot is a read-only snapshot of task state
type TaskSnapshot struct {
	ID         string `json:"id"`
	AvID       string `json:"avid"`
	CID        int64  `json:"cid"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	TotalSize  int64  `json:"totalSize"`
	Downloaded int64  `json:"downloaded"`
	Speed      int64  `json:"speed"`
	FileName   string `json:"fileName"`
	ErrMsg     string `json:"errMsg"`
}

// Snapshot returns a thread-safe snapshot of the task
func (t *Task) Snapshot() TaskSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return TaskSnapshot{
		ID:         t.ID,
		AvID:       t.AvID,
		CID:        t.CID,
		Title:      t.Title,
		Status:     string(t.Status),
		TotalSize:  t.TotalSize,
		Downloaded: t.Downloaded,
		Speed:      t.Speed,
		FileName:   t.FileName,
		ErrMsg:     t.ErrMsg,
	}
}

func (d *Downloader) execute(task *Task) {
	d.pool <- struct{}{} // acquire concurrency slot
	defer func() { <-d.pool }()

	task.mu.Lock()
	task.Status = StatusDownloading
	task.mu.Unlock()

	cfg := config.Get()
	parts := strings.SplitN(task.URL, "#", 2)
	videoURL := parts[0]
	audioURL := ""
	if len(parts) > 1 {
		audioURL = parts[1]
	}

	baseName := fmt.Sprintf("%s-%d", task.AvID, task.CID)
	saveDir := cfg.SavePath
	videoFile := filepath.Join(saveDir, baseName+"_video.m4s")
	audioFile := filepath.Join(saveDir, baseName+"_audio.m4s")
	outputFile := filepath.Join(saveDir, baseName+".mp4")
	task.FileName = baseName + ".mp4"

	// Download video
	if videoURL != "" {
		ok, err := d.downloadWithProgress(task, videoURL, videoFile, task.AvID, false)
		if err != nil {
			task.mu.Lock()
			task.Status = StatusFail
			task.ErrMsg = "视频下载失败: " + err.Error()
			task.mu.Unlock()
			return
		}
		if !ok {
			return // stopped
		}
	}

	// Download audio
	if audioURL != "" {
		ok, err := d.downloadWithProgress(task, audioURL, audioFile, task.AvID, true)
		if err != nil {
			task.mu.Lock()
			task.Status = StatusFail
			task.ErrMsg = "音频下载失败: " + err.Error()
			task.mu.Unlock()
			return
		}
		if !ok {
			return // stopped
		}
	}

	// Merge with ffmpeg
	task.mu.Lock()
	task.Status = StatusMerging
	task.mu.Unlock()

	if err := mergeStreams(cfg.FFmpegPath, videoFile, audioFile, outputFile); err != nil {
		task.mu.Lock()
		task.Status = StatusFail
		task.ErrMsg = "合并失败: " + err.Error()
		task.mu.Unlock()
		return
	}

	// Cleanup temp files
	os.Remove(videoFile)
	os.Remove(audioFile)

	task.mu.Lock()
	task.Status = StatusSuccess
	task.mu.Unlock()
}

func (d *Downloader) downloadWithProgress(task *Task, url, savePath, bvid string, isAudio bool) (bool, error) {
	headers := util.BiliWwwM4sHeaders(bvid)

	lastTime := time.Now()
	lastDownloaded := task.Downloaded

	ok, err := d.http.Download(url, savePath, headers, func(downloaded, total int64) {
		task.mu.Lock()
		if isAudio {
			// For audio, add to progress
		}
		task.Downloaded = downloaded
		if total > 0 {
			task.TotalSize = total
		}
		// Calculate speed
		now := time.Now()
		elapsed := now.Sub(lastTime).Seconds()
		if elapsed >= 1.0 {
			delta := downloaded - lastDownloaded
			if delta > 0 {
				task.Speed = int64(float64(delta) / elapsed)
			}
			lastTime = now
			lastDownloaded = downloaded
		}
		task.mu.Unlock()

		// Check stop signal
		select {
		case <-task.stopCh:
		default:
		}
	})
	return ok, err
}

// mergeStreams merges video and audio using ffmpeg
func mergeStreams(ffmpegPath, videoFile, audioFile, outputFile string) error {
	ffmpeg := "ffmpeg"
	if ffmpegPath != "" {
		ffmpeg = ffmpegPath
	}

	var args []string
	args = append(args, "-y")
	if videoFile != "" {
		args = append(args, "-i", videoFile)
	}
	if audioFile != "" {
		args = append(args, "-i", audioFile)
	}
	args = append(args, "-c", "copy", outputFile)

	cmd := exec.Command(ffmpeg, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// FormatSize formats bytes to human readable string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
