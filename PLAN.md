# TODO

## Database

### meals
- id (int)
- name (string)
- notes (text)
- quick (bool)
- family (bool)
- main_ingredient (string)

### meals_ingredients
- meal_id
- ingredient_id

### ingredients
- id (int)
- name (string)

### schedule
- id (int)
- date (date)
- meal_id (int) nullable

## Rules
- Main ingredient cannot be repeated in a week
- Meal cannot be repeated in a week
- Rules only apply to meal plan generation

## Endpoints
GET /
    - shows current mealplan for the week
    - individual or multiple days can be generated or edited
    - each day can have parameters to generate a meal (e.g. quick/family)
    - typeahead to select meals
    - button to create meal
    - calendar to select week (can select future dates)
POST /plan/{DATE}
    - save meal for a date 
GET /meals - meal list page
    - links to add/edit meals
GET /meals/{meal ID} - edit meal page
POST /meals/{meal ID} - edit meal page
    - ingredients can be added directly when editing a meal if they do not exist
    - typeahead to select ingredients
GET /ingredients - list of ingredients
POST /ingredients - add ingredient
POST /ingredients/{ingredient_id} - update ingredient

## Stretch goals
- User accounts (make public)
    - Defaults for daily meal requirements
- Import/export meals
