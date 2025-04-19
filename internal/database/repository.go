package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"tg-video-bot/internal/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// VideoRepository представляет репозиторий для работы с видео
type VideoRepository struct {
	db *sql.DB
}

// NewVideoRepository создает новый экземпляр репозитория
func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func InitDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	fmt.Println(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %v", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка соединения с ретраями
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := retry(3, 2*time.Second, func() error {
		return db.PingContext(ctx)
	}); err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %v", err)
	}

	// Применяем миграции
	migrator := NewMigrator(db)
	if err := migrator.Run(); err != nil {
		return nil, fmt.Errorf("migrations failed: %v", err)
	}

	return db, nil
}

// SaveVideo сохраняет видео в базу данных
func (r *VideoRepository) SaveVideo(video models.Video) (int64, error) {
	result, err := r.db.Exec(
		"INSERT INTO videos (file_id, caption) VALUES (?, ?)",
		video.FileID,
		video.Caption,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка сохранения видео: %v", err)
	}

	return result.LastInsertId()
}

// GetVideoByID возвращает видео по его ID
func (r *VideoRepository) GetVideoByID(id int64) (models.Video, error) {
	var video models.Video
	err := r.db.QueryRow(
		"SELECT id, file_id, caption FROM videos WHERE id = ?",
		id,
	).Scan(&video.ID, &video.FileID, &video.Caption)
	fmt.Println(id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return video, fmt.Errorf("видео с ID %d не найдено", id)
		}
		return video, fmt.Errorf("ошибка получения видео: %v", err)
	}

	// Получаем теги для видео
	tags, err := r.GetVideoTags(id)
	if err != nil {
		return video, fmt.Errorf("ошибка получения тегов: %v", err)
	}
	video.Tags = tags

	return video, nil
}

// GetVideosByTag возвращает все видео с указанным тегом
func (r *VideoRepository) GetVideosByTag(tag string) ([]models.Video, error) {
	rows, err := r.db.Query(`
		SELECT v.id, v.file_id, v.caption 
		FROM videos v
		JOIN video_tags vt ON v.id = vt.video_id
		JOIN tags t ON vt.tag_id = t.id
		WHERE t.name = ?
	`, tag)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса видео по тегу: %v", err)
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var video models.Video
		if err := rows.Scan(&video.ID, &video.FileID, &video.Caption); err != nil {
			return nil, fmt.Errorf("ошибка сканирования видео: %v", err)
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// AddTagsToVideo добавляет теги к видео
func (r *VideoRepository) AddTagsToVideo(videoID int64, tags []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, tagName := range tags {
		// Нормализуем тег
		tagName = strings.TrimSpace(strings.ToLower(tagName))
		if tagName == "" {
			continue
		}

		// Добавляем тег или получаем существующий ID
		var tagID int64
		err = tx.QueryRow(
			"SELECT id FROM tags WHERE name = ?",
			tagName,
		).Scan(&tagID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result, err := tx.Exec(
					"INSERT INTO tags (name) VALUES (?)",
					tagName,
				)
				if err != nil {
					return fmt.Errorf("ошибка добавления тега: %v", err)
				}
				tagID, _ = result.LastInsertId()
			} else {
				return fmt.Errorf("ошибка запроса тега: %v", err)
			}
		}

		// Связываем видео и тег
		_, err = tx.Exec(
			"INSERT IGNORE INTO video_tags (video_id, tag_id) VALUES (?, ?)",
			videoID,
			tagID,
		)
		if err != nil {
			return fmt.Errorf("ошибка связывания видео и тега: %v", err)
		}
	}

	return tx.Commit()
}

// GetVideoTags возвращает все теги для видео
func (r *VideoRepository) GetVideoTags(videoID int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT t.name 
		FROM tags t
		JOIN video_tags vt ON t.id = vt.tag_id
		WHERE vt.video_id = ?
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса тегов: %v", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("ошибка сканирования тега: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// IsVideoSent проверяет, отправлялось ли видео в указанный чат
func (r *VideoRepository) IsVideoSent(chatID, videoID int64) bool {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM sent_videos 
			WHERE chat_id = ? AND video_id = ?
		)
	`, chatID, videoID).Scan(&exists)

	if err != nil {
		log.Printf("Ошибка проверки отправки видео: %v", err)
		return false
	}

	return exists
}

// MarkVideoSent отмечает видео как отправленное в чат
func (r *VideoRepository) MarkVideoSent(chatID, videoID int64) error {
	_, err := r.db.Exec(
		"INSERT INTO sent_videos (chat_id, video_id) VALUES (?, ?)",
		chatID,
		videoID,
	)
	return err
}

// GetPopularTags возвращает самые популярные теги
func (r *VideoRepository) GetPopularTags(limit int) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT t.name, COUNT(vt.video_id) as count
		FROM tags t
		JOIN video_tags vt ON t.id = vt.tag_id
		GROUP BY t.name
		ORDER BY count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса популярных тегов: %v", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, fmt.Errorf("ошибка сканирования тега: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// VideoExists проверяет существование видео по ID
func (r *VideoRepository) VideoExists(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM videos WHERE id = ?)",
		id,
	).Scan(&exists)

	return exists, err
}

// GetRandomUnsentVideo возвращает случайное видео, которое еще не было отправлено в указанный чат
func (r *VideoRepository) GetRandomUnsentVideo(chatID int64) (models.Video, error) {
	var video models.Video

	query := `
		SELECT v.id, v.file_id, v.caption 
		FROM videos v
		WHERE NOT EXISTS (
			SELECT 1 FROM sent_videos sv 
			WHERE sv.video_id = v.id AND sv.chat_id = ?
		)
		ORDER BY RAND()
		LIMIT 1`

	err := r.db.QueryRow(query, chatID).Scan(
		&video.ID,
		&video.FileID,
		&video.Caption,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return video, fmt.Errorf("no unsent videos available")
		}
		return video, fmt.Errorf("failed to get random video: %v", err)
	}

	// Получаем теги для видео
	tags, err := r.GetVideoTags(video.ID)
	if err != nil {
		return video, fmt.Errorf("failed to get video tags: %v", err)
	}
	video.Tags = tags

	return video, nil
}

func (r *VideoRepository) GetAllVideos() ([]models.Video, error) {
	rows, err := r.db.Query(`
		SELECT id, file_id, caption 
		FROM videos 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.ID, &v.FileID, &v.Caption); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}

	return videos, nil
}

func (r *VideoRepository) DeleteVideo(id int64) error {
	_, err := r.db.Exec("DELETE FROM videos WHERE id = ?", id)
	return err
}

func retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}
