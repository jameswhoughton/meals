package database

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{
		db: db,
	}
}

type TagRepository struct {
	db *sql.DB
}

func (tr *TagRepository) GetById(ctx context.Context, id int) (meals.Tag, error) {
	var tag meals.Tag

	err := tr.db.QueryRowContext(ctx, "SELECT id, user_id, name FROM tags WHERE id = ?", id).Scan(
		&tag.Id,
		&tag.UserId,
		&tag.Name,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return meals.Tag{}, meals.ErrorTagNotFound{Id: id}
		}

		return meals.Tag{}, fmt.Errorf("TagRepository.Get: query error: %v", err)
	}

	return tag, nil
}

func (tr *TagRepository) Find(ctx context.Context, search string, userId int) ([]meals.Tag, error) {
	tags := []meals.Tag{}

	rows, err := tr.db.QueryContext(
		ctx,
		"SELECT id, user_id, name FROM tags WHERE name LIKE ? AND user_id = ?",
		"%"+search+"%",
		userId,
	)

	if err != nil {
		return tags, fmt.Errorf("TagRepository.Find: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var tag meals.Tag

		if err = rows.Scan(&tag.Id, &tag.UserId, &tag.Name); err != nil {
			return tags, fmt.Errorf("TagRepository.Find: row parse error: %v", err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (tr *TagRepository) Update(ctx context.Context, tag meals.Tag) error {
	tr.db.ExecContext(ctx, "UPDATE tags SET name = ? WHERE id = ? AND user_id = ?", tag.Name, tag.Id, tag.UserId)

	return nil
}
func (tr *TagRepository) FromNames(ctx context.Context, names []string, userId int) (map[string]int, error) {
	tagMap := make(map[string]int, len(names))
	var values []any

	for _, name := range names {
		values = append(values, name)
	}

	values = append(values, userId)
	params := slices.Repeat([]string{"?"}, len(names))

	rows, err := tr.db.QueryContext(
		ctx,
		"SELECT id, name FROM tags WHERE name IN ("+strings.Join(params, ", ")+") AND user_id = ?",
		values...,
	)

	if err != nil {
		return tagMap, fmt.Errorf("TagRepository.FromNames: query error: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var name string

		if err = rows.Scan(&id, &name); err != nil {
			return tagMap, fmt.Errorf("TagRepository.FromNames: row parse error: %v", err)
		}

		tagMap[name] = id
	}

	return tagMap, nil
}
