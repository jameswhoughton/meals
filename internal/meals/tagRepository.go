package meals

import (
	"fmt"
)

/*
Specification:
- Has a single owner
- Can belong to multiple meals
- Cannot exist if it isn't associated with a meal
- Name should be unique
*/
type Tag struct {
	Id     int               `json:"id"`
	UserId int               `json:"-"`
	Name   string            `json:"name"`
	Errors map[string]string `json:"-"`
}

func (m *Tag) Validate() bool {
	m.Errors = make(map[string]string)

	if m.Name == "" {
		m.Errors["Name"] = "Name cannot be blank"
	}

	return len(m.Errors) == 0
}

type ErrorTagNotFound struct {
	Id int
}

func (e ErrorTagNotFound) Error() string {
	return fmt.Sprintf("Tag with the id: %d does not exist.", e.Id)
}

type TagRepository interface {
	Find(search string, userId int) ([]Tag, error)
	GetById(id int) (Tag, error)
	Update(tag Tag) error
	FromNames(names []string, userId int) (map[string]int, error)
}
