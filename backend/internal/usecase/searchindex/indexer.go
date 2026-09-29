package searchindex

import (
	"context"
	"fmt"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

const reindexBatchSize = 500

// Indexer はメッセージの現在の状態を全文検索インデックスへ反映します
type Indexer struct {
	messageRepo domainrepository.MessageRepository
	index       domainrepository.MessageSearchIndex
	logger      service.Logger
}

func NewIndexer(messageRepo domainrepository.MessageRepository, index domainrepository.MessageSearchIndex, logger service.Logger) *Indexer {
	return &Indexer{messageRepo: messageRepo, index: index, logger: logger}
}

// Sync は指定したメッセージを登録し直し、削除済みのものはインデックスから外します。
// 検索は補助機能のため、失敗してもログに残すだけで呼び出し元の処理は失敗させない
func (i *Indexer) Sync(ctx context.Context, messageIDs ...string) {
	if err := i.sync(context.WithoutCancel(ctx), messageIDs); err != nil {
		i.logger.Warn("検索インデックスの更新に失敗しました", service.LogField{Key: "messageIDs", Value: messageIDs}, service.LogField{Key: "error", Value: err.Error()})
	}
}

func (i *Indexer) sync(ctx context.Context, messageIDs []string) error {
	docs, err := i.messageRepo.FindSearchDocuments(ctx, messageIDs)
	if err != nil {
		return fmt.Errorf("failed to load search documents: %w", err)
	}
	if err := i.index.Upsert(ctx, docs); err != nil {
		return fmt.Errorf("failed to upsert documents: %w", err)
	}

	live := make(map[string]bool, len(docs))
	for _, d := range docs {
		live[d.ID] = true
	}
	removed := []string{}
	for _, id := range messageIDs {
		if !live[id] {
			removed = append(removed, id)
		}
	}
	if err := i.index.Delete(ctx, removed); err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	return nil
}

// Reindex はインデックスを空にし、削除されていない全メッセージを登録し直して件数を返します
func (i *Indexer) Reindex(ctx context.Context) (int, error) {
	if err := i.index.DeleteAll(ctx); err != nil {
		return 0, fmt.Errorf("failed to clear index: %w", err)
	}
	total, after := 0, ""
	for {
		docs, err := i.messageRepo.FindSearchDocumentsAfter(ctx, after, reindexBatchSize)
		if err != nil {
			return total, fmt.Errorf("failed to load search documents: %w", err)
		}
		if len(docs) == 0 {
			return total, nil
		}
		if err := i.index.Upsert(ctx, docs); err != nil {
			return total, fmt.Errorf("failed to upsert documents: %w", err)
		}
		total += len(docs)
		after = docs[len(docs)-1].ID
	}
}
