package contracts

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jameswhoughton/meals"
)

type toPtrValue interface {
	string | bool | time.Time | []int
}

func toPtr[T toPtrValue](v T) *T { return &v }

type MealRepository struct {
	Repo func() (meals.MealRepository, func(userId int), func())
}

func (i MealRepository) Test(t *testing.T) {
	t.Run("Can create get update and delete a meal", func(t *testing.T) {
		repo, seedUser, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		seedUser(1)

		newMeal := meals.Meal{
			Name:   "Bolognese",
			UserId: 1,
			Tags: []meals.Tag{
				{
					Name: "Easy",
				},
				{
					Name: "Family",
				},
			},
			Ingredients: []meals.Ingredient{
				{
					Name:     "Beef mince",
					Quantity: 500,
					Unit:     "gram",
				},
				{
					Name:     "Tinned tomatoes",
					Quantity: 2,
					Unit:     "can",
				},
				{
					Name:     "Garlic",
					Quantity: 3,
					Unit:     "clove",
				},
			},
		}

		createdMeal, err := repo.Create(ctx, newMeal)

		if err != nil {
			t.Errorf("Creating ingredient: unexpected error: %v", err)
		}

		if createdMeal.Id == 0 {
			t.Error("Expected non zero ID")
		}

		if newMeal.Name != createdMeal.Name {
			t.Errorf("Expected Name %s, got %s", newMeal.Name, createdMeal.Name)
		}

		if len(newMeal.Ingredients) != len(createdMeal.Ingredients) {
			t.Errorf("Expected created meal to have %d ingredients, found: %d", len(newMeal.Ingredients), len(createdMeal.Ingredients))
		}

		if len(newMeal.Tags) != len(createdMeal.Tags) {
			t.Errorf("Expected created meal to have %d tags, found: %d", len(newMeal.Tags), len(createdMeal.Tags))
		}

		for _, ingredient := range newMeal.Ingredients {
			if ingredient.Id == 0 {
				t.Errorf("Ingredient %s has an ID of 0", ingredient.Name)
			}
		}

		for _, tag := range newMeal.Tags {
			if tag.Id == 0 {
				t.Errorf("Tag %s has an ID of 0", tag.Name)
			}
		}

		updatedMeal, err := repo.Get(ctx, createdMeal.Id)

		if err != nil {
			t.Errorf("Fetching meal: unexpected error: %v", err)
		}

		updatedMeal.Tags = append(updatedMeal.Tags, meals.Tag{
			Name: "Quick",
		})

		updatedMeal.Ingredients = append(updatedMeal.Ingredients, meals.Ingredient{
			Name:     "Onion",
			Quantity: 2,
		})

		err = repo.Update(ctx, updatedMeal)

		if err != nil {
			t.Errorf("Updating meal: unexpected error: %v", err)
		}

		fetchedMeal, err := repo.Get(ctx, createdMeal.Id)

		if err != nil {
			t.Errorf("Fetching meal: unexpected error: %v", err)
		}

		if createdMeal.Id != fetchedMeal.Id {
			t.Errorf("Expected ID %d, got %d", createdMeal.Id, fetchedMeal.Id)
		}

		if fetchedMeal.Name != updatedMeal.Name {
			t.Errorf("Expected Name %s, got %s", fetchedMeal.Name, updatedMeal.Name)
		}

		if len(updatedMeal.Ingredients) != len(fetchedMeal.Ingredients) {
			t.Errorf("Expected updated meal to have %d ingredients, found: %d", len(updatedMeal.Ingredients), len(fetchedMeal.Ingredients))
		}

		if len(updatedMeal.Tags) != len(fetchedMeal.Tags) {
			t.Errorf("Expected updated meal to have %d tags, found: %d", len(updatedMeal.Tags), len(fetchedMeal.Tags))
		}

		for _, ingredient := range fetchedMeal.Ingredients {
			if ingredient.Id == 0 {
				t.Errorf("Ingredient %s has an ID of 0", ingredient.Name)
			}
		}

		for _, tag := range fetchedMeal.Tags {
			if tag.Id == 0 {
				t.Errorf("Tag %s has an ID of 0", tag.Name)
			}
		}

		err = repo.Destroy(ctx, fetchedMeal.Id)

		if err != nil {
			t.Errorf("Destroying meal: unexpected error: %v", err)
		}

		_, err = repo.Get(ctx, fetchedMeal.Id)

		if err == nil {
			t.Error("Expected error, got none")
		}

		if !errors.Is(err, meals.ErrMealNotFound) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrMealNotFound, err, err)
		}
	})

	t.Run("Can filter a list of meals", func(t *testing.T) {
		repo, seedUser, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		seedUser(1)
		seedUser(2)

		mealA, _ := repo.Create(ctx, meals.Meal{
			UserId: 1,
			Name:   "Meal A",
			Tags: []meals.Tag{
				{
					Id:   1,
					Name: "Quick",
				},
				{
					Id:   2,
					Name: "Family",
				},
				{
					Id:   3,
					Name: "Easy",
				},
			},
			Ingredients: []meals.Ingredient{
				{
					Id: 1,
				},
			},
		})

		repo.Create(ctx, meals.Meal{
			UserId: 1,
			Name:   "Meal B",
			Tags: []meals.Tag{
				{
					Id:   2,
					Name: "Family",
				},
				{
					Id:   3,
					Name: "Easy",
				},
			},
			Ingredients: []meals.Ingredient{
				{
					Id: 2,
				},
				{
					Id: 1,
				},
			},
		})

		repo.Create(ctx, meals.Meal{
			UserId: 2,
			Name:   "Meal C",
			Tags: []meals.Tag{
				{
					Id:   1,
					Name: "Quick",
				},
				{
					Id:   2,
					Name: "Family",
				},
				{
					Id:   3,
					Name: "Easy",
				},
			},
			Ingredients: []meals.Ingredient{
				{
					Id: 1,
				},
			},
		})

		type testCase struct {
			label         string
			filters       meals.MealFilter
			expectedMeals []int
		}

		testCases := []testCase{
			{
				label: "All tags",
				filters: meals.MealFilter{
					UserId: 1,
					Tags:   []string{"Quick", "Family", "Easy"},
				},
				expectedMeals: []int{mealA.Id},
			},
			{
				label: "By name",
				filters: meals.MealFilter{
					UserId: 1,
					Name:   toPtr("eAl a"),
				},
				expectedMeals: []int{mealA.Id},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.label, func(t *testing.T) {
				meals, err := repo.Find(ctx, testCase.filters)

				if err != nil {
					t.Errorf("Unexpected error %v", err)
				}

				if len(meals) != len(testCase.expectedMeals) {
					t.Errorf("Expected %d results, got %d", len(testCase.expectedMeals), len(meals))
				}

				for _, meal := range meals {
					if !slices.Contains(testCase.expectedMeals, meal.Id) {
						t.Errorf("Unexpected meal ID %d in results", meal.Id)
					}
				}
			})
		}
	})
}

