package searchindex

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type stubMessageRepo struct {
	domainrepository.MessageRepository
	docs []domainrepository.MessageSearchDocument
}

func (r *stubMessageRepo) FindSearchDocuments(_ context.Context, ids []string) ([]domainrepository.MessageSearchDocument, error) {
	found := []domainrepository.MessageSearchDocument{}
	for _, d := range r.docs {
		for _, id := range ids {
			if d.ID == id {
				found = append(found, d)
			}
		}
	}
	return found, nil
}

func (r *stubMessageRepo) FindSearchDocumentsAfter(_ context.Context, afterID string, limit int) ([]domainrepository.MessageSearchDocument, error) {
	found := []domainrepository.MessageSearchDocument{}
	for _, d := range r.docs {
		if d.ID > afterID && len(found) < limit {
			found = append(found, d)
		}
	}
	return found, nil
}

type stubIndex struct {
	upserted []string
	deleted  []string
	cleared  bool
	empty    bool
}

func (i *stubIndex) EnsureSettings(context.Context) error { return nil }

func (i *stubIndex) IsEmpty(context.Context) (bool, error) { return i.empty, nil }

func (i *stubIndex) Search(context.Context, domainrepository.MessageSearchCriteria) (*domainrepository.MessageSearchResult, error) {
	return nil, errors.New("not implemented")
}

func (i *stubIndex) Upsert(_ context.Context, docs []domainrepository.MessageSearchDocument) error {
	for _, d := range docs {
		i.upserted = append(i.upserted, d.ID)
	}
	return nil
}

func (i *stubIndex) Delete(_ context.Context, ids []string) error {
	i.deleted = append(i.deleted, ids...)
	return nil
}

func (i *stubIndex) DeleteAll(context.Context) error {
	i.cleared = true
	return nil
}

func docs(ids ...string) []domainrepository.MessageSearchDocument {
	result := []domainrepository.MessageSearchDocument{}
	for _, id := range ids {
		result = append(result, domainrepository.MessageSearchDocument{ID: id})
	}
	return result
}

func TestSyncUpsertsLiveAndDeletesRemoved(t *testing.T) {
	index := &stubIndex{}
	indexer := NewIndexer(&stubMessageRepo{docs: docs("a", "b")}, index, stubMentionService{})

	indexer.Sync(context.Background(), "a", "deleted", "b")

	if !reflect.DeepEqual(index.upserted, []string{"a", "b"}) {
		t.Errorf("登録した文書が期待と異なります: %v", index.upserted)
	}
	if !reflect.DeepEqual(index.deleted, []string{"deleted"}) {
		t.Errorf("削除した文書が期待と異なります: %v", index.deleted)
	}
}

func TestPrepareSkipsNonEmptyIndex(t *testing.T) {
	index := &stubIndex{}
	indexer := NewIndexer(&stubMessageRepo{docs: docs("a")}, index, stubMentionService{})

	if count, err := indexer.Prepare(context.Background(), false); err != nil || count != 0 || index.cleared {
		t.Errorf("空でないインデックスは登録し直さないはず: count=%d err=%v", count, err)
	}
}

func TestPrepareRegistersAllInBatches(t *testing.T) {
	ids := []string{}
	for i := range reindexBatchSize + 3 {
		ids = append(ids, fmt.Sprintf("%04d", i))
	}
	index := &stubIndex{}
	indexer := NewIndexer(&stubMessageRepo{docs: docs(ids...)}, index, stubMentionService{})

	count, err := indexer.Prepare(context.Background(), true)
	if err != nil {
		t.Fatalf("再インデックスに失敗しました: %v", err)
	}
	if !index.cleared || count != len(ids) || !reflect.DeepEqual(index.upserted, ids) {
		t.Errorf("全件が登録されていません: cleared=%v count=%d upserted=%d", index.cleared, count, len(index.upserted))
	}
}

type stubMentionService struct {
	service.MentionService
}

func (stubMentionService) RenderPlain(_ context.Context, body string) (string, error) {
	return body, nil
}
