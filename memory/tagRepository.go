package memory

import (
	"strings"

	"github.com/jameswhoughton/meals/internal/meals"
)

type TagRepository struct {
	Store []meals.Tag
}

func (ir *TagRepository) GetById(id int) (meals.Tag, error) {
	for _, tag := range ir.Store {
		if tag.Id == id {
			return tag, nil
		}
	}

	return meals.Tag{}, meals.ErrorTagNotFound{Id: id}
}

func (ir *TagRepository) Find(search string, userId int) ([]meals.Tag, error) {
	var tags []meals.Tag

	search = strings.ToLower(search)

	for _, tag := range ir.Store {
		if tag.UserId != userId {
			continue
		}

		if strings.Contains(strings.ToLower(tag.Name), search) {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func (ir *TagRepository) Update(tag meals.Tag) error {
	for i, existingTag := range ir.Store {
		if tag.Id == existingTag.Id {
			ir.Store[i].Name = tag.Name
			break
		}
	}

	return nil
}

func (ir *TagRepository) FromNames(names []string, userId int) (map[string]int, error) {
	tagMap := make(map[string]int, len(names))

	for _, tag := range ir.Store {
		if tag.UserId != userId {
			continue
		}

		tagMap[tag.Name] = tag.Id
	}

	return tagMap, nil
}