type MealMetaDataRepository struct {
	Repo func([]meals.Meal) (meals.MealMetaDataRepository, func())
}

func (i MealMetaDataRepository) Test(t *testing.T) {
	t.Run("Can filter a list of distinct ingredient names", func(t *testing.T) {
		testData := []meals.Meal{
			{
				UserId: 1,
				Ingredients: []meals.Ingredient{
					{
						Id:   2,
						Name: "Onion",
					},
					{
						Id:   1,
						Name: "Spring Onion",
					},
				},
			},
			{
				UserId: 2,
				Ingredients: []meals.Ingredient{
					{
						Id:   9,
						Name: "Red Onion",
					},
					{
						Id:   5,
						Name: "Onion",
					},
					{
						Id:   7,
						Name: "Cheese",
					},
				},
			},
		}

		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ctx := context.Background()

		searchString := "Onio"
		ingredients, err := repo.FindIngredientNames(ctx, searchString)

		if err != nil {
			t.Errorf("List ingredients: Unexpected error: %v", err)
		}

		expectedResults := []string{"Onion", "Spring Onion", "Red Onion"}

		if len(ingredients) != len(expectedResults) {
			t.Errorf("Expected %d results, got %d", len(expectedResults), len(ingredients))
		}

		for _, name := range expectedResults {
			if !slices.Contains(ingredients, name) {
				t.Errorf("%s missing in the result set", name)
			}
		}

	})
	t.Run("Can filter a list of distinct unit names", func(t *testing.T) {
		testData := []meals.Meal{
			{
				UserId: 1,
				Ingredients: []meals.Ingredient{
					{
						Id:   2,
						Name: "tomatoes",
						Unit: "cans",
					},
					{
						Id:   1,
						Name: "Spring Onion",
						Unit: "Bunches",
					},
					{
						Id:   100,
						Name: "Diced Chicken",
						Unit: "KG",
					},
				},
			},
			{
				UserId: 2,
				Ingredients: []meals.Ingredient{
					{
						Id:   7,
						Name: "Cheese",
						Unit: "Grams",
					},
				},
			},
		}
		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ctx := context.Background()

		searchString := "S"
		units, err := repo.FindUnitNames(ctx, searchString)

		if err != nil {
			t.Errorf("List units: Unexpected error: %v", err)
		}

		expectedResults := []string{"cans", "Bunches", "Grams"}

		if len(units) != len(expectedResults) {
			t.Errorf("Expected %d results, got %d", len(expectedResults), len(units))
		}

		for _, name := range expectedResults {
			if !slices.Contains(units, name) {
				t.Errorf("%s missing in the result set", name)
			}
		}

	})
	t.Run("Can filter a list of distinct tag names", func(t *testing.T) {
		testData := []meals.Meal{
			{
				UserId: 1,
				Tags: []meals.Tag{
					{
						Id:   2,
						Name: "AAA",
					},
					{
						Id:   1,
						Name: "baa",
					},
				},
			},
			{
				UserId: 2,
				Tags: []meals.Tag{
					{
						Id:   9,
						Name: "caa",
					},
					{
						Id:   5,
						Name: "Quick",
					},
					{
						Id:   7,
						Name: "Family",
					},
				},
			},
		}
		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ctx := context.Background()

		searchString := "aa"
		tags, err := repo.FindTagNames(ctx, searchString)

		if err != nil {
			t.Errorf("List Tags: Unexpected error: %v", err)
		}

		expectedResults := []string{"AAA", "baa", "caa"}

		if len(tags) != len(expectedResults) {
			t.Errorf("Expected %d results, got %d", len(expectedResults), len(tags))
		}

		for _, name := range expectedResults {
			if !slices.Contains(tags, name) {
				t.Errorf("%s missing in the result set", name)
			}
		}

	})
	t.Run("Can list distinct tag names for a user", func(t *testing.T) {
		testData := []meals.Meal{
			{
				UserId: 1,
				Tags: []meals.Tag{
					{
						Id:   2,
						Name: "Quick",
					},
					{
						Id:   1,
						Name: "Family",
					},
				},
			},
			{
				UserId: 1,
				Tags: []meals.Tag{
					{
						Id:   5,
						Name: "Quick",
					},
					{
						Id:   7,
						Name: "Family",
					},
				},
			},
		}
		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ctx := context.Background()

		tags, err := repo.TagNamesForUser(ctx, 1)

		if err != nil {
			t.Errorf("List Tags for user: Unexpected error: %v", err)
		}

		expectedResults := []string{"Quick", "Family"}

		if len(tags) != len(expectedResults) {
			t.Errorf("Expected %d results, got %d", len(expectedResults), len(tags))
		}

		for _, name := range expectedResults {
			if !slices.Contains(tags, name) {
				t.Errorf("%s missing in the result set", name)
			}
		}

	})

	t.Run("GetTotalIngredients counts meals multiple times if they are included more than once", func(t *testing.T) {
		testData := []meals.Meal{
			{
				Id:     1,
				UserId: 1,
				Ingredients: []meals.Ingredient{
					{
						Name:     "Salmon",
						Quantity: 400,
						Unit:     "g",
					},
				},
			},
		}

		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ingredients, err := repo.GetTotalIngredients(context.Background(), []int{1, 1})

		if err != nil {
			t.Errorf("Unexpected error fetching ingredients: %v", err)
		}

		expectedIngredients := []meals.IngredientTotal{
			{
				Name:     "Salmon",
				Quantity: 800,
				Unit:     "g",
			},
		}

		if len(ingredients) != len(expectedIngredients) {
			t.Errorf("The number of ingredients (%d) doesn't match the expected: %d", len(ingredients), len(expectedIngredients))
		}

		sort := func(a, b meals.IngredientTotal) int {
			return strings.Compare(a.Name, b.Name)
		}

		slices.SortFunc(ingredients, sort)
		slices.SortFunc(expectedIngredients, sort)

		for i, ingredient := range ingredients {
			if ingredient != expectedIngredients[i] {
				t.Errorf("Ingredients do not match, expected: %+v, got %+v", expectedIngredients[i], ingredient)
			}
		}
	})

	t.Run("Can fetch list of ingredient totals for list of meals", func(t *testing.T) {
		testData := []meals.Meal{
			{
				Id:     1,
				UserId: 1,
				Name:   "Salmon stir-fry",
				Ingredients: []meals.Ingredient{
					{
						Name:     "Salmon",
						Quantity: 400,
						Unit:     "g",
					},
					{
						Name:     "Bell Pepper",
						Quantity: 2,
					},
					{
						Name:     "Onion",
						Quantity: 1,
					},
				},
			},
			{
				Id:     3,
				UserId: 1,
				Name:   "Salmon curry",
				Ingredients: []meals.Ingredient{
					{
						Name:     "Salmon",
						Quantity: 400,
						Unit:     "g",
					},
					{
						Name:     "Bell Pepper",
						Quantity: 1,
					},
					{
						Name:     "Onion",
						Quantity: 1,
					},
					{
						Name:     "Ginger",
						Quantity: 3,
						Unit:     "cm",
					},
				},
			},
		}

		repo, closeDown := i.Repo(testData)
		defer closeDown()

		ingredients, err := repo.GetTotalIngredients(context.Background(), []int{1, 3})

		if err != nil {
			t.Errorf("Unexpected error fetching ingredients: %v", err)
		}

		expectedIngredients := []meals.IngredientTotal{
			{
				Name:     "Salmon",
				Quantity: 800,
				Unit:     "g",
			},
			{
				Name:     "Bell Pepper",
				Quantity: 3,
			},
			{
				Name:     "Onion",
				Quantity: 2,
			},
			{
				Name:     "Ginger",
				Quantity: 3,
				Unit:     "cm",
			},
		}

		if len(ingredients) != len(expectedIngredients) {
			t.Errorf("The number of ingredients (%d) doesn't match the expected: %d", len(ingredients), len(expectedIngredients))
		}

		sort := func(a, b meals.IngredientTotal) int {
			return strings.Compare(a.Name, b.Name)
		}

		slices.SortFunc(ingredients, sort)
		slices.SortFunc(expectedIngredients, sort)

		for i, ingredient := range ingredients {
			if ingredient != expectedIngredients[i] {
				t.Errorf("Ingredients do not match, expected: %+v, got %+v", expectedIngredients[i], ingredient)
			}
		}
	})
}
