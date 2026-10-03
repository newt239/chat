package realtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members map[string]bool
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	if !r.members[workspaceID+"/"+userID] {
		return nil, nil
	}
	return &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID}, nil
}

type stubSessionRepo struct {
	domainrepository.SessionRepository
	sessions map[string]*entity.Session
}

func (r *stubSessionRepo) FindByID(_ context.Context, id string) (*entity.Session, error) {
	return r.sessions[id], nil
}

type memoryTicketStore map[string]Ticket

func (s memoryTicketStore) Save(_ context.Context, hash string, ticket Ticket, _ time.Duration) error {
	s[hash] = ticket
	return nil
}

func (s memoryTicketStore) Consume(_ context.Context, hash string) (*Ticket, error) {
	ticket, ok := s[hash]
	if !ok {
		return nil, nil
	}
	delete(s, hash)
	return &ticket, nil
}

func newInteractor() (*Interactor, *stubWorkspaceRepo) {
	workspaces := &stubWorkspaceRepo{members: map[string]bool{"ws/alice": true}}
	sessions := &stubSessionRepo{sessions: map[string]*entity.Session{"s1": {ID: "s1", UserID: "alice"}}}
	return New(memoryTicketStore{}, workspaces, sessions), workspaces
}

func TestTicketCanBeUsedOnlyOnce(t *testing.T) {
	uc, _ := newInteractor()
	ticket := Ticket{UserID: "alice", SessionID: "s1", WorkspaceID: "ws"}

	token, err := uc.IssueTicket(context.Background(), ticket)
	if err != nil {
		t.Fatal(err)
	}
	got, err := uc.ConsumeTicket(context.Background(), token)
	if err != nil || *got != ticket {
		t.Fatalf("発行したチケットで接続できません: %v %v", got, err)
	}
	if _, err := uc.ConsumeTicket(context.Background(), token); !errors.Is(err, domerr.ErrInvalidToken) {
		t.Errorf("使用済みのチケットは拒否するはず: %v", err)
	}
}

func TestTicketRequiresMembership(t *testing.T) {
	uc, workspaces := newInteractor()
	if _, err := uc.IssueTicket(context.Background(), Ticket{UserID: "alice", SessionID: "s1", WorkspaceID: "other"}); !errors.Is(err, domerr.ErrForbidden) {
		t.Errorf("メンバーでないワークスペースのチケットは発行しないはず: %v", err)
	}

	token, err := uc.IssueTicket(context.Background(), Ticket{UserID: "alice", SessionID: "s1", WorkspaceID: "ws"})
	if err != nil {
		t.Fatal(err)
	}
	// 発行後にワークスペースから外された
	workspaces.members["ws/alice"] = false
	if _, err := uc.ConsumeTicket(context.Background(), token); !errors.Is(err, domerr.ErrForbidden) {
		t.Errorf("接続時に所属を確かめ直すはず: %v", err)
	}
}
