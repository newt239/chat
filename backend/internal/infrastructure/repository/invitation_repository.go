package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/invitation"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type invitationRepository struct {
	client *ent.Client
}

func NewInvitationRepository(client *ent.Client) domainrepository.InvitationRepository {
	return &invitationRepository{client: client}
}

func (r *invitationRepository) query(ctx context.Context) *ent.InvitationQuery {
	return transaction.ResolveClient(ctx, r.client).Invitation.Query()
}

func (r *invitationRepository) Create(ctx context.Context, inv *entity.Invitation) error {
	inviterID, err := parseUUID(inv.InvitedBy, "user ID")
	if err != nil {
		return err
	}
	created, err := transaction.ResolveClient(ctx, r.client).Invitation.Create().
		SetWorkspaceID(inv.WorkspaceID).
		SetInvitedByID(inviterID).
		SetEmail(inv.Email).
		SetRole(string(inv.Role)).
		SetTokenHash(inv.TokenHash).
		SetExpiresAt(inv.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}
	inv.ID = created.ID.String()
	return nil
}

func (r *invitationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Invitation, error) {
	found, err := orNil(r.query(ctx).Where(invitation.TokenHash(tokenHash)).Only(ctx))
	if found == nil {
		return nil, err
	}
	return invitationToEntity(found), nil
}

func (r *invitationRepository) FindPendingByEmail(ctx context.Context, email string, now time.Time) ([]*entity.Invitation, error) {
	return r.findPending(ctx, now, invitation.Email(email))
}

func (r *invitationRepository) FindPendingByWorkspaceID(ctx context.Context, workspaceID string, now time.Time) ([]*entity.Invitation, error) {
	return r.findPending(ctx, now, invitation.WorkspaceID(workspaceID))
}

func (r *invitationRepository) findPending(ctx context.Context, now time.Time, where predicate.Invitation) ([]*entity.Invitation, error) {
	found, err := r.query(ctx).
		Where(where, invitation.AcceptedAtIsNil(), invitation.ExpiresAtGT(now)).
		Order(ent.Desc(invitation.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(found, invitationToEntity), nil
}

func (r *invitationRepository) MarkAccepted(ctx context.Context, id string, acceptedAt time.Time) error {
	invitationID, err := parseUUID(id, "invitation ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Invitation.UpdateOneID(invitationID).SetAcceptedAt(acceptedAt).Exec(ctx)
}

func (r *invitationRepository) Delete(ctx context.Context, workspaceID, id string) error {
	invitationID, err := parseUUID(id, "invitation ID")
	if err != nil {
		return err
	}
	deleted, err := transaction.ResolveClient(ctx, r.client).Invitation.Delete().
		Where(invitation.ID(invitationID), invitation.WorkspaceID(workspaceID)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domerr.ErrNotFound
	}
	return nil
}

func invitationToEntity(i *ent.Invitation) *entity.Invitation {
	return &entity.Invitation{
		ID:          i.ID.String(),
		Email:       i.Email,
		Role:        entity.WorkspaceRole(i.Role),
		TokenHash:   i.TokenHash,
		ExpiresAt:   i.ExpiresAt,
		AcceptedAt:  i.AcceptedAt,
		WorkspaceID: i.WorkspaceID,
		InvitedBy:   i.InvitedByID.String(),
	}
}
