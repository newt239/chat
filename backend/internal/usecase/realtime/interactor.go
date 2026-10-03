package realtime

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// TicketTTL は発行したチケットで WebSocket に接続できる期限
const TicketTTL = 30 * time.Second

// Ticket は WebSocket の接続に使う 1 回限りのチケットが表す接続者です
type Ticket struct {
	UserID      string `json:"userId"`
	SessionID   string `json:"sessionId"`
	WorkspaceID string `json:"workspaceId"`
}

// TicketStore はチケットのハッシュをキーに期限付きで保存し、1 回だけ取り出せるようにします
type TicketStore interface {
	Save(ctx context.Context, hash string, ticket Ticket, ttl time.Duration) error
	// Consume は取り出したチケットを消します。期限切れや使用済みなら nil を返します
	Consume(ctx context.Context, hash string) (*Ticket, error)
}

type Interactor struct {
	tickets       TicketStore
	workspaceRepo domainrepository.WorkspaceRepository
	sessionRepo   domainrepository.SessionRepository
}

func NewInteractor(tickets TicketStore, workspaceRepo domainrepository.WorkspaceRepository, sessionRepo domainrepository.SessionRepository) *Interactor {
	return &Interactor{tickets: tickets, workspaceRepo: workspaceRepo, sessionRepo: sessionRepo}
}

// IssueTicket はワークスペースのメンバーにだけチケットを発行します
func (i *Interactor) IssueTicket(ctx context.Context, ticket Ticket) (string, error) {
	if err := i.verify(ctx, ticket); err != nil {
		return "", err
	}
	token, hash, err := entity.NewSecretToken()
	if err != nil {
		return "", err
	}
	if err := i.tickets.Save(ctx, hash, ticket, TicketTTL); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeTicket はチケットを使用済みにし、発行後に退出やログアウトをしていないか確かめ直します
func (i *Interactor) ConsumeTicket(ctx context.Context, token string) (*Ticket, error) {
	ticket, err := i.tickets.Consume(ctx, entity.HashSecretToken(token))
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, domerr.ErrInvalidToken
	}
	if err := i.verify(ctx, *ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

func (i *Interactor) verify(ctx context.Context, ticket Ticket) error {
	session, err := i.sessionRepo.FindByID(ctx, ticket.SessionID)
	if err != nil {
		return err
	}
	if session == nil || session.RevokedAt != nil || session.UserID != ticket.UserID {
		return domerr.ErrInvalidToken
	}
	member, err := i.workspaceRepo.FindMember(ctx, ticket.WorkspaceID, ticket.UserID)
	if err != nil {
		return err
	}
	if member == nil {
		return domerr.ErrForbidden
	}
	return nil
}
