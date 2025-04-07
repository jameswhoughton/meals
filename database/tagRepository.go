package database

import (
	"database/sql"
	"fmt"

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

func (ir *TagRepository) GetById(id int) (meals.Tag, error) {
	var tag meals.Tag

	err := ir.db.QueryRow("SELECT id, user_id, name FROM tags WHERE id = ?", id).Scan(
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

func (ir *TagRepository) Find(search string, userId int) ([]meals.Tag, error) {
	tags := []meals.Tag{}

	rows, err := ir.db.Query(
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

func (ir *TagRepository) Update(tag meals.Tag) error {
	ir.db.Exec("UPDATE tags SET name = ? WHERE id = ? AND user_id = ?", tag.Name, tag.Id, tag.UserId)

	return nil
}
