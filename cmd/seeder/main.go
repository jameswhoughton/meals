package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jameswhoughton/meals/database"
	"github.com/jameswhoughton/meals/internal/account"
	"github.com/jameswhoughton/meals/internal/meals"
)

func randomSelectionOfMealNames(n int) []string {
	meals := []string{"Spaghetti Bolognese", "Chicken Alfredo", "Beef Stroganoff", "Grilled Salmon with Asparagus", "Fish Tacos", "Vegetable Stir Fry", "Chickpea Curry", "Lentil Soup", "Mushroom Risotto", "Butternut Squash Soup",
		"Stuffed Bell Peppers", "Shepherd’s Pie", "Chicken Tikka Masala", "Tofu Pad Thai", "Shrimp Scampi", "Eggplant Parmesan", "Vegetarian Chili", "Beef Burrito", "Veggie Burrito", "Fried Rice",
		"Cauliflower Tacos", "Sweet Potato and Black Bean Bowl", "Quinoa Salad", "Macaroni and Cheese", "Chicken Caesar Salad", "Greek Salad with Feta", "Ramen with Soft-Boiled Egg", "Pho", "Lasagna", "Zucchini Noodles with Pesto",
		"BBQ Pulled Pork Sandwich", "Grilled Cheese and Tomato Soup", "Shakshuka", "Fajitas", "Enchiladas", "Falafel Wrap", "Bibimbap", "Gnocchi with Marinara", "Coconut Chickpea Stew", "Baked Ziti",
		"Spinach and Ricotta Cannelloni", "Teriyaki Chicken Bowl", "Turkey Meatballs with Spaghetti", "Vegan Buddha Bowl", "Pasta Primavera", "Miso Soup with Tofu", "Pan-Seared Tuna Steak", "Roasted Veggie Grain Bowl",
		"Chicken Fried Rice", "Vegetarian Sushi Rolls", "Lamb Kebabs", "Paneer Butter Masala", "Thai Green Curry", "Baked Potatoes with Toppings", "Cabbage Stir Fry", "Pita with Hummus and Veggies"}

	selection := make([]string, n)

	randomIndices := rand.Perm(len(meals))[0:n]

	for i := range n {
		selection[i] = meals[randomIndices[i]]
	}

	return selection
}

func randomSelectionOfTags(n int) []meals.Tag {
	tags := []string{"Vegetarian", "Vegan", "Gluten-Free", "Dairy-Free", "Low-Carb", "Keto", "Paleo", "High-Protein", "Low-Fat", "Whole30",
		"Quick", "Easy", "One-Pot", "5 Ingredients or Less", "Meal Prep", "30-Minute Meal", "No-Cook", "Freezer-Friendly",
		"Comfort Food", "Healthy", "Clean Eating", "Low-Calorie", "Budget-Friendly", "Kid-Friendly", "Family Meal", "Party Food",
		"Spicy", "Mild", "Savory", "Sweet", "Smoky", "Tangy", "Rich",
		"Breakfast", "Brunch", "Lunch", "Dinner", "Snack", "Side Dish", "Main Course", "Appetizer", "Dessert",
		"Asian", "Italian", "Mexican", "Indian", "Mediterranean", "Middle Eastern", "American", "French", "Thai", "Japanese",
		"BBQ", "Grilled", "Baked", "Fried", "Roasted", "Steamed", "Raw", "Slow Cooker", "Instant Pot", "Air Fryer"}

	selection := make([]meals.Tag, n)

	randomIndices := rand.Perm(len(tags))[0:n]

	for i := range n {
		selection[i] = meals.Tag{
			Name: tags[randomIndices[i]],
		}
	}

	return selection
}

