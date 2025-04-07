package meals_test

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

type tagRepositoryContract struct {
	// As the tagRepository is only responsible for fetching/editing existing tags,
	// any tags required for the test should be added directly to the store
	repo func(tags []meals.Tag) (meals.TagRepository, func())
}

func (i tagRepositoryContract) Test(t *testing.T) {
	t.Run("Can filter a list of tags by name", func(t *testing.T) {
		tags := []meals.Tag{
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

		repo, closeDown := i.repo(tags)
		defer closeDown()

		searchString := "ag "
		tags, err := repo.Find(searchString, 1)

		if err != nil {
			t.Errorf("List tags: Unexpected error: %v", err)
		}

		if len(tags) != 2 {
			t.Errorf("Expected 2 results, got %d", len(tags))
		}

	})

	t.Run("Can update the name of an tag", func(t *testing.T) {
		tags := []meals.Tag{
			{
				Id:   1,
				Name: "quck",
			},
		}

		repo, closeDown := i.repo(tags)
		defer closeDown()

		newName := "Quick"

		err := repo.Update(meals.Tag{Id: 1, Name: newName})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		fetchedtag, err := repo.GetById(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fetchedtag.Name != newName {
			t.Errorf("Expected tag name to be updated to %s, found %s", newName, fetchedtag.Name)
		}
	})

	t.Run("Can Fetch tag by ID", func(t *testing.T) {
		tags := []meals.Tag{
			{
				Id:   1,
				Name: "Quick",
			},
			{
				Id:   2,
				Name: "Easy",
			},
		}

		repo, closeDown := i.repo(tags)
		defer closeDown()

		tag, err := repo.GetById(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if tag.Name != "Quick" {
			t.Errorf("Expected tag with name 'quick' got '%s'", tag.Name)
		}

		_, err = repo.GetById(10)

		if err == nil {
			t.Errorf("Expected error, got none")
		}

		if !errors.As(err, &meals.ErrorTagNotFound{Id: 10}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorTagNotFound{}, err, err)
		}

	})
}

func TestDatabasetagRepository(t *testing.T) {
	init := func(tags []meals.Tag) (meals.TagRepository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		for _, tag := range tags {
			_, err := conn.Exec("INSERT INTO tags (id, user_id, name) VALUES (?, ?, ?)", tag.Id, tag.UserId, tag.Name)

			if err != nil {
				log.Fatalf("Error inserting test data: %v", err)
			}
		}

		closeDown := func() {
			os.Remove("meals.db")
		}
		return database.NewTagRepository(conn), closeDown
	}

	contract := tagRepositoryContract{
		init,
	}

	contract.Test(t)

}

func TestMemorytagRepository(t *testing.T) {
	init := func(tags []meals.Tag) (meals.TagRepository, func()) {
		return &memory.TagRepository{
			Store: tags,
		}, func() {}
	}

	contract := tagRepositoryContract{
		init,
	}

	contract.Test(t)

}
