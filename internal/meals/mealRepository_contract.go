package meals

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

type toPtrValue interface {
	string | bool | time.Time | []int
}

func toPtr[T toPtrValue](v T) *T { return &v }

type MealRepositoryContract struct {
	Repo func() (MealRepository, func())
}

func (i MealRepositoryContract) Test(t *testing.T) {
	t.Run("Can create get update and delete a meal", func(t *testing.T) {
		repo, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		newMeal := Meal{
			Name:   "Bolognese",
			UserId: 1,
			Tags: []Tag{
				{
					Name: "Easy",
				},
				{
					Name: "Family",
				},
			},
			Ingredients: []MealIngredient{
				{
					Name:     "Beef mince",
					Quantity: 500,
					Unit:     "gram",
					IsMain:   true,
				},
				{
					Name:     "Tinned tomatoes",
					Quantity: 2,
					Unit:     "can",
					IsMain:   false,
				},
				{
					Name:     "Garlic",
					Quantity: 3,
					Unit:     "clove",
					IsMain:   false,
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

		updatedMeal.Tags = append(updatedMeal.Tags, Tag{
			Name: "Quick",
		})

		updatedMeal.Ingredients = append(updatedMeal.Ingredients, MealIngredient{
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

		if !errors.Is(err, ErrorMealNotFound{Id: fetchedMeal.Id}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorMealNotFound{}, err, err)
		}
	})

	t.Run("Can filter a list of meals", func(t *testing.T) {
		repo, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		mealA, _ := repo.Create(ctx, Meal{
			UserId: 1,
			Name:   "Meal A",
			Tags: []Tag{
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
			Ingredients: []MealIngredient{
				{
					Id:     1,
					IsMain: true,
				},
			},
		})

		mealB, _ := repo.Create(ctx, Meal{
			UserId: 1,
			Name:   "Meal B",
			Tags:   []Tag{},
			Ingredients: []MealIngredient{
				{
					Id:     4,
					IsMain: true,
				},
			},
		})

		mealC, _ := repo.Create(ctx, Meal{
			UserId: 1,
			Name:   "Meal C",
			Tags: []Tag{
				{
					Id:   2,
					Name: "Family",
				},
				{
					Id:   3,
					Name: "Easy",
				},
			},
			Ingredients: []MealIngredient{
				{
					Id:     2,
					IsMain: true,
				},
				{
					Id:     1,
					IsMain: false,
				},
			},
		})

		repo.Create(ctx, Meal{
			UserId: 2,
			Name:   "Meal D",
			Tags: []Tag{
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
			Ingredients: []MealIngredient{
				{
					Id:     1,
					IsMain: true,
				},
			},
		})

		err := repo.AssignToDate(ctx, mealB.Id, time.Date(2025, time.March, 9, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		err = repo.AssignToDate(ctx, mealC.Id, time.Date(2025, time.March, 11, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		type testCase struct {
			label         string
			filters       MealFilter
			expectedMeals []int
		}

		testCases := []testCase{
			{
				label: "All attributes true",
				filters: MealFilter{
					UserId:  1,
					HasTags: []int{1, 2, 3},
				},
				expectedMeals: []int{mealA.Id, mealC.Id},
			},
			{
				label: "By name",
				filters: MealFilter{
					UserId: 1,
					Name:   toPtr("eAl a"),
				},
				expectedMeals: []int{mealA.Id},
			},
			{
				label: "Exclude ingredient",
				filters: MealFilter{
					UserId:                1,
					ExcludeMainIngredient: []int{1},
				},
				expectedMeals: []int{mealB.Id, mealC.Id},
			},
			{
				label: "Exclude recent",
				filters: MealFilter{
					UserId: 1,
					DateRange: &DateRange{
						End: toPtr(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC)),
					},
				},
				expectedMeals: []int{mealA.Id, mealB.Id},
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

	t.Run("Must include UserId when filtering meals", func(t *testing.T) {
		repo, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		_, err := repo.Find(ctx, MealFilter{})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorMealFilterInvalid{}, err, err)
		}

	})

	t.Run("filter field DateRange validation", func(t *testing.T) {
		repo, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		futureTime := time.Now().Add(time.Hour * 12)

		// Start date should not be in the future
		_, err := repo.Find(ctx, MealFilter{
			UserId: 1,
			DateRange: &DateRange{
				Start: &futureTime,
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorMealFilterInvalid{}, err, err)
		}

		// End date should not be in the future
		_, err = repo.Find(ctx, MealFilter{
			UserId: 1,
			DateRange: &DateRange{
				End: &futureTime,
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorMealFilterInvalid{}, err, err)
		}

		// End date can't be before start date
		_, err = repo.Find(ctx, MealFilter{
			UserId: 1,
			DateRange: &DateRange{
				Start: toPtr(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC)),
				End:   toPtr(time.Date(2024, time.March, 10, 0, 0, 0, 0, time.UTC)),
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", ErrorMealFilterInvalid{}, err, err)
		}
	})

	t.Run("Can assign a meal to a given date and fetch meals from a range", func(t *testing.T) {
		repo, closeDown := i.Repo()
		defer closeDown()

		ctx := context.Background()

		chickenPie, _ := repo.Create(ctx, Meal{
			Name:   "Chicken Pie",
			UserId: 1,
		})
		pizza, _ := repo.Create(ctx, Meal{
			Name:   "Pizza",
			UserId: 2,
		})
		pestoSalmon, _ := repo.Create(ctx, Meal{
			Name:   "Pesto Salmon",
			UserId: 1,
		})

		err := repo.AssignToDate(ctx, chickenPie.Id, time.Date(2025, time.March, 5, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		err = repo.AssignToDate(ctx, pizza.Id, time.Date(2025, time.March, 5, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		err = repo.AssignToDate(ctx, pestoSalmon.Id, time.Date(2025, time.March, 15, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		dateRange := DateRange{
			Start: toPtr(time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)),
			End:   toPtr(time.Date(2025, time.March, 9, 0, 0, 0, 0, time.UTC)),
		}

		meals, err := repo.Find(ctx, MealFilter{
			UserId:    1,
			DateRange: &dateRange,
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(meals) != 1 {
			t.Errorf("Expected 1 meal, got %d", len(meals))
		}

		if meals[0].Id != chickenPie.Id {
			t.Errorf("Expected Id %d, got %d (%s)", chickenPie.Id, meals[0].Id, meals[0].Name)
		}
	})

}