func randomSelectionOfIngredients(n int) []meals.Ingredient {
	ingredients := make([]meals.Ingredient, n)

	mainIngredients := []string{"Ground beef", "Ribeye steak", "Sirloin steak", "Flank steak", "Filet mignon", "T-bone steak", "Beef stew meat", "Cubed chuck", "Beef round", "Beef brisket", "Flat cut brisket", "Point cut brisket", "Beef short ribs", "English cut short ribs", "Flanken cut short ribs",
		"Pork chops", "Bone-in pork chops", "Boneless pork chops", "Thick-cut pork chops", "Pork tenderloin", "Marinated pork tenderloin", "Ground pork", "Pork shoulder", "Boston butt", "Picnic shoulder", "Bacon", "Streaky bacon", "Back bacon", "Turkey bacon", "Ham", "Smoked ham", "Honey-roast ham", "Cured ham", "Sausages", "Italian sausage", "Bratwurst", "Chorizo", "Breakfast links",
		"Lamb chops", "Loin lamb chops", "Rib lamb chops", "Shoulder lamb chops", "Ground lamb", "Lamb shank", "Bone-in lamb shank", "Boneless lamb shank", "Leg of lamb", "Bone-in leg of lamb", "Boneless leg of lamb", "Butterflied leg of lamb",
		"Chicken breast", "Boneless skinless chicken breast", "Bone-in chicken breast", "Chicken tenderloins", "Chicken thighs", "Boneless skinless chicken thighs", "Bone-in chicken thighs", "Whole chicken", "Fryer chicken", "Roaster chicken", "Chicken drumsticks", "Skin-on chicken drumsticks", "Skinless chicken drumsticks", "Ground chicken",
		"Turkey breast", "Whole turkey breast", "Boneless turkey breast", "Sliced turkey breast", "Ground turkey", "Duck", "Whole duck", "Duck breast", "Duck confit",
		"Veal cutlets", "Veal chops", "Ground veal", "Venison steaks", "Rabbit legs", "Bison ground", "Bison steaks", "Roast beef", "Smoked turkey", "Pastrami", "Salami",
		"Cod", "Atlantic cod", "Pacific cod", "Haddock fillets", "Haddock loins", "Tilapia", "Halibut steaks", "Halibut fillets", "Sea bass", "Chilean sea bass", "European sea bass",
		"Salmon", "Atlantic salmon", "Sockeye salmon", "Coho salmon", "Salmon fillets", "Salmon steaks", "Tuna", "Fresh tuna steaks", "Canned tuna in oil", "Canned tuna in water", "Mackerel", "Fresh mackerel", "Smoked mackerel", "Canned mackerel", "Sardines", "Canned sardines in oil", "Canned sardines in water", "Canned sardines in tomato sauce",
		"Shrimp", "Raw shrimp", "Cooked shrimp", "Peeled shrimp", "Jumbo shrimp", "Prawns", "Bay scallops", "Sea scallops", "Mussels", "Live mussels", "Cooked mussels", "Shelled mussels", "Clams", "Fresh clams", "Canned clams", "Littleneck clams", "Cherrystone clams", "Oysters", "Raw oysters", "Smoked oysters", "Shucked oysters",
		"Lobster", "Whole lobster", "Lobster tails", "Lobster claws", "Crab", "Blue crab", "Dungeness crab", "King crab", "Snow crab", "Canned crab", "Jumbo prawns", "Tiger prawns", "White prawns"}
	secondaryIngredients := []string{"Onion", "Garlic", "Bell pepper", "Carrot", "Broccoli", "Cauliflower", "Spinach", "Kale", "Zucchini", "Mushroom", "Tomato", "Cherry tomato", "Cucumber", "Green beans", "Eggplant", "Sweet potato", "White potato", "Corn", "Peas", "Lettuce", "Arugula", "Cabbage", "Brussels sprouts",
		"Rice", "Brown rice", "White rice", "Jasmine rice", "Basmati rice", "Pasta", "Spaghetti", "Fusilli", "Penne", "Macaroni", "Lasagna sheets", "Quinoa", "Couscous", "Polenta", "Oats", "Tortillas", "Bread", "Whole wheat bread", "Flatbread", "Naan",
		"Chickpeas", "Black beans", "Kidney beans", "Lentils", "Green lentils", "Red lentils", "Brown lentils", "Cannellini beans", "Pinto beans", "Edamame",
		"Tofu", "Firm tofu", "Silken tofu", "Tempeh", "Seitan",
		"Eggs", "Milk", "Almond milk", "Soy milk", "Oat milk", "Coconut milk", "Cream", "Heavy cream", "Cheddar cheese", "Mozzarella cheese", "Parmesan cheese", "Feta cheese", "Greek yogurt", "Butter", "Ghee",
		"Olive oil", "Vegetable oil", "Canola oil", "Sesame oil", "Soy sauce", "Tamari", "Miso paste", "Coconut aminos", "Tomato paste", "Canned tomatoes", "Crushed tomatoes", "Tomato sauce", "Vinegar", "Balsamic vinegar", "Apple cider vinegar", "White vinegar",
		"Salt", "Black pepper", "Paprika", "Smoked paprika", "Cumin", "Turmeric", "Coriander", "Curry powder", "Chili flakes", "Chili powder", "Oregano", "Basil", "Thyme", "Rosemary", "Parsley", "Cilantro", "Bay leaves",
		"Flour", "All-purpose flour", "Whole wheat flour", "Cornstarch", "Baking powder", "Baking soda", "Yeast", "Sugar", "Brown sugar", "Maple syrup", "Honey",
		"Lemon", "Lime", "Avocado", "Nuts", "Almonds", "Cashews", "Walnuts", "Seeds", "Sunflower seeds", "Chia seeds", "Pumpkin seeds", "Tahini", "Peanut butter", "Almond butter",
		"Vegetable broth", "Mushroom broth", "Nutritional yeast", "Plant-based milk", "Coconut cream", "Molasses"}
	quantities := []int{1, 2, 4, 10, 100, 200, 400, 600}
	units := []string{"", "g", "kg", "ml", "l"}

	ingredients[0] = meals.Ingredient{
		Name:     mainIngredients[rand.Intn(len(mainIngredients))],
		Quantity: quantities[rand.Intn(len(quantities))],
		Unit:     units[rand.Intn(len(units))],
	}

	mealIndicies := rand.Perm(len(secondaryIngredients))[0 : n-1]

	for i := 1; i < n; i++ {
		ingredients[i] = meals.Ingredient{
			Name:     secondaryIngredients[mealIndicies[i-1]],
			Quantity: quantities[rand.Intn(len(quantities))],
			Unit:     units[rand.Intn(len(units))],
		}
	}

	return ingredients
}

