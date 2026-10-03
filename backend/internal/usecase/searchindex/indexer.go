package searchindex

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

// syncTimeout は呼び出し元が終わったあとも続ける検索インデックスの更新の上限時間
const syncTimeout = 10 * time.Second

const reindexBatchSize = 500

// Indexer はメッセージの現在の状態を全文検索インデックスへ反映します
type Indexer struct {
	messageRepo domainrepository.MessageRepository
	index       domainrepository.MessageSearchIndex
	mentionSvc  service.MentionService
}

func NewIndexer(messageRepo domainrepository.MessageRepository, index domainrepository.MessageSearchIndex, mentionSvc service.MentionService) *Indexer {
	return &Indexer{messageRepo: messageRepo, index: index, mentionSvc: mentionSvc}
}

// upsert は本文の ID 記法を名前に置き換え、名前で検索できるようにして登録します
func (i *Indexer) upsert(ctx context.Context, docs []domainrepository.MessageSearchDocument) error {
	for j := range docs {
		body, err := i.mentionSvc.RenderPlain(ctx, docs[j].Body)
		if err != nil {
			return fmt.Errorf("failed to render message body: %w", err)
		}
		docs[j].Body = body
	}
	return i.index.Upsert(ctx, docs)
}

// Sync は指定したメッセージを登録し直し、削除済みのものは外します。検索は補助機能のため失敗してもログに残すだけにする
func (i *Indexer) Sync(ctx context.Context, messageIDs ...string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), syncTimeout)
	defer cancel()
	if err := i.sync(ctx, messageIDs); err != nil {
		slog.WarnContext(ctx, "検索インデックスの更新に失敗しました", "messageIDs", messageIDs, "error", err)
	}
}

func (i *Indexer) sync(ctx context.Context, messageIDs []string) error {
	docs, err := i.messageRepo.FindSearchDocuments(ctx, messageIDs)
	if err != nil {
		return fmt.Errorf("failed to load search documents: %w", err)
	}
	if err := i.upsert(ctx, docs); err != nil {
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

// Prepare は検索に必要な設定を反映し、インデックスが空か force なら削除されていない全メッセージを登録し直して件数を返します
func (i *Indexer) Prepare(ctx context.Context, force bool) (int, error) {
	if err := i.index.EnsureSettings(ctx); err != nil {
		return 0, fmt.Errorf("failed to configure index: %w", err)
	}
	if !force {
		empty, err := i.index.IsEmpty(ctx)
		if err != nil || !empty {
			return 0, err
		}
	}
	return i.reindex(ctx)
}

func (i *Indexer) reindex(ctx context.Context) (int, error) {
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
		if err := i.upsert(ctx, docs); err != nil {
			return total, fmt.Errorf("failed to upsert documents: %w", err)
		}
		total += len(docs)
		after = docs[len(docs)-1].ID
	}
}
