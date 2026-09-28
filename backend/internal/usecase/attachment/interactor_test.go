package attachment

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubAttachmentRepo struct {
	domainrepository.AttachmentRepository
	created *entity.Attachment
}

func (r *stubAttachmentRepo) CreatePending(_ context.Context, attachment *entity.Attachment) error {
	r.created = attachment
	return nil
}

type stubChannelAccess struct{}

func (stubChannelAccess) EnsureChannelAccess(_ context.Context, channelID string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: channelID}, nil
}

type stubStorage struct {
	uploadedMimeType string
}

func (s *stubStorage) GenerateUploadURL(_ string, mimeType string, _ int64, _ interface{}) (string, error) {
	s.uploadedMimeType = mimeType
	return "https://storage.example.com/upload", nil
}

func (s *stubStorage) GenerateDownloadURL(_ string, _ interface{}) (string, error) {
	return "", nil
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
