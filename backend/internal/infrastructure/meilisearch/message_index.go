package meilisearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	meili "github.com/meilisearch/meilisearch-go"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

const messageIndexUID = "messages"

var primaryKey = "id"

// 本文と添付ファイル名を日本語として分かち書きする
var localizedAttributes = []*meili.LocalizedAttributes{{AttributePatterns: []string{"body", "attachment_names"}, Locales: []string{"jpn"}}}

// sort を先頭に置き、新しい順の指定を関連度より優先する。sort を指定しなければ関連度順になる
var messageIndexSettings = &meili.Settings{
	RankingRules:         []string{"sort", "words", "typo", "proximity", "attribute", "exactness"},
	SearchableAttributes: []string{"body", "attachment_names"},
	FilterableAttributes: []string{
		"workspace_id", "channel_id", "sender_id", "parent_id", "has", "has_replies", "pinned",
		"mentioned_user_ids", "mentions_channel", "created_at",
	},
	SortableAttributes:  []string{"created_at"},
	LocalizedAttributes: localizedAttributes,
	Pagination:          &meili.Pagination{MaxTotalHits: 10000},
}

type MessageIndex struct {
	index meili.IndexManager
}

var _ domainrepository.MessageSearchIndex = (*MessageIndex)(nil)

func NewMessageIndex(url, apiKey string) *MessageIndex {
	return &MessageIndex{index: meili.New(url, meili.WithAPIKey(apiKey)).Index(messageIndexUID)}
}

// messageDocument はインデックスに保存する形です。日時はミリ秒の UNIX 時刻で持ち、範囲で絞り込む
type messageDocument struct {
	ID               string                                `json:"id"`
	WorkspaceID      string                                `json:"workspace_id"`
	ChannelID        string                                `json:"channel_id"`
	SenderID         string                                `json:"sender_id"`
	ParentID         *string                               `json:"parent_id"`
	Body             string                                `json:"body"`
	AttachmentNames  []string                              `json:"attachment_names"`
	Has              []domainrepository.MessageContentKind `json:"has"`
	MentionedUserIDs []string                              `json:"mentioned_user_ids"`
	MentionsChannel  bool                                  `json:"mentions_channel"`
	Pinned           bool                                  `json:"pinned"`
	HasReplies       bool                                  `json:"has_replies"`
	CreatedAt        int64                                 `json:"created_at"`
}

func (i *MessageIndex) EnsureSettings(ctx context.Context) error {
	task, err := i.index.UpdateSettingsWithContext(ctx, messageIndexSettings)
	if err != nil {
		return err
	}
	return i.wait(ctx, task)
}

func (i *MessageIndex) IsEmpty(ctx context.Context) (bool, error) {
	stats, err := i.index.GetStatsWithContext(ctx, nil)
	if err != nil {
		return false, err
	}
	return stats.NumberOfDocuments == 0, nil
}

func (i *MessageIndex) Search(ctx context.Context, c domainrepository.MessageSearchCriteria) (*domainrepository.MessageSearchResult, error) {
	perPage := int64(c.PerPage)
	req := &meili.SearchRequest{
		Filter:               buildFilter(c),
		MatchingStrategy:     meili.All,
		AttributesToRetrieve: []string{"id"},
		Page:                 int64(c.Page),
		HitsPerPage:          &perPage,
		Locales:              []string{"jpn"},
	}
	if c.Sort != domainrepository.MessageSearchSortRelevance {
		req.Sort = []string{"created_at:desc"}
	}
	res, err := i.index.SearchWithContext(ctx, strings.Join(c.Terms, " "), req)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(res.Hits))
	for _, hit := range res.Hits {
		var doc struct {
			ID string `json:"id"`
		}
		if err := hit.DecodeInto(&doc); err != nil {
			return nil, err
		}
		ids = append(ids, doc.ID)
	}
	return &domainrepository.MessageSearchResult{MessageIDs: ids, Total: int(res.TotalHits)}, nil
}

func (i *MessageIndex) Upsert(ctx context.Context, documents []domainrepository.MessageSearchDocument) error {
	if len(documents) == 0 {
		return nil
	}
	docs := make([]messageDocument, 0, len(documents))
	for _, d := range documents {
		docs = append(docs, messageDocument{
			ID:               d.ID,
			WorkspaceID:      d.WorkspaceID,
			ChannelID:        d.ChannelID,
			SenderID:         d.SenderID,
			ParentID:         d.ParentID,
			Body:             d.Body,
			AttachmentNames:  d.AttachmentNames,
			Has:              d.Has,
			MentionedUserIDs: d.MentionedUserIDs,
			MentionsChannel:  d.MentionsChannel,
			Pinned:           d.Pinned,
			HasReplies:       d.HasReplies,
			CreatedAt:        d.CreatedAt.UnixMilli(),
		})
	}
	_, err := i.index.AddDocumentsWithContext(ctx, docs, &meili.DocumentOptions{PrimaryKey: &primaryKey})
	return err
}

func (i *MessageIndex) Delete(ctx context.Context, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}
	_, err := i.index.DeleteDocumentsWithContext(ctx, messageIDs, nil)
	return err
}

func (i *MessageIndex) DeleteAll(ctx context.Context) error {
	_, err := i.index.DeleteAllDocumentsWithContext(ctx, nil)
	return err
}

func (i *MessageIndex) wait(ctx context.Context, info *meili.TaskInfo) error {
	task, err := i.index.WaitForTaskWithContext(ctx, info.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status != meili.TaskStatusSucceeded {
		return fmt.Errorf("meilisearch task %d %s: %s", task.UID, task.Status, task.Error.Message)
	}
	return nil
}

// buildFilter は条件を Meilisearch のフィルタ式にします
func buildFilter(c domainrepository.MessageSearchCriteria) string {
	filters := []string{
		"workspace_id = " + quote(c.WorkspaceID),
		"channel_id IN " + list(c.ChannelIDs),
	}
	if len(c.AuthorIDs) > 0 {
		filters = append(filters, "sender_id IN "+list(c.AuthorIDs))
	}
	for _, kind := range c.Has {
		filters = append(filters, "has = "+quote(string(kind)))
	}
	if c.PinnedOnly {
		filters = append(filters, "pinned = true")
	}
	if c.ThreadOnly {
		filters = append(filters, "(parent_id IS NOT NULL OR has_replies = true)")
	}
	if c.ExcludeReplies {
		filters = append(filters, "parent_id IS NULL")
	}
	if m := c.Mention; m != nil {
		mentions := []string{"mentioned_user_ids = " + quote(m.UserID)}
		if len(m.JoinedChannelIDs) > 0 {
			mentions = append(mentions, "(mentions_channel = true AND channel_id IN "+list(m.JoinedChannelIDs)+")")
		}
		filters = append(filters, "("+strings.Join(mentions, " OR ")+")")
	}
	if c.After != nil {
		filters = append(filters, fmt.Sprintf("created_at >= %d", c.After.UnixMilli()))
	}
	if c.Before != nil {
		filters = append(filters, fmt.Sprintf("created_at < %d", c.Before.UnixMilli()))
	}
	return strings.Join(filters, " AND ")
}

func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

func list(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = quote(v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
