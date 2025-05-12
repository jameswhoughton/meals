# TODO

## Features

- [x] debounce request-typeahead
- [x] Make typeahead accessible
- [x] Update navigation
- [x] Create meals list page
- [x] Add delete button and route for meals page (and edit meal page)
- [x] Create page to edit meal
- [x] Add note to edit ingredient page 'This will affect all meals with this ingredient'
- [x] Convert attributes to 'tags'
    - [x] Tags are created when a meal is created
    - [x] Tags are removed if all associated meals no longer exist
    - [x] Add tags page
    - [x] Add edit tag page
    - [x] Add new TagRepository
    - [x] Update existing tests
- [x] Create planner page
    - [x] Add account config for week start day
    - [x] page listing days of the week from user's start day
    - [x] Each day shows the name of the chosen meal or a button to add one
    - [x] Days with meals have an edit/delete button
    - [x] Can search for meals by name or filter by tag
- [x] Ensure session cookie refreshes on request
- [ ] Add background pattern
- [x] Remove requirement to provide a unit
- [ ] Show success messages when creating/updating meals
- [ ] Meal/ingredient/tag pagination
- [x] Add meal handler tests
- [x] Add meal service tests
- [ ] Remove id from update meal form (and others)
- [ ] Add required labels
- [x] Ensure ingredients are unique per user
- [x] Ensure tags are unique per user
- [ ] Method to seed the DB for dev/testing
- [ ] Replace old alerts with alert-message component
- [ ] Prevent user updating name of ingredient/tag to one that already exists
- [ ] Add tests for tag handlers
- [x] Move auth handlers to web
- [ ] Add tests for account handlers
- [ ] Add confirmation modal for deletion
- [ ] Add test for SessionRepository.Refresh
- [ ] Add validation to restrict user to assigning only their own meals
- [ ] Review and polish UI
- [x] Remove ingredients list/edit pages and handlers
- [x] Remove the ingredients table
- [x] Add name field to meal_ingredients table
- [x] Update Find function in IR to search distinct ingredients in meal_ingredients
- [x] Update meal_repository to remove createIngredient
- [x] Update meals.Service to remove the populateIngredientIds function
- [x] Update meal form page to populate the name of an ingredient from the typeahead
- [x] Make the name editable for existinng ingredients from the meal page
- [ ] Add typeahead search for units

## Bugs

- [x] Main ingredient bug
- [x] Checkboxes not checked when viewing a meal

## Roadmap

- [ ] Password reset
    - [ ] Integrate email service
    - [ ] Configure mailhog/mailtrap with Docker
- [ ] Meal import/export
- [ ] Add default filters for days
- [ ] User can delete their account and data
