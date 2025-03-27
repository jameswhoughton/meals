package meals_test

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/meals"
	"github.com/jameswhoughton/meals/memory"
)

type toPtrValue interface {
	string | bool | time.Time | []int
}

func toPtr[T toPtrValue](v T) *T { return &v }

type RepositoryContract struct {
	repo func() (meals.Repository, func())
}

func (i RepositoryContract) Test(t *testing.T) {
	t.Run("Can create get update and delete a meal", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		newMeal := meals.Meal{
			Name:   "Bolognese",
			UserId: 1,
			Attributes: meals.MealAttributes{
				Easy:   true,
				Quick:  false,
				Family: true,
			},
			Ingredients: []meals.MealIngredient{
				{
					Id:       1,
					Name:     "Beef mince",
					Quantity: 500,
					Unit:     "gram",
					IsMain:   true,
				},
				{
					Id:       2,
					Name:     "Tinned tomatoes",
					Quantity: 2,
					Unit:     "can",
					IsMain:   false,
				},
				{
					// New ingredient
					Name:     "Garlic",
					Quantity: 3,
					Unit:     "clove",
					IsMain:   false,
				},
			},
		}

		createdMeal, err := repo.Create(newMeal)

		if err != nil {
			t.Errorf("Creating ingredient: unexpected error: %v", err)
		}

		if createdMeal.Id == 0 {
			t.Error("Expected non zero ID")
		}

		if newMeal.Name != createdMeal.Name {
			t.Errorf("Expected Name %s, got %s", newMeal.Name, createdMeal.Name)
		}

		for _, ingredient := range newMeal.Ingredients {
			if ingredient.Id == 0 {
				t.Errorf("Ingredient %s has an ID of 0", ingredient.Name)
			}
		}

		updatedMeal := createdMeal

		updatedMeal.Attributes.Easy = false

		updatedMeal.Ingredients = append(updatedMeal.Ingredients, meals.MealIngredient{
			Id:       4,
			Name:     "Onion",
			Quantity: 2,
		})

		err = repo.Update(updatedMeal)

		if err != nil {
			t.Errorf("Updating meal: unexpected error: %v", err)
		}

		fetchedMeal, err := repo.Get(createdMeal.Id)

		if err != nil {
			t.Errorf("Fetching meal: unexpected error: %v", err)
		}

		if createdMeal.Id != fetchedMeal.Id {
			t.Errorf("Expected ID %d, got %d", createdMeal.Id, fetchedMeal.Id)
		}

		if fetchedMeal.Name != updatedMeal.Name {
			t.Errorf("Expected Name %s, got %s", fetchedMeal.Name, updatedMeal.Name)
		}

		for _, ingredient := range fetchedMeal.Ingredients {
			if ingredient.Id == 0 {
				t.Errorf("Ingredient %s has an ID of 0", ingredient.Name)
			}
		}

		err = repo.Destroy(fetchedMeal.Id)

		if err != nil {
			t.Errorf("Destroying meal: unexpected error: %v", err)
		}

		_, err = repo.Get(fetchedMeal.Id)

		if err == nil {
			t.Error("Expected error, got none")
		}

		if !errors.Is(err, meals.ErrorMealNotFound{Id: fetchedMeal.Id}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorMealNotFound{}, err, err)
		}
	})

	t.Run("Can filter a list of meals", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		mealA, _ := repo.Create(meals.Meal{
			UserId: 1,
			Name:   "Meal A",
			Attributes: meals.MealAttributes{
				Quick:  true,
				Family: true,
				Easy:   true,
			},
			Ingredients: []meals.MealIngredient{
				{
					Id:     1,
					IsMain: true,
				},
			},
		})

		mealB, _ := repo.Create(meals.Meal{
			UserId: 1,
			Name:   "Meal B",
			Attributes: meals.MealAttributes{
				Quick:  false,
				Family: false,
				Easy:   false,
			},
			Ingredients: []meals.MealIngredient{
				{
					Id:     4,
					IsMain: true,
				},
			},
		})

		mealC, _ := repo.Create(meals.Meal{
			UserId: 1,
			Name:   "Meal C",
			Attributes: meals.MealAttributes{
				Quick:  false,
				Family: true,
				Easy:   true,
			},
			Ingredients: []meals.MealIngredient{
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

		repo.Create(meals.Meal{
			UserId: 2,
			Name:   "Meal D",
			Attributes: meals.MealAttributes{
				Quick:  true,
				Family: true,
				Easy:   true,
			},
			Ingredients: []meals.MealIngredient{
				{
					Id:     1,
					IsMain: true,
				},
			},
		})

		err := repo.AssignToDate(mealB.Id, time.Date(2025, time.March, 9, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		err = repo.AssignToDate(mealC.Id, time.Date(2025, time.March, 11, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		type testCase struct {
			label         string
			filters       meals.MealFilter
			expectedMeals []int
		}

		testCases := []testCase{
			{
				label: "All attributes true",
				filters: meals.MealFilter{
					UserId: 1,
					Quick:  toPtr(true),
					Family: toPtr(true),
					Easy:   toPtr(true),
				},
				expectedMeals: []int{mealA.Id},
			},
			{
				label: "Quick false",
				filters: meals.MealFilter{
					UserId: 1,
					Quick:  toPtr(false),
				},
				expectedMeals: []int{mealB.Id, mealC.Id},
			},
			{
				label: "Exclude ingredient",
				filters: meals.MealFilter{
					UserId:                1,
					ExcludeMainIngredient: []int{1},
				},
				expectedMeals: []int{mealB.Id, mealC.Id},
			},
			{
				label: "Exclude recent",
				filters: meals.MealFilter{
					UserId: 1,
					DateRange: &meals.DateRange{
						End: toPtr(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC)),
					},
				},
				expectedMeals: []int{mealA.Id, mealB.Id},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.label, func(t *testing.T) {
				meals, err := repo.List(testCase.filters)

				if err != nil {
					t.Errorf("Unexpected error %v", err)
				}

				if len(meals) != len(testCase.expectedMeals) {
					t.Errorf("Expected %d results, got %d", len(testCase.expectedMeals), len(meals))
				}

				for _, meal := range meals {
					if !slices.Contains[[]int, int](testCase.expectedMeals, meal.Id) {
						t.Errorf("Unexpected meal ID %d in results", meal.Id)
					}
				}
			})
		}
	})

	t.Run("Must include UserId when filtering meals", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		_, err := repo.List(meals.MealFilter{})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &meals.ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorMealFilterInvalid{}, err, err)
		}

	})

	t.Run("filter field DateRange validation", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		futureTime := time.Now().Add(time.Hour * 12)

		// Start date should not be in the future
		_, err := repo.List(meals.MealFilter{
			UserId: 1,
			DateRange: &meals.DateRange{
				Start: &futureTime,
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &meals.ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorMealFilterInvalid{}, err, err)
		}

		// End date should not be in the future
		_, err = repo.List(meals.MealFilter{
			UserId: 1,
			DateRange: &meals.DateRange{
				End: &futureTime,
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &meals.ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorMealFilterInvalid{}, err, err)
		}

		// End date can't be before start date
		_, err = repo.List(meals.MealFilter{
			UserId: 1,
			DateRange: &meals.DateRange{
				Start: toPtr(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC)),
				End:   toPtr(time.Date(2024, time.March, 10, 0, 0, 0, 0, time.UTC)),
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if !errors.As(err, &meals.ErrorMealFilterInvalid{}) {
			t.Errorf("Expected error of type %T, got %T (%v)", meals.ErrorMealFilterInvalid{}, err, err)
		}
	})

	t.Run("Can assign a meal to a given date and fetch meals from a range", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		chickenPie, _ := repo.Create(meals.Meal{
			Name:   "Chicken Pie",
			UserId: 1,
		})
		pizza, _ := repo.Create(meals.Meal{
			Name:   "Pizza",
			UserId: 2,
		})
		pestoSalmon, _ := repo.Create(meals.Meal{
			Name:   "Pesto Salmon",
			UserId: 1,
		})

		err := repo.AssignToDate(chickenPie.Id, time.Date(2025, time.March, 5, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		err = repo.AssignToDate(pizza.Id, time.Date(2025, time.March, 5, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		err = repo.AssignToDate(pestoSalmon.Id, time.Date(2025, time.March, 15, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Errorf("Unexpected error when assigning date: %v", err)
		}

		dateRange := meals.DateRange{
			Start: toPtr(time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)),
			End:   toPtr(time.Date(2025, time.March, 9, 0, 0, 0, 0, time.UTC)),
		}

		meals, err := repo.List(meals.MealFilter{
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

	t.Run("Can filter a list of ingredients by name", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		mealA := meals.Meal{
			Name:   "Bolognese",
			UserId: 1,
			Ingredients: []meals.MealIngredient{
				{
					Name: "Beef mince",
				},
				{
					Name: "Tinned tomatoes",
				},
				{
					Name: "Garlic",
				},
				{
					Name: "Onion",
				},
			},
		}

		mealB := meals.Meal{
			Name:   "Stir fry",
			UserId: 1,
			Ingredients: []meals.MealIngredient{
				{
					Name: "Spring onion",
				},
				{
					Name: "Chicken",
				},
			},
		}

		mealC := meals.Meal{
			Name:   "Fajitas",
			UserId: 2,
			Ingredients: []meals.MealIngredient{
				{
					Name: "Chicken",
				},
				{
					Name: "Red Onion",
				},
			},
		}
		repo.Create(mealA)
		repo.Create(mealB)
		repo.Create(mealC)
		searchString := "Onio"
		ingredients, err := repo.FindIngredients(searchString, 1)

		if err != nil {
			t.Errorf("List ingredients: Unexpected error: %v", err)
		}

		if len(ingredients) != 2 {
			t.Errorf("Expected 2 results, got %d", len(ingredients))
		}

	})

	t.Run("Can update the name of an ingredient", func(t *testing.T) {
		repo, closeDown := i.repo()
		defer closeDown()

		meal := meals.Meal{
			Name:   "Stir fry",
			UserId: 1,
			Ingredients: []meals.MealIngredient{
				{
					Name: "Spring onin",
				},
			},
		}

		createdMeal, err := repo.Create(meal)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		newName := "Spring onion"

		err = repo.UpdateIngredient(meals.Ingredient{Id: createdMeal.Ingredients[0].Id, UserId: 1, Name: newName})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		fetchedMeal, err := repo.Get(createdMeal.Id)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fetchedMeal.Ingredients[0].Name != newName {
			t.Errorf("Expected ingredient name to be updated to %s, found %s", newName, fetchedMeal.Ingredients[0].Name)
		}
	})
}

func TestDatabaseRepository(t *testing.T) {
	init := func() (meals.Repository, func()) {
		conn, err := sql.Open("sqlite3", "meals.db")

		if err != nil {
			log.Fatal(err)
		}

		err = database.Migrate(conn)

		if err != nil {
			log.Fatal(err)
		}

		closeDown := func() {
			os.Remove("meals.db")
		}
		return database.NewMealRepository(conn), closeDown
	}

	contract := RepositoryContract{
		init,
	}

	contract.Test(t)

}

func TestMemoryRepository(t *testing.T) {
	init := func() (meals.Repository, func()) {
		return &memory.MealRepository{
			Store:    []meals.Meal{},
			Calendar: make(map[int][]time.Time),
		}, func() {}
	}

	contract := RepositoryContract{
		init,
	}

	contract.Test(t)

}