func confirm() bool {
	r := bufio.NewReader(os.Stdin)

	fmt.Fprintln(os.Stdout, "Warning, running the seeder will remove all existing users from the DB, are you sure you want to continue? (yes/no)")

	s, _ := r.ReadString('\n')
	s = strings.TrimSpace(strings.ToLower(s))

	if s == "yes" {
		return true
	}

	return false
}

func main() {
	// User should explicitly confirm they want to run this as it is destructive
	if !confirm() {
		log.Println("Exiting")

		return
	}
	dsn := os.Getenv("MEALS_DB_USERNAME")

	if os.Getenv("MEALS_BD_PASSWORD") != "" {
		dsn += ":" + os.Getenv("MEALS_DB_PASSWORD")
	}

	dsn = fmt.Sprintf(
		"%s@tcp(%s:%s)/meals?parseTime=true",
		dsn,
		os.Getenv("MEALS_DB_HOST"),
		os.Getenv("MEALS_DB_PORT"),
	)

	conn, err := sql.Open("mysql", dsn)
	defer conn.Close()

	if err != nil {
		panic(err)
	}

	userCountPtr := flag.Int("user-count", 1, "How many users to seed")

	flag.Parse()

	accountRepository := database.NewAccountRespository(conn)
	mealRepository := database.NewMealRepository(conn)

	accountService := account.NewService(accountRepository)
	mealService := meals.NewService(mealRepository)

	maxMealsPerUser := 40
	maxIngredientsPerMeal := 15
	maxTagsPerMeal := 5

	// Seed rand with the time
	rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range *userCountPtr {
		email := "user-" + strconv.Itoa(i) + "@example.com"

		existingUser, _ := accountRepository.Get(context.Background(), account.GetForm{Email: &email})

		if existingUser.Id > 0 {
			log.Printf("Deleting user %s\n", email)
			err := accountRepository.Delete(context.Background(), existingUser.Id)

			if err != nil {
				log.Fatalf("error deleting user: %v", err)
			}
		}
		form := account.UserFormCreate{
			Email:           email,
			Password:        "password123",
			PasswordConfirm: "password123",
			Name:            "seeded user " + strconv.Itoa(i),
		}

		log.Printf("Creating user %s\n", email)
		user, err := accountService.CreateUser(context.Background(), &form)

		if err != nil {
			log.Fatalf("error seeding user: %v, %+v", err, form)
		}

		randomMealNames := randomSelectionOfMealNames(maxMealsPerUser)

		for j := range maxMealsPerUser {
			meal := meals.Meal{
				Name:        randomMealNames[j],
				UserId:      user.Id,
				Tags:        randomSelectionOfTags(rand.Intn(maxTagsPerMeal) + 1),
				Ingredients: randomSelectionOfIngredients(rand.Intn(maxIngredientsPerMeal) + 1),
			}

			mealService.CreateMeal(context.Background(), &meal)
		}
	}
}
