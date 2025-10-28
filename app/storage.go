package app

import (
	"context"
	"nomnomhub/internal/model"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Storage struct {
	db *bun.DB
}

func NewStorage(db *bun.DB) *Storage {
	return &Storage{db: db}
}

func WithTx(ctx context.Context, db *bun.DB, fn func(ctx context.Context, tx bun.Tx) error) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx, tx)
	})
}

func (s *Storage) UpsertUserByLineID(ctx context.Context, d model.User) error {
	_, err := s.db.NewInsert().
		Model(d).
		On("CONFLICT (line_id) DO UPDATE").
		Set("display_name = EXCLUDED.display_name").
		Returning("*").
		Exec(ctx)
	return err
}

func (s *Storage) CreatePlace(ctx context.Context, d model.Place) error {
	_, err := s.db.NewInsert().Model(d).Exec(ctx)
	return err
}

func (s *Storage) AddTagsToPlace(ctx context.Context, placeID uuid.UUID, tagNames []string) error {
	if len(tagNames) == 0 {
		return nil
	}

	norm := make([]string, 0, len(tagNames))
	for _, t := range tagNames {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		norm = append(norm, t)
	}
	if len(norm) == 0 {
		return nil
	}

	return WithTx(ctx, s.db, func(ctx context.Context, tx bun.Tx) error {
		var tags []*model.Tag
		for _, name := range norm {
			tag := &model.Tag{Name: name}
			_, err := tx.NewInsert().
				Model(tag).
				On("CONFLICT (name) DO UPDATE SET name = EXCLUDED.name").
				Returning("*").
				Exec(ctx)
			if err != nil {
				return err
			}
			tags = append(tags, tag)
		}

		for _, t := range tags {
			pt := &model.PlaceTag{PlaceID: placeID, TagID: t.ID}
			_, err := tx.NewInsert().
				Model(pt).
				On("CONFLICT (place_id, tag_id) DO NOTHING").
				Exec(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
