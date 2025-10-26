package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

// User represents the GORM model for the users table.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"type:text;uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:text;not null"`
	DisplayName  string    `gorm:"type:text;not null"`
	AvatarURL    *string   `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (User) TableName() string {
	return "users"
}

func (m *User) FromEntity(e *entity.User) {
	if e == nil {
		*m = User{}
		return
	}

	userID, _ := utils.ParseUUID(e.ID, "user ID")
	*m = User{
		ID:           userID,
		Email:        e.Email,
		PasswordHash: e.PasswordHash,
		DisplayName:  e.DisplayName,
		AvatarURL:    e.AvatarURL,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func (m *User) ToEntity() *entity.User {
	if m == nil {
		return nil
	}

	return &entity.User{
		ID:           utils.UUIDToString(m.ID),
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		DisplayName:  m.DisplayName,
		AvatarURL:    m.AvatarURL,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// Session represents the sessions table.
type Session struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	RefreshTokenHash string     `gorm:"type:text;not null"`
	ExpiresAt        time.Time  `gorm:"type:timestamptz;not null;index"`
	RevokedAt        *time.Time `gorm:"type:timestamptz"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (Session) TableName() string {
	return "sessions"
}

func (m *Session) FromEntity(e *entity.Session) {
	if e == nil {
		*m = Session{}
		return
	}

	sessionID, _ := utils.ParseUUID(e.ID, "session ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: e.RefreshTokenHash,
		ExpiresAt:        e.ExpiresAt,
		RevokedAt:        e.RevokedAt,
		CreatedAt:        e.CreatedAt,
	}
}

func (m *Session) ToEntity() *entity.Session {
	if m == nil {
		return nil
	}

	return &entity.Session{
		ID:               utils.UUIDToString(m.ID),
		UserID:           utils.UUIDToString(m.UserID),
		RefreshTokenHash: m.RefreshTokenHash,
		ExpiresAt:        m.ExpiresAt,
		RevokedAt:        m.RevokedAt,
		CreatedAt:        m.CreatedAt,
	}
}

// Workspace represents the workspaces table.
type Workspace struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"type:text;not null"`
	Description *string   `gorm:"type:text"`
	IconURL     *string   `gorm:"type:text"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

