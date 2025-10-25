package database

import (
	"time"

	"github.com/google/uuid"

	"github.com/example/chat/internal/domain/entity"
)

func parseUUID(id string) uuid.UUID {
	if id == "" {
		return uuid.Nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}

func parseUUIDPtr(id *string) *uuid.UUID {
	if id == nil {
		return nil
	}
	parsed, err := uuid.Parse(*id)
	if err != nil {
		return nil
	}
	val := parsed
	return &val
}

func uuidToString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func uuidPtrToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	value := uuidToString(*id)
	if value == "" {
		return nil
	}
	return &value
}

func cloneTime(t time.Time) time.Time {
	return t
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	value := *t
	return &value
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	value := *s
	return &value
}

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

	*m = User{
		ID:           parseUUID(e.ID),
		Email:        e.Email,
		PasswordHash: e.PasswordHash,
		DisplayName:  e.DisplayName,
		AvatarURL:    cloneStringPtr(e.AvatarURL),
		CreatedAt:    cloneTime(e.CreatedAt),
		UpdatedAt:    cloneTime(e.UpdatedAt),
	}
}

func (m *User) ToEntity() *entity.User {
	if m == nil {
		return nil
	}

	return &entity.User{
		ID:           uuidToString(m.ID),
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		DisplayName:  m.DisplayName,
		AvatarURL:    cloneStringPtr(m.AvatarURL),
		CreatedAt:    cloneTime(m.CreatedAt),
		UpdatedAt:    cloneTime(m.UpdatedAt),
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

	*m = Session{
		ID:               parseUUID(e.ID),
		UserID:           parseUUID(e.UserID),
		RefreshTokenHash: e.RefreshTokenHash,
		ExpiresAt:        cloneTime(e.ExpiresAt),
		RevokedAt:        cloneTimePtr(e.RevokedAt),
		CreatedAt:        cloneTime(e.CreatedAt),
	}
}

func (m *Session) ToEntity() *entity.Session {
	if m == nil {
		return nil
	}

	return &entity.Session{
		ID:               uuidToString(m.ID),
		UserID:           uuidToString(m.UserID),
		RefreshTokenHash: m.RefreshTokenHash,
		ExpiresAt:        cloneTime(m.ExpiresAt),
		RevokedAt:        cloneTimePtr(m.RevokedAt),
		CreatedAt:        cloneTime(m.CreatedAt),
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

	*m = Workspace{
		ID:          parseUUID(e.ID),
		Name:        e.Name,
		Description: cloneStringPtr(e.Description),
		IconURL:     cloneStringPtr(e.IconURL),
		CreatedBy:   parseUUID(e.CreatedBy),
		CreatedAt:   cloneTime(e.CreatedAt),
		UpdatedAt:   cloneTime(e.UpdatedAt),
	}
}

func (m *Workspace) ToEntity() *entity.Workspace {
	if m == nil {
		return nil
	}

	return &entity.Workspace{
		ID:          uuidToString(m.ID),
		Name:        m.Name,
		Description: cloneStringPtr(m.Description),
		IconURL:     cloneStringPtr(m.IconURL),
		CreatedBy:   uuidToString(m.CreatedBy),
		CreatedAt:   cloneTime(m.CreatedAt),
		UpdatedAt:   cloneTime(m.UpdatedAt),
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

	*m = WorkspaceMember{
		WorkspaceID: parseUUID(e.WorkspaceID),
		UserID:      parseUUID(e.UserID),
		Role:        string(e.Role),
		JoinedAt:    cloneTime(e.JoinedAt),
	}
}

func (m *WorkspaceMember) ToEntity() *entity.WorkspaceMember {
	if m == nil {
		return nil
	}

	return &entity.WorkspaceMember{
		WorkspaceID: uuidToString(m.WorkspaceID),
		UserID:      uuidToString(m.UserID),
		Role:        entity.WorkspaceRole(m.Role),
		JoinedAt:    cloneTime(m.JoinedAt),
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

	*m = Channel{
		ID:          parseUUID(e.ID),
		WorkspaceID: parseUUID(e.WorkspaceID),
		Name:        e.Name,
		Description: cloneStringPtr(e.Description),
		IsPrivate:   e.IsPrivate,
		CreatedBy:   parseUUID(e.CreatedBy),
		CreatedAt:   cloneTime(e.CreatedAt),
		UpdatedAt:   cloneTime(e.UpdatedAt),
	}
}

func (m *Channel) ToEntity() *entity.Channel {
	if m == nil {
		return nil
	}

	return &entity.Channel{
		ID:          uuidToString(m.ID),
		WorkspaceID: uuidToString(m.WorkspaceID),
		Name:        m.Name,
		Description: cloneStringPtr(m.Description),
		IsPrivate:   m.IsPrivate,
		CreatedBy:   uuidToString(m.CreatedBy),
		CreatedAt:   cloneTime(m.CreatedAt),
		UpdatedAt:   cloneTime(m.UpdatedAt),
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

	*m = ChannelMember{
		ChannelID: parseUUID(e.ChannelID),
		UserID:    parseUUID(e.UserID),
		Role:      string(e.Role),
		JoinedAt:  cloneTime(e.JoinedAt),
	}
}

func (m *ChannelMember) ToEntity() *entity.ChannelMember {
	if m == nil {
		return nil
	}

	return &entity.ChannelMember{
		ChannelID: uuidToString(m.ChannelID),
		UserID:    uuidToString(m.UserID),
		Role:      entity.ChannelRole(m.Role),
		JoinedAt:  cloneTime(m.JoinedAt),
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

	*m = Message{
		ID:        parseUUID(e.ID),
		ChannelID: parseUUID(e.ChannelID),
		UserID:    parseUUID(e.UserID),
		ParentID:  parseUUIDPtr(e.ParentID),
		Body:      e.Body,
		CreatedAt: cloneTime(e.CreatedAt),
		EditedAt:  cloneTimePtr(e.EditedAt),
		DeletedAt: cloneTimePtr(e.DeletedAt),
		DeletedBy: parseUUIDPtr(e.DeletedBy),
	}
}

func (m *Message) ToEntity() *entity.Message {
	if m == nil {
		return nil
	}

	return &entity.Message{
		ID:        uuidToString(m.ID),
		ChannelID: uuidToString(m.ChannelID),
		UserID:    uuidToString(m.UserID),
		ParentID:  uuidPtrToStringPtr(m.ParentID),
		Body:      m.Body,
		CreatedAt: cloneTime(m.CreatedAt),
		EditedAt:  cloneTimePtr(m.EditedAt),
		DeletedAt: cloneTimePtr(m.DeletedAt),
		DeletedBy: uuidPtrToStringPtr(m.DeletedBy),
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

	*m = MessageReaction{
		MessageID: parseUUID(e.MessageID),
		UserID:    parseUUID(e.UserID),
		Emoji:     e.Emoji,
		CreatedAt: cloneTime(e.CreatedAt),
	}
}

func (m *MessageReaction) ToEntity() *entity.MessageReaction {
	if m == nil {
		return nil
	}

	return &entity.MessageReaction{
		MessageID: uuidToString(m.MessageID),
		UserID:    uuidToString(m.UserID),
		Emoji:     m.Emoji,
		CreatedAt: cloneTime(m.CreatedAt),
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

	*m = ChannelReadState{
		ChannelID:  parseUUID(e.ChannelID),
		UserID:     parseUUID(e.UserID),
		LastReadAt: cloneTime(e.LastReadAt),
	}
}

func (m *ChannelReadState) ToEntity() *entity.ChannelReadState {
	if m == nil {
		return nil
	}

	return &entity.ChannelReadState{
		ChannelID:  uuidToString(m.ChannelID),
		UserID:     uuidToString(m.UserID),
		LastReadAt: cloneTime(m.LastReadAt),
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

	*m = Attachment{
		ID:         parseUUID(e.ID),
		MessageID:  parseUUIDPtr(e.MessageID),
		UploaderID: parseUUID(e.UploaderID),
		ChannelID:  parseUUID(e.ChannelID),
		FileName:   e.FileName,
		MimeType:   e.MimeType,
		SizeBytes:  e.SizeBytes,
		StorageKey: e.StorageKey,
		Status:     string(e.Status),
		UploadedAt: cloneTimePtr(e.UploadedAt),
		ExpiresAt:  cloneTimePtr(e.ExpiresAt),
		CreatedAt:  cloneTime(e.CreatedAt),
	}
}

func (m *Attachment) ToEntity() *entity.Attachment {
	if m == nil {
		return nil
	}

	return &entity.Attachment{
		ID:         uuidToString(m.ID),
		MessageID:  uuidPtrToStringPtr(m.MessageID),
		UploaderID: uuidToString(m.UploaderID),
		ChannelID:  uuidToString(m.ChannelID),
		FileName:   m.FileName,
		MimeType:   m.MimeType,
		SizeBytes:  m.SizeBytes,
		StorageKey: m.StorageKey,
		Status:     entity.AttachmentStatus(m.Status),
		UploadedAt: cloneTimePtr(m.UploadedAt),
		ExpiresAt:  cloneTimePtr(m.ExpiresAt),
		CreatedAt:  cloneTime(m.CreatedAt),
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

	*m = UserGroup{
		ID:          parseUUID(e.ID),
		WorkspaceID: parseUUID(e.WorkspaceID),
		Name:        e.Name,
		Description: cloneStringPtr(e.Description),
		CreatedBy:   parseUUID(e.CreatedBy),
		CreatedAt:   cloneTime(e.CreatedAt),
		UpdatedAt:   cloneTime(e.UpdatedAt),
	}
}

func (m *UserGroup) ToEntity() *entity.UserGroup {
	if m == nil {
		return nil
	}

	return &entity.UserGroup{
		ID:          uuidToString(m.ID),
		WorkspaceID: uuidToString(m.WorkspaceID),
		Name:        m.Name,
		Description: cloneStringPtr(m.Description),
		CreatedBy:   uuidToString(m.CreatedBy),
		CreatedAt:   cloneTime(m.CreatedAt),
		UpdatedAt:   cloneTime(m.UpdatedAt),
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

	*m = UserGroupMember{
		GroupID:  parseUUID(e.GroupID),
		UserID:   parseUUID(e.UserID),
		JoinedAt: cloneTime(e.JoinedAt),
	}
}

func (m *UserGroupMember) ToEntity() *entity.UserGroupMember {
	if m == nil {
		return nil
	}

	return &entity.UserGroupMember{
		GroupID:  uuidToString(m.GroupID),
		UserID:   uuidToString(m.UserID),
		JoinedAt: cloneTime(m.JoinedAt),
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

	*m = MessageUserMention{
		MessageID: parseUUID(e.MessageID),
		UserID:    parseUUID(e.UserID),
		CreatedAt: cloneTime(e.CreatedAt),
	}
}

func (m *MessageUserMention) ToEntity() *entity.MessageUserMention {
	if m == nil {
		return nil
	}

	return &entity.MessageUserMention{
		MessageID: uuidToString(m.MessageID),
		UserID:    uuidToString(m.UserID),
		CreatedAt: cloneTime(m.CreatedAt),
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

	*m = MessageGroupMention{
		MessageID: parseUUID(e.MessageID),
		GroupID:   parseUUID(e.GroupID),
		CreatedAt: cloneTime(e.CreatedAt),
	}
}

func (m *MessageGroupMention) ToEntity() *entity.MessageGroupMention {
	if m == nil {
		return nil
	}

	return &entity.MessageGroupMention{
		MessageID: uuidToString(m.MessageID),
		GroupID:   uuidToString(m.GroupID),
		CreatedAt: cloneTime(m.CreatedAt),
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

	*m = MessageLink{
		ID:          parseUUID(e.ID),
		MessageID:   parseUUID(e.MessageID),
		URL:         e.URL,
		Title:       cloneStringPtr(e.Title),
		Description: cloneStringPtr(e.Description),
		ImageURL:    cloneStringPtr(e.ImageURL),
		SiteName:    cloneStringPtr(e.SiteName),
		CardType:    cloneStringPtr(e.CardType),
		CreatedAt:   cloneTime(e.CreatedAt),
	}
}

func (m *MessageLink) ToEntity() *entity.MessageLink {
	if m == nil {
		return nil
	}

	return &entity.MessageLink{
		ID:          uuidToString(m.ID),
		MessageID:   uuidToString(m.MessageID),
		URL:         m.URL,
		Title:       cloneStringPtr(m.Title),
		Description: cloneStringPtr(m.Description),
		ImageURL:    cloneStringPtr(m.ImageURL),
		SiteName:    cloneStringPtr(m.SiteName),
		CardType:    cloneStringPtr(m.CardType),
		CreatedAt:   cloneTime(m.CreatedAt),
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
		participantIDs = append(participantIDs, parseUUID(id))
	}

	*m = ThreadMetadata{
		MessageID:          parseUUID(e.MessageID),
		ReplyCount:         e.ReplyCount,
		LastReplyAt:        cloneTimePtr(e.LastReplyAt),
		LastReplyUserID:    parseUUIDPtr(e.LastReplyUserID),
		ParticipantUserIDs: participantIDs,
		CreatedAt:          cloneTime(e.CreatedAt),
		UpdatedAt:          cloneTime(e.UpdatedAt),
	}
}

func (m *ThreadMetadata) ToEntity() *entity.ThreadMetadata {
	if m == nil {
		return nil
	}

	participantIDs := make([]string, 0, len(m.ParticipantUserIDs))
	for _, id := range m.ParticipantUserIDs {
		participantIDs = append(participantIDs, uuidToString(id))
	}

	return &entity.ThreadMetadata{
		MessageID:          uuidToString(m.MessageID),
		ReplyCount:         m.ReplyCount,
		LastReplyAt:        cloneTimePtr(m.LastReplyAt),
		LastReplyUserID:    uuidPtrToStringPtr(m.LastReplyUserID),
		ParticipantUserIDs: participantIDs,
		CreatedAt:          cloneTime(m.CreatedAt),
		UpdatedAt:          cloneTime(m.UpdatedAt),
	}
}
