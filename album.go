package tb

import (
	"sync"
	"time"

	"github.com/xnumb/tb/utils"
	tele "gopkg.in/telebot.v4"
)

type Album struct {
	Tid       int64
	List      []string
	Text      string
	CreatedAt time.Time
	timer     *time.Timer // 为每个相册关联一个定时器
}

// AlbumManager 负责管理所有相册逻辑
type AlbumManager struct {
	albums map[string]*Album
	mu     sync.Mutex
}

// NewAlbumManager 创建一个新的管理器
func NewAlbumManager() *AlbumManager {
	m := &AlbumManager{
		albums: make(map[string]*Album),
	}
	// 启动一个独立的后台任务，每分钟清理一次过期相册
	go m.startCleaner(10 * time.Minute) // 清理周期可以设置长一点
	return m
}

// Handle 是处理 Telegram 消息的入口
func (m *AlbumManager) Handle(c tele.Context, fn func(Album)) {
	msg := c.Message()
	aid := msg.AlbumID
	if aid == "" {
		return
	}

	fid := ""
	if msg.Photo != nil {
		fid = msg.Photo.FileID
	} else if msg.Video != nil {
		fid = "_" + msg.Video.FileID // 使用前缀区分，比纯粹的"_"更清晰
	} else {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if album, exist := m.albums[aid]; exist {
		// 相册已存在，添加图片并重置定时器
		album.List = append(album.List, fid)
		if msg.Caption != "" {
			album.Text = msg.Caption
		}
		// 核心：重置定时器，给新图片2秒的等待时间
		album.timer.Reset(2 * time.Second)
	} else {
		// 新相册，创建并启动定时器
		album := &Album{
			Tid:       msg.Sender.ID,
			List:      []string{fid},
			Text:      msg.Caption,
			CreatedAt: utils.GetNow(),
		}

		// 定时器触发后，执行回调并从 map 中删除自己
		album.timer = time.AfterFunc(2*time.Second, func() {
			m.processAlbum(aid, fn)
		})

		m.albums[aid] = album
	}
}

// processAlbum 在定时器触发后执行
func (m *AlbumManager) processAlbum(aid string, fn func(Album)) {
	m.mu.Lock()
	// 再次检查相册是否存在，并取出数据
	album, exist := m.albums[aid]
	if !exist {
		m.mu.Unlock()
		return
	}
	// 从 map 中删除，防止重复处理和内存泄漏
	delete(m.albums, aid)
	m.mu.Unlock()

	// 在锁外执行回调，避免阻塞其他消息处理
	fn(*album)
}

// startCleaner 启动一个周期性的清理任务
func (m *AlbumManager) startCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := utils.GetNow()
		for key, album := range m.albums {
			// 清理那些因为各种异常情况而残留超过10分钟的相册
			if album.CreatedAt.Add(10 * time.Minute).Before(now) {
				album.timer.Stop() // 别忘了停止定时器
				delete(m.albums, key)
			}
		}
		m.mu.Unlock()
	}
}
