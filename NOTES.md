# Notes

- Explain overall structure
    - Justify memory store
- Checking meals/ingredients exist/permission in handlers rather than middleware, why?
- typeahead component
- Why sqlite?
- Session management
- Seeding the DB

## Spcification

### Meal Characteristics

- Has a single owner
- Has atleast one ingredient
- Has one main ingredient
- Can have one or more tags
- Meal names are not unique
- Associated ingredients must have a non-zero quantity

### Ingredient Characteristics

- Has a single owner
- Can belong to multiple meals
- Cannot exist if it isn't associated with a meal
- Name should be unique

### Tag Characteristics

- Has a single owner
- Can belong to multiple meals
- Cannot exist if it isn't associated with a meal
- Name should be unique

### User Characteristics

- Email address must be unique
- There are no granular permissions (only one level of user)
- Can only access meals/tags/ingredients that they own
- User sessions last 1hr after last activity

