package db

import (
	"time"

	"github.com/google/uuid"
)

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

type WorkspaceMember struct {
	WorkspaceID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Role        string    `gorm:"type:text;not null"`
	JoinedAt    time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

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

type ChannelMember struct {
	ChannelID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	JoinedAt  time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (ChannelMember) TableName() string {
	return "channel_members"
}

type Message struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ChannelID uuid.UUID  `gorm:"type:uuid;not null;index:idx_channel_created"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	ParentID  *uuid.UUID `gorm:"type:uuid;index:idx_parent_created"`
	Body      string     `gorm:"type:text;not null"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;default:now();index:idx_channel_created,idx_parent_created"`
	EditedAt  *time.Time `gorm:"type:timestamptz"`
	DeletedAt *time.Time `gorm:"type:timestamptz"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageReaction struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Emoji     string    `gorm:"type:text;primaryKey"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageReaction) TableName() string {
	return "message_reactions"
}

type ChannelReadState struct {
	ChannelID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_user_last_read"`
	LastReadAt time.Time `gorm:"type:timestamptz;not null;default:now();index:idx_user_last_read"`
}

func (ChannelReadState) TableName() string {
	return "channel_read_states"
}

type Attachment struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MessageID  uuid.UUID `gorm:"type:uuid;not null;index"`
	FileName   string    `gorm:"type:text;not null"`
	MimeType   string    `gorm:"type:text;not null"`
	SizeBytes  int64     `gorm:"type:bigint;not null"`
	StorageKey string    `gorm:"type:text;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Attachment) TableName() string {
	return "attachments"
}

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

type UserGroupMember struct {
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	JoinedAt  time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (UserGroupMember) TableName() string {
	return "user_group_members"
}

type MessageUserMention struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageUserMention) TableName() string {
	return "message_user_mentions"
}

type MessageGroupMention struct {
	MessageID uuid.UUID `gorm:"type:uuid;primaryKey"`
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (MessageGroupMention) TableName() string {
	return "message_group_mentions"
}

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
