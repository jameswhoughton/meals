package meals

import (
	"context"
	"errors"
	"testing"
)

type TagRepositoryContract struct {
	// As the tagRepository is only responsible for fetching/editing existing tags,
	// any tags required for the test should be added directly to the store
	Repo func(tags []Tag) (TagRepository, func())
}

func (i TagRepositoryContract) Test(t *testing.T) {
	t.Run("Can filter a list of tags by name", func(t *testing.T) {
		tags := []Tag{
			{
				Id:     1,
				UserId: 1,
				Name:   "tag 1",
			},
			{
				Id:     2,
				UserId: 1,
				Name:   "tag 2",
			},
			{
				Id:     3,
				UserId: 2,
				Name:   "Quick",
			},
		}

		repo, closeDown := i.Repo(tags)
		defer closeDown()

		ctx := context.Background()

		searchString := "ag "
		tags, err := repo.Find(ctx, searchString, 1)

		if err != nil {
			t.Errorf("List tags: Unexpected error: %v", err)
		}

		if len(tags) != 2 {
			t.Errorf("Expected 2 results, got %d", len(tags))
		}

	})

	t.Run("Can update the name of an tag", func(t *testing.T) {
		tags := []Tag{
			{
				Id:   1,
				Name: "quck",
			},
		}

		repo, closeDown := i.Repo(tags)
		defer closeDown()

		ctx := context.Background()

		newName := "Quick"

		err := repo.Update(ctx, Tag{Id: 1, Name: newName})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		fetchedtag, err := repo.GetById(ctx, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fetchedtag.Name != newName {
			t.Errorf("Expected tag name to be updated to %s, found %s", newName, fetchedtag.Name)
		}
	})

	t.Run("Can Fetch tag by ID", func(t *testing.T) {
		tags := []Tag{
			{
				Id:   1,
				Name: "Quick",
			},
			{
				Id:   2,
				Name: "Easy",
			},
		}

		repo, closeDown := i.Repo(tags)
		defer closeDown()

		ctx := context.Background()

		tag, err := repo.GetById(ctx, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if tag.Name != "Quick" {
			t.Errorf("Expected tag with name 'quick' got '%s'", tag.Name)
		}

		_, err = repo.GetById(ctx, 10)

		if err == nil {
			t.Errorf("Expected error, got none")
		}

		if !errors.As(err, &ErrorTagNotFound{Id: 10}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorTagNotFound{}, err, err)
		}

	})
}
