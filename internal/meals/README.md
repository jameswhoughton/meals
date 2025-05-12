# internal/meals

## Overview

This package contains the primary domain logic for handling meals and their descendants (ingredients and tags)

## Notable Files

- **repository.go** Definition of a Meal entity along with definition of the Repository interface.
- **service.go** Service to add/update meals, the service level is responsible for form validation.

## What is a Tag?

Tags are used to group meals together making it easier to find a suitable meal when planning your week, for example, on a Thursay I might want to have a 'Quick', 'Family' meal in which case I can filter my meal list down to meals that have both these tags.


