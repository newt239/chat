package attachment

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type stubAttachmentRepo struct {
	domainrepository.AttachmentRepository
	created *entity.Attachment
	found   *entity.Attachment
}

func (r *stubAttachmentRepo) FindByID(_ context.Context, _ string) (*entity.Attachment, error) {
	return r.found, nil
}

func (r *stubAttachmentRepo) CreatePending(_ context.Context, attachment *entity.Attachment) error {
	r.created = attachment
	return nil
}

type stubChannelAccess struct {
	service.ChannelAccessService
}

func (stubChannelAccess) EnsureChannelAccess(_ context.Context, channelID string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: channelID}, nil
}

type stubStorage struct {
	uploadedMimeType string
}

func (s *stubStorage) GenerateUploadURL(key string, mimeType string, _ int64, _ interface{}) (string, error) {
	if !strings.HasSuffix(key, "-thumbnail") {
		s.uploadedMimeType = mimeType
	}
	return "https://storage.example.com/" + key, nil
}

func (s *stubStorage) GenerateDownloadURL(key string, _ interface{}) (string, error) {
	return "https://storage.example.com/" + key, nil
}

func (s *stubStorage) DeleteObject(_ string) error {
	return nil
}

type stubStorageConfig struct{}

func (stubStorageConfig) GetMaxFileSize() int64           { return 1 << 30 }
func (stubStorageConfig) GetUploadExpires() interface{}   { return time.Minute }
func (stubStorageConfig) GetDownloadExpires() interface{} { return time.Minute }

func TestPresignStoresMediaMetadata(t *testing.T) {
	repo := &stubAttachmentRepo{}
	storage := &stubStorage{}
	interactor := NewInteractor(repo, nil, stubChannelAccess{}, storage, stubStorageConfig{})
	width, height, duration := int32(1920), int32(1080), 84.5

	_, err := interactor.Presign(context.Background(), &PresignInput{
		UserID:    "u1",
		ChannelID: "ch1",
		FileName:  "demo.mp4",
		MimeType:  "Video/MP4; codecs=avc1",
		SizeBytes: 1024,
		Media:     entity.MediaMetadata{Width: &width, Height: &height, DurationSeconds: &duration},
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if repo.created.MimeType != "video/mp4" || storage.uploadedMimeType != "video/mp4" {
		t.Errorf("MIME タイプが正規化されていません: saved=%s upload=%s", repo.created.MimeType, storage.uploadedMimeType)
	}
	if media := repo.created.Media; *media.Width != width || *media.Height != height || *media.DurationSeconds != duration {
		t.Errorf("メディアのメタデータが保存されていません: %+v", media)
	}
}

func TestNormalizeMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		fileName string
		want     string
	}{
		{mimeType: "image/png", fileName: "a.png", want: "image/png"},
		{mimeType: "application/octet-stream", fileName: "photo.JPG", want: "image/jpeg"},
		{mimeType: "", fileName: "icon.webp", want: "image/webp"},
		{mimeType: "", fileName: "unknown.zzz", want: "application/octet-stream"},
	}
	for _, tt := range tests {
		if got := normalizeMimeType(tt.mimeType, tt.fileName); got != tt.want {
			t.Errorf("normalizeMimeType(%q, %q) = %q, want %q", tt.mimeType, tt.fileName, got, tt.want)
		}
	}
}

func TestPresignThumbnail(t *testing.T) {
	repo := &stubAttachmentRepo{}
	interactor := NewInteractor(repo, nil, stubChannelAccess{}, &stubStorage{}, stubStorageConfig{})
	thumbnail := &ThumbnailInput{MimeType: "image/jpeg", SizeBytes: 2048, Width: 640, Height: 360}

	out, err := interactor.Presign(context.Background(), &PresignInput{
		UserID: "u1", ChannelID: "ch1", FileName: "demo.mp4", MimeType: "video/mp4", SizeBytes: 1024, Thumbnail: thumbnail,
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	saved := repo.created.Media.Thumbnail
	if saved == nil || saved.StorageKey != repo.created.StorageKey+"-thumbnail" || saved.Width != 640 || saved.Height != 360 {
		t.Fatalf("サムネイルが保存されていません: %+v", saved)
	}
	if out.ThumbnailUploadURL == nil || !strings.HasSuffix(*out.ThumbnailUploadURL, saved.StorageKey) {
		t.Errorf("サムネイルのアップロード URL がありません: %v", out.ThumbnailUploadURL)
	}

	_, err = interactor.Presign(context.Background(), &PresignInput{
		UserID: "u1", ChannelID: "ch1", FileName: "a.png", MimeType: "image/png", SizeBytes: 1024, Thumbnail: thumbnail,
	})
	if !errors.Is(err, ErrThumbnailNotAllowed) {
		t.Errorf("動画以外のサムネイルを受け付けています: %v", err)
	}
}

func TestGetDownloadURLThumbnail(t *testing.T) {
	repo := &stubAttachmentRepo{found: &entity.Attachment{ID: "a1", ChannelID: "ch1", StorageKey: "attachments/ch1/a1"}}
	interactor := NewInteractor(repo, nil, stubChannelAccess{}, &stubStorage{}, stubStorageConfig{})

	if _, err := interactor.GetDownloadURL(context.Background(), "u1", "a1", true); !errors.Is(err, ErrThumbnailNotFound) {
		t.Errorf("サムネイルがないときのエラー = %v", err)
	}

	repo.found.Media.Thumbnail = &entity.Thumbnail{StorageKey: "attachments/ch1/a1-thumbnail", Width: 1, Height: 1}
	out, err := interactor.GetDownloadURL(context.Background(), "u1", "a1", true)
	if err != nil || !strings.HasSuffix(out.URL, "a1-thumbnail") {
		t.Errorf("サムネイルの URL = %v, %v", out, err)
	}
	out, err = interactor.GetDownloadURL(context.Background(), "u1", "a1", false)
	if err != nil || !strings.HasSuffix(out.URL, "attachments/ch1/a1") {
		t.Errorf("本体の URL = %v, %v", out, err)
	}
}