func (m *Workspace) FromEntity(e *entity.Workspace) {
	if e == nil {
		*m = Workspace{}
		return
	}

	workspaceID, _ := utils.ParseUUID(e.ID, "workspace ID")
	createdByID, _ := utils.ParseUUID(e.CreatedBy, "created by ID")
	*m = Workspace{
		ID:          workspaceID,
		Name:        e.Name,
		Description: e.Description,
		IconURL:     e.IconURL,
		CreatedBy:   createdByID,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (m *Workspace) ToEntity() *entity.Workspace {
	if m == nil {
		return nil
	}

	return &entity.Workspace{
		ID:          utils.UUIDToString(m.ID),
		Name:        m.Name,
		Description: m.Description,
		IconURL:     m.IconURL,
		CreatedBy:   utils.UUIDToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// WorkspaceMember represents the workspace_members table.
type WorkspaceMember struct {
	WorkspaceID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Role        string    `gorm:"type:text;not null"`
	JoinedAt    time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

func (m *WorkspaceMember) FromEntity(e *entity.WorkspaceMember) {
	if e == nil {
		*m = WorkspaceMember{}
		return
	}

	workspaceID, _ := utils.ParseUUID(e.WorkspaceID, "workspace ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        string(e.Role),
		JoinedAt:    e.JoinedAt,
	}
}

func (m *WorkspaceMember) ToEntity() *entity.WorkspaceMember {
	if m == nil {
		return nil
	}

	return &entity.WorkspaceMember{
		WorkspaceID: utils.UUIDToString(m.WorkspaceID),
		UserID:      utils.UUIDToString(m.UserID),
		Role:        entity.WorkspaceRole(m.Role),
		JoinedAt:    m.JoinedAt,
	}
}

// Channel represents the channels table.
type Channel struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_channels_workspace_name"`
	Name        string    `gorm:"type:text;not null;uniqueIndex:idx_channels_workspace_name"`
	Description *string   `gorm:"type:text"`
	IsPrivate   bool      `gorm:"type:boolean;not null;default:false;index:idx_workspace_private"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Channel) TableName() string {
	return "channels"
}

func (m *Channel) FromEntity(e *entity.Channel) {
	if e == nil {
		*m = Channel{}
		return
	}

	channelID, _ := utils.ParseUUID(e.ID, "channel ID")
	workspaceID, _ := utils.ParseUUID(e.WorkspaceID, "workspace ID")
	createdByID, _ := utils.ParseUUID(e.CreatedBy, "created by ID")
	*m = Channel{
		ID:          channelID,
		WorkspaceID: workspaceID,
		Name:        e.Name,
		Description: e.Description,
		IsPrivate:   e.IsPrivate,
		CreatedBy:   createdByID,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (m *Channel) ToEntity() *entity.Channel {
	if m == nil {
		return nil
	}

	return &entity.Channel{
		ID:          utils.UUIDToString(m.ID),
		WorkspaceID: utils.UUIDToString(m.WorkspaceID),
		Name:        m.Name,
		Description: m.Description,
		IsPrivate:   m.IsPrivate,
		CreatedBy:   utils.UUIDToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ChannelMember represents the channel_members table.
type ChannelMember struct {
	ChannelID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Role      string    `gorm:"type:text;not null;default:member"`
	JoinedAt  time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (ChannelMember) TableName() string {
	return "channel_members"
}

func (m *ChannelMember) FromEntity(e *entity.ChannelMember) {
	if e == nil {
		*m = ChannelMember{}
		return
	}

	channelID, _ := utils.ParseUUID(e.ChannelID, "channel ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
		Role:      string(e.Role),
		JoinedAt:  e.JoinedAt,
	}
}

func (m *ChannelMember) ToEntity() *entity.ChannelMember {
	if m == nil {
		return nil
	}

	return &entity.ChannelMember{
		ChannelID: utils.UUIDToString(m.ChannelID),
		UserID:    utils.UUIDToString(m.UserID),
		Role:      entity.ChannelRole(m.Role),
		JoinedAt:  m.JoinedAt,
	}
}

// Message represents the messages table.
type Message struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ChannelID uuid.UUID  `gorm:"type:uuid;not null;index:idx_channel_created"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	ParentID  *uuid.UUID `gorm:"type:uuid;index:idx_parent_created"`
	Body      string     `gorm:"type:text;not null"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;default:now();index:idx_channel_created,idx_parent_created"`
	EditedAt  *time.Time `gorm:"type:timestamptz"`
	DeletedAt *time.Time `gorm:"type:timestamptz"`
	DeletedBy *uuid.UUID `gorm:"type:uuid"`
}

func (Message) TableName() string {
	return "messages"
}

func (m *Message) FromEntity(e *entity.Message) {
	if e == nil {
		*m = Message{}
		return
	}

	messageID, _ := utils.ParseUUID(e.ID, "message ID")
	channelID, _ := utils.ParseUUID(e.ChannelID, "channel ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = Message{
		ID:        messageID,
		ChannelID: channelID,
		UserID:    userID,
		ParentID:  utils.ParseUUIDPtr(e.ParentID),
		Body:      e.Body,
		CreatedAt: e.CreatedAt,
		EditedAt:  e.EditedAt,
		DeletedAt: e.DeletedAt,
		DeletedBy: utils.ParseUUIDPtr(e.DeletedBy),
	}
}

func (m *Message) ToEntity() *entity.Message {
	if m == nil {
		return nil
	}

	return &entity.Message{
		ID:        utils.UUIDToString(m.ID),
		ChannelID: utils.UUIDToString(m.ChannelID),
		UserID:    utils.UUIDToString(m.UserID),
		ParentID:  utils.UUIDPtrToStringPtr(m.ParentID),
		Body:      m.Body,
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		DeletedAt: m.DeletedAt,
		DeletedBy: utils.UUIDPtrToStringPtr(m.DeletedBy),
	}
}

// MessageReaction represents the message_reactions table.
type MessageReaction struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Emoji     string    `gorm:"type:text;primaryKey"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageReaction) TableName() string {
	return "message_reactions"
}

func (m *MessageReaction) FromEntity(e *entity.MessageReaction) {
	if e == nil {
		*m = MessageReaction{}
		return
	}

	messageID, _ := utils.ParseUUID(e.MessageID, "message ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     e.Emoji,
		CreatedAt: e.CreatedAt,
	}
}

func (m *MessageReaction) ToEntity() *entity.MessageReaction {
	if m == nil {
		return nil
	}

	return &entity.MessageReaction{
		MessageID: utils.UUIDToString(m.MessageID),
		UserID:    utils.UUIDToString(m.UserID),
		Emoji:     m.Emoji,
		CreatedAt: m.CreatedAt,
	}
}

// MessageBookmark represents the message_bookmarks table.
type MessageBookmark struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageBookmark) TableName() string {
	return "message_bookmarks"
}

func (m *MessageBookmark) FromEntity(e *entity.MessageBookmark) {
	if e == nil {
		*m = MessageBookmark{}
		return
	}

	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	messageID, _ := utils.ParseUUID(e.MessageID, "message ID")
	*m = MessageBookmark{
		UserID:    userID,
		MessageID: messageID,
		CreatedAt: e.CreatedAt,
	}
}

func (m *MessageBookmark) ToEntity() *entity.MessageBookmark {
	if m == nil {
		return nil
	}

	return &entity.MessageBookmark{
		UserID:    utils.UUIDToString(m.UserID),
		MessageID: utils.UUIDToString(m.MessageID),
		CreatedAt: m.CreatedAt,
	}
}

// ChannelReadState represents the channel_read_states table.
type ChannelReadState struct {
	ChannelID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_user_last_read"`
	LastReadAt time.Time `gorm:"type:timestamptz;not null;default:now();index:idx_user_last_read"`
}

func (ChannelReadState) TableName() string {
	return "channel_read_states"
}

func (m *ChannelReadState) FromEntity(e *entity.ChannelReadState) {
	if e == nil {
		*m = ChannelReadState{}
		return
	}

	channelID, _ := utils.ParseUUID(e.ChannelID, "channel ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = ChannelReadState{
		ChannelID:  channelID,
		UserID:     userID,
		LastReadAt: e.LastReadAt,
	}
}

func (m *ChannelReadState) ToEntity() *entity.ChannelReadState {
	if m == nil {
		return nil
	}

	return &entity.ChannelReadState{
		ChannelID:  utils.UUIDToString(m.ChannelID),
		UserID:     utils.UUIDToString(m.UserID),
		LastReadAt: m.LastReadAt,
	}
}

// Attachment represents the attachments table.
type Attachment struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MessageID  *uuid.UUID `gorm:"type:uuid;index"`
	UploaderID uuid.UUID  `gorm:"type:uuid;not null;index:idx_uploader_status"`
	ChannelID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	FileName   string     `gorm:"type:text;not null"`
	MimeType   string     `gorm:"type:text;not null"`
	SizeBytes  int64      `gorm:"type:bigint;not null"`
	StorageKey string     `gorm:"type:text;not null"`
	Status     string     `gorm:"type:text;not null;default:pending;index:idx_uploader_status"`
	UploadedAt *time.Time `gorm:"type:timestamptz"`
	ExpiresAt  *time.Time `gorm:"type:timestamptz"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (Attachment) TableName() string {
	return "attachments"
}

func (m *Attachment) FromEntity(e *entity.Attachment) {
	if e == nil {
		*m = Attachment{}
		return
	}

	attachmentID, _ := utils.ParseUUID(e.ID, "attachment ID")
	uploaderID, _ := utils.ParseUUID(e.UploaderID, "uploader ID")
	channelID, _ := utils.ParseUUID(e.ChannelID, "channel ID")
	*m = Attachment{
		ID:         attachmentID,
		MessageID:  utils.ParseUUIDPtr(e.MessageID),
		UploaderID: uploaderID,
		ChannelID:  channelID,
		FileName:   e.FileName,
		MimeType:   e.MimeType,
		SizeBytes:  e.SizeBytes,
		StorageKey: e.StorageKey,
		Status:     string(e.Status),
		UploadedAt: e.UploadedAt,
		ExpiresAt:  e.ExpiresAt,
		CreatedAt:  e.CreatedAt,
	}
}

func (m *Attachment) ToEntity() *entity.Attachment {
	if m == nil {
		return nil
	}

	return &entity.Attachment{
		ID:         utils.UUIDToString(m.ID),
		MessageID:  utils.UUIDPtrToStringPtr(m.MessageID),
		UploaderID: utils.UUIDToString(m.UploaderID),
		ChannelID:  utils.UUIDToString(m.ChannelID),
		FileName:   m.FileName,
		MimeType:   m.MimeType,
		SizeBytes:  m.SizeBytes,
		StorageKey: m.StorageKey,
		Status:     entity.AttachmentStatus(m.Status),
		UploadedAt: m.UploadedAt,
		ExpiresAt:  m.ExpiresAt,
		CreatedAt:  m.CreatedAt,
	}
}

// UserGroup represents the user_groups table.
type UserGroup struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_workspace_name"`
	Name        string    `gorm:"type:text;not null;uniqueIndex:idx_workspace_name"`
	Description *string   `gorm:"type:text"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (UserGroup) TableName() string {
	return "user_groups"
}

func (m *UserGroup) FromEntity(e *entity.UserGroup) {
	if e == nil {
		*m = UserGroup{}
		return
	}

	groupID, _ := utils.ParseUUID(e.ID, "group ID")
	workspaceID, _ := utils.ParseUUID(e.WorkspaceID, "workspace ID")
	createdByID, _ := utils.ParseUUID(e.CreatedBy, "created by ID")
	*m = UserGroup{
		ID:          groupID,
		WorkspaceID: workspaceID,
		Name:        e.Name,
		Description: e.Description,
		CreatedBy:   createdByID,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (m *UserGroup) ToEntity() *entity.UserGroup {
	if m == nil {
		return nil
	}

	return &entity.UserGroup{
		ID:          utils.UUIDToString(m.ID),
		WorkspaceID: utils.UUIDToString(m.WorkspaceID),
		Name:        m.Name,
		Description: m.Description,
		CreatedBy:   utils.UUIDToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// UserGroupMember represents the user_group_members table.
type UserGroupMember struct {
	GroupID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	JoinedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (UserGroupMember) TableName() string {
	return "user_group_members"
}

func (m *UserGroupMember) FromEntity(e *entity.UserGroupMember) {
	if e == nil {
		*m = UserGroupMember{}
		return
	}

	groupID, _ := utils.ParseUUID(e.GroupID, "group ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = UserGroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: e.JoinedAt,
	}
}

func (m *UserGroupMember) ToEntity() *entity.UserGroupMember {
	if m == nil {
		return nil
	}

	return &entity.UserGroupMember{
		GroupID:  utils.UUIDToString(m.GroupID),
		UserID:   utils.UUIDToString(m.UserID),
		JoinedAt: m.JoinedAt,
	}
}

// MessageUserMention represents the message_user_mentions table.
type MessageUserMention struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageUserMention) TableName() string {
	return "message_user_mentions"
}

func (m *MessageUserMention) FromEntity(e *entity.MessageUserMention) {
	if e == nil {
		*m = MessageUserMention{}
		return
	}

	messageID, _ := utils.ParseUUID(e.MessageID, "message ID")
	userID, _ := utils.ParseUUID(e.UserID, "user ID")
	*m = MessageUserMention{
		MessageID: messageID,
		UserID:    userID,
		CreatedAt: e.CreatedAt,
	}
}

func (m *MessageUserMention) ToEntity() *entity.MessageUserMention {
	if m == nil {
		return nil
	}

	return &entity.MessageUserMention{
		MessageID: utils.UUIDToString(m.MessageID),
		UserID:    utils.UUIDToString(m.UserID),
		CreatedAt: m.CreatedAt,
	}
}

// MessageGroupMention represents the message_group_mentions table.
type MessageGroupMention struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageGroupMention) TableName() string {
	return "message_group_mentions"
}

func (m *MessageGroupMention) FromEntity(e *entity.MessageGroupMention) {
	if e == nil {
		*m = MessageGroupMention{}
		return
	}

	messageID, _ := utils.ParseUUID(e.MessageID, "message ID")
	groupID, _ := utils.ParseUUID(e.GroupID, "group ID")
	*m = MessageGroupMention{
		MessageID: messageID,
		GroupID:   groupID,
		CreatedAt: e.CreatedAt,
	}
}

func (m *MessageGroupMention) ToEntity() *entity.MessageGroupMention {
	if m == nil {
		return nil
	}

	return &entity.MessageGroupMention{
		MessageID: utils.UUIDToString(m.MessageID),
		GroupID:   utils.UUIDToString(m.GroupID),
		CreatedAt: m.CreatedAt,
	}
}

// MessageLink represents the message_links table.
type MessageLink struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MessageID   uuid.UUID `gorm:"type:uuid;not null;index"`
	URL         string    `gorm:"type:text;not null;uniqueIndex"`
	Title       *string   `gorm:"type:text"`
	Description *string   `gorm:"type:text"`
	ImageURL    *string   `gorm:"type:text"`
	SiteName    *string   `gorm:"type:text"`
	CardType    *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageLink) TableName() string {
	return "message_links"
}

func (m *MessageLink) FromEntity(e *entity.MessageLink) {
	if e == nil {
		*m = MessageLink{}
		return
	}

	linkID, _ := utils.ParseUUID(e.ID, "link ID")
	messageID, _ := utils.ParseUUID(e.MessageID, "message ID")
	*m = MessageLink{
		ID:          linkID,
		MessageID:   messageID,
		URL:         e.URL,
		Title:       e.Title,
		Description: e.Description,
		ImageURL:    e.ImageURL,
		SiteName:    e.SiteName,
		CardType:    e.CardType,
		CreatedAt:   e.CreatedAt,
	}
}

func (m *MessageLink) ToEntity() *entity.MessageLink {
	if m == nil {
		return nil
	}

	return &entity.MessageLink{
		ID:          utils.UUIDToString(m.ID),
		MessageID:   utils.UUIDToString(m.MessageID),
		URL:         m.URL,
		Title:       m.Title,
		Description: m.Description,
		ImageURL:    m.ImageURL,
		SiteName:    m.SiteName,
		CardType:    m.CardType,
		CreatedAt:   m.CreatedAt,
	}
}

// ThreadMetadata represents the thread_metadata table.
type ThreadMetadata struct {
	MessageID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	ReplyCount         int         `gorm:"type:integer;not null;default:0"`
	LastReplyAt        *time.Time  `gorm:"type:timestamptz;index"`
	LastReplyUserID    *uuid.UUID  `gorm:"type:uuid"`
	ParticipantUserIDs []uuid.UUID `gorm:"type:uuid[];not null;default:'{}'"`
	CreatedAt          time.Time   `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt          time.Time   `gorm:"type:timestamptz;not null;default:now()"`
}

func (ThreadMetadata) TableName() string {
	return "thread_metadata"
}

func (m *ThreadMetadata) FromEntity(e *entity.ThreadMetadata) {
	if e == nil {
		*m = ThreadMetadata{}
		return
	}

	participantIDs := make([]uuid.UUID, 0, len(e.ParticipantUserIDs))
	for _, id := range e.ParticipantUserIDs {
		participantID, _ := utils.ParseUUID(id, "participant ID")
		participantIDs = append(participantIDs, participantID)
	}

	messageID, _ := utils.ParseUUID(e.MessageID, "message ID")
	*m = ThreadMetadata{
		MessageID:          messageID,
		ReplyCount:         e.ReplyCount,
		LastReplyAt:        e.LastReplyAt,
		LastReplyUserID:    utils.ParseUUIDPtr(e.LastReplyUserID),
		ParticipantUserIDs: participantIDs,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}

func (m *ThreadMetadata) ToEntity() *entity.ThreadMetadata {
	if m == nil {
		return nil
	}

	participantIDs := make([]string, 0, len(m.ParticipantUserIDs))
	for _, id := range m.ParticipantUserIDs {
		participantIDs = append(participantIDs, utils.UUIDToString(id))
	}

	return &entity.ThreadMetadata{
		MessageID:          utils.UUIDToString(m.MessageID),
		ReplyCount:         m.ReplyCount,
		LastReplyAt:        m.LastReplyAt,
		LastReplyUserID:    utils.UUIDPtrToStringPtr(m.LastReplyUserID),
		ParticipantUserIDs: participantIDs,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}
