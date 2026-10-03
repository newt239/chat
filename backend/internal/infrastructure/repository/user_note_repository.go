package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/ent/usernote"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type userNoteRepository struct {
	client *ent.Client
}

func NewUserNoteRepository(client *ent.Client) domainrepository.UserNoteRepository {
	return &userNoteRepository{client: client}
}

func notePredicate(ownerID, targetID uuid.UUID) predicate.UserNote {
	return usernote.And(usernote.OwnerID(ownerID), usernote.TargetID(targetID))
}

func parseNoteUsers(ownerID, targetID string) (uuid.UUID, uuid.UUID, error) {
	oid, err := parseUUID(ownerID, "owner ID")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	tid, err := parseUUID(targetID, "target user ID")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return oid, tid, nil
}

func (r *userNoteRepository) Find(ctx context.Context, ownerID string, targetID string) (*entity.UserNote, error) {
	oid, tid, err := parseNoteUsers(ownerID, targetID)
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	note, err := client.UserNote.Query().Where(notePredicate(oid, tid)).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.UserNote{
		OwnerID:   ownerID,
		TargetID:  targetID,
		Nickname:  note.Nickname,
		Memo:      note.Memo,
		UpdatedAt: note.UpdatedAt,
	}, nil
}

func (r *userNoteRepository) FindNicknames(ctx context.Context, ownerID string) (map[string]string, error) {
	oid, err := parseUUID(ownerID, "owner ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	notes, err := client.UserNote.Query().
		Where(usernote.OwnerID(oid), usernote.NicknameNotNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(notes))
	for _, note := range notes {
		result[note.TargetID.String()] = *note.Nickname
	}
	return result, nil
}

func (r *userNoteRepository) Upsert(ctx context.Context, note *entity.UserNote) error {
	oid, tid, err := parseNoteUsers(note.OwnerID, note.TargetID)
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	existing, err := client.UserNote.Query().Where(notePredicate(oid, tid)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	var saved *ent.UserNote
	if existing == nil {
		saved, err = client.UserNote.Create().
			SetOwnerID(oid).
			SetTargetID(tid).
			SetNillableNickname(note.Nickname).
			SetNillableMemo(note.Memo).
			Save(ctx)
	} else {
		update := existing.Update()
		if note.Nickname == nil {
			update.ClearNickname()
		} else {
			update.SetNickname(*note.Nickname)
		}
		if note.Memo == nil {
			update.ClearMemo()
		} else {
			update.SetMemo(*note.Memo)
		}
		saved, err = update.Save(ctx)
	}
	if err != nil {
		return err
	}

	note.UpdatedAt = saved.UpdatedAt
	return nil
}

func (r *userNoteRepository) Delete(ctx context.Context, ownerID string, targetID string) error {
	oid, tid, err := parseNoteUsers(ownerID, targetID)
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.UserNote.Delete().Where(notePredicate(oid, tid)).Exec(ctx)
	return err
}
