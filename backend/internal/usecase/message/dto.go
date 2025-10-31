package message

import (
	"errors"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

var (
	ErrChannelNotFound       = errors.New("チャンネルが見つかりません")
	ErrUnauthorized          = errors.New("この操作を行う権限がありません")
	ErrParentMessageNotFound = errors.New("親メッセージが見つかりません")
	ErrMessageNotFound       = errors.New("メッセージが見つかりません")
	ErrMessageAlreadyDeleted = errors.New("メッセージは既に削除されています")
	ErrCannotEditDeleted     = errors.New("削除済みメッセージは編集できません")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type ListMessagesInput struct {
	ChannelID string
	UserID    string
	Limit     int
	Since     *time.Time
	Until     *time.Time
}

type CreateMessageInput struct {
	ChannelID     string
	UserID        string
	Body          string
	ParentID      *string
	AttachmentIDs []string
}

type UpdateMessageInput struct {
	MessageID string
	ChannelID string
	EditorID  string
	Body      string
}

type DeleteMessageInput struct {
	MessageID  string
	ChannelID  string
	ExecutorID string
}

type UserInfo struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type UserMention struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
}

type GroupMention struct {
	GroupID string `json:"groupId"`
	Name    string `json:"name"`
}

type LinkInfo struct {
	ID          string  `json:"id"`
	URL         string  `json:"url"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ImageURL    *string `json:"imageUrl"`
	SiteName    *string `json:"siteName"`
	CardType    *string `json:"cardType"`
}

type ReactionInfo struct {
	User      UserInfo  `json:"user"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
}

type AttachmentInfo struct {
	ID        string `json:"id"`
	FileName  string `json:"fileName"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

type MessageOutput struct {
	ID          string           `json:"id"`
	ChannelID   string           `json:"channelId"`
	UserID      string           `json:"userId"`
	User        UserInfo         `json:"user"`
	ParentID    *string          `json:"parentId"`
	Body        string           `json:"body"`
	Mentions    []UserMention    `json:"mentions"`
	Groups      []GroupMention   `json:"groups"`
	Links       []LinkInfo       `json:"links"`
	Reactions   []ReactionInfo   `json:"reactions"`
	Attachments []AttachmentInfo `json:"attachments"`
	CreatedAt   time.Time        `json:"createdAt"`
	EditedAt    *time.Time       `json:"editedAt"`
	DeletedAt   *time.Time       `json:"deletedAt"`
	IsDeleted   bool             `json:"isDeleted"`
	DeletedBy   *UserInfo        `json:"deletedBy,omitempty"`
}

type ListMessagesOutput struct {
    Messages []TimelineItem `json:"messages"`
	HasMore  bool            `json:"hasMore"`
}

type ThreadMetadataOutput struct {
	MessageID          string     `json:"messageId"`
	ReplyCount         int        `json:"replyCount"`
	LastReplyAt        *time.Time `json:"lastReplyAt"`
	LastReplyUser      *UserInfo  `json:"lastReplyUser"`
	ParticipantUserIDs []string   `json:"participantUserIds"`
}

type GetThreadRepliesInput struct {
	MessageID string
	UserID    string
	Limit     int
}

type GetThreadRepliesOutput struct {
	ParentMessage MessageOutput   `json:"parentMessage"`
	Replies       []MessageOutput `json:"replies"`
	HasMore       bool            `json:"hasMore"`
}

type GetThreadMetadataInput struct {
	MessageID string
	UserID    string
}

type MessageWithThreadOutput struct {
	MessageOutput
	ThreadMetadata *ThreadMetadataOutput `json:"threadMetadata,omitempty"`
}

// SystemMessageOutput はシステムメッセージの出力です
type SystemMessageOutput struct {
    ID        string                 `json:"id"`
    ChannelID string                 `json:"channelId"`
    Kind      string                 `json:"kind"`
    Payload   map[string]any         `json:"payload"`
    ActorID   *string                `json:"actorId,omitempty"`
    CreatedAt time.Time              `json:"createdAt"`
}

// TimelineItem はユーザー/システム両メッセージの統合タイムライン項目です
type TimelineItem struct {
    Type          string               `json:"type"` // "user" | "system"
    UserMessage   *MessageOutput       `json:"userMessage,omitempty"`
    SystemMessage *SystemMessageOutput `json:"systemMessage,omitempty"`
    CreatedAt     time.Time            `json:"createdAt"`
}

// RelatedData はメッセージに関連するデータをまとめた構造体です
type RelatedData struct {
	UserMentions  []*entity.MessageUserMention
	GroupMentions []*entity.MessageGroupMention
	Links         []*entity.MessageLink
	Reactions     map[string][]*entity.MessageReaction
	Attachments   map[string][]*entity.Attachment
}

// MessageOutputAssembler はMessageOutputの構築を担当するコンポーネントです
type MessageOutputAssembler struct{}

// NewMessageOutputAssembler は新しいMessageOutputAssemblerを作成します
func NewMessageOutputAssembler() *MessageOutputAssembler {
	return &MessageOutputAssembler{}
}

// AssembleMessageOutput はメッセージと関連データからMessageOutputを構築します
func (a *MessageOutputAssembler) AssembleMessageOutput(
	message *entity.Message,
	user *entity.User,
	userMentions []*entity.MessageUserMention,
	groupMentions []*entity.MessageGroupMention,
	links []*entity.MessageLink,
	reactions []*entity.MessageReaction,
	attachments []*entity.Attachment,
	groups map[string]*entity.UserGroup,
	userMap map[string]*entity.User,
) MessageOutput {
	userInfo := a.buildUserInfo(user)

	isDeleted := message.DeletedAt != nil

	var deletedByInfo *UserInfo
	if message.DeletedBy != nil {
		if deletedByUser, exists := userMap[*message.DeletedBy]; exists && deletedByUser != nil {
			deletedByInfo = &UserInfo{
				ID:          deletedByUser.ID,
				DisplayName: deletedByUser.DisplayName,
				AvatarURL:   deletedByUser.AvatarURL,
			}
		} else {
			deletedByInfo = &UserInfo{
				ID:          *message.DeletedBy,
				DisplayName: "Unknown User",
				AvatarURL:   nil,
			}
		}
	}

	return MessageOutput{
		ID:          message.ID,
		ChannelID:   message.ChannelID,
		UserID:      message.UserID,
		User:        userInfo,
		ParentID:    message.ParentID,
		Body:        message.Body,
		Mentions:    a.buildUserMentions(userMentions),
		Groups:      a.buildGroupMentions(groupMentions, groups),
		Links:       a.buildLinks(links),
		Reactions:   a.buildReactions(reactions, userMap),
		Attachments: a.buildAttachments(attachments),
		CreatedAt:   message.CreatedAt,
		EditedAt:    message.EditedAt,
		DeletedAt:   message.DeletedAt,
		IsDeleted:   isDeleted,
		DeletedBy:   deletedByInfo,
	}
}

// buildUserInfo はユーザー情報を構築します
func (a *MessageOutputAssembler) buildUserInfo(user *entity.User) UserInfo {
	if user == nil {
		return UserInfo{
			ID:          "",
			DisplayName: "Unknown User",
			AvatarURL:   nil,
		}
	}

	return UserInfo{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
	}
}

// buildUserMentions はユーザーメンションを構築します
func (a *MessageOutputAssembler) buildUserMentions(userMentions []*entity.MessageUserMention) []UserMention {
	mentions := make([]UserMention, 0, len(userMentions))
	for _, mention := range userMentions {
		mentions = append(mentions, UserMention{
			UserID:      mention.UserID,
			DisplayName: "", // 必要に応じてユーザー情報を取得
		})
	}
	return mentions
}

// buildGroupMentions はグループメンションを構築します
func (a *MessageOutputAssembler) buildGroupMentions(groupMentions []*entity.MessageGroupMention, groups map[string]*entity.UserGroup) []GroupMention {
	groupMentionsOutput := make([]GroupMention, 0, len(groupMentions))
	for _, mention := range groupMentions {
		groupName := ""
		if group, exists := groups[mention.GroupID]; exists {
			groupName = group.Name
		}
		groupMentionsOutput = append(groupMentionsOutput, GroupMention{
			GroupID: mention.GroupID,
			Name:    groupName,
		})
	}
	return groupMentionsOutput
}

// buildLinks はリンク情報を構築します
func (a *MessageOutputAssembler) buildLinks(links []*entity.MessageLink) []LinkInfo {
	linksOutput := make([]LinkInfo, 0, len(links))
	for _, link := range links {
		linksOutput = append(linksOutput, LinkInfo{
			ID:          link.ID,
			URL:         link.URL,
			Title:       link.Title,
			Description: link.Description,
			ImageURL:    link.ImageURL,
			SiteName:    link.SiteName,
			CardType:    link.CardType,
		})
	}
	return linksOutput
}

// buildReactions はリアクション情報を構築します
func (a *MessageOutputAssembler) buildReactions(reactions []*entity.MessageReaction, userMap map[string]*entity.User) []ReactionInfo {
	reactionsOutput := make([]ReactionInfo, 0, len(reactions))
	for _, reaction := range reactions {
		reactionUser, exists := userMap[reaction.UserID]
		reactionUserInfo := UserInfo{
			ID:          reaction.UserID,
			DisplayName: "Unknown User",
			AvatarURL:   nil,
		}
		if exists && reactionUser != nil {
			reactionUserInfo = UserInfo{
				ID:          reactionUser.ID,
				DisplayName: reactionUser.DisplayName,
				AvatarURL:   reactionUser.AvatarURL,
			}
		}
		reactionsOutput = append(reactionsOutput, ReactionInfo{
			User:      reactionUserInfo,
			Emoji:     reaction.Emoji,
			CreatedAt: reaction.CreatedAt,
		})
	}
	return reactionsOutput
}

// buildAttachments は添付ファイル情報を構築します
func (a *MessageOutputAssembler) buildAttachments(attachments []*entity.Attachment) []AttachmentInfo {
	attachmentsOutput := make([]AttachmentInfo, 0, len(attachments))
	for _, attachment := range attachments {
		attachmentsOutput = append(attachmentsOutput, AttachmentInfo{
			ID:        attachment.ID,
			FileName:  attachment.FileName,
			MimeType:  attachment.MimeType,
			SizeBytes: attachment.SizeBytes,
		})
	}
	return attachmentsOutput
}
