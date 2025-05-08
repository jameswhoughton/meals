# Meal Planner

This is a learning project which is currently in progress, as such some features are unfinished or may change.

## Introduction

The aim of the application is to simplify the process of planning the weekly food shop by providing a single recipe store, a way to assign meals to days and generate a shopping list.

This is not meant to be used in production (although I plan to use it personally).

## Architecture Overview

The project is split across multiple packages (In some you will find additional README files). 

### internal/

The business logic is all contained within the internal packages, split by domain. Here you will find the the repository interfaces and domain services.

### internal/auth

This package contains all business logic related to interacting with a user and session management.

### internal/meals

The meals package contains all logic related to interacting with a meal.

### database

The database package contains implementations of the internal repositories using SQLite as a store.

### memory

The memory package contains in memory implementations of the repositories defined in internal, these are intended for use in tests only.

### web

This package contains all handlers, frontend assets and templates to build the UI. As the UI is fairly straight forward, I have decided to not reach for a frontend framework, instead using a few custom web components where necessary.

### cmd/server

This is the main web app executeable.

## Status

- [x] New user can register
- [ ] Existing user can:
    - [x] Log in
    - [x] Udate their account information
    - [ ] Reset their password
    - [x] Add a meal
    - [x] Edit a meal
    - [x] Delete a meal
    - [x] Edit ingredients
    - [x] Edit tags
    - [x] Assign a meal to a day
    - [ ] Generate an ingredients list

## Build and run

- Install Go.
- Install Node (required to build Tailwind assets).
- Install Make.
- Navigate to the root of the project and run `make build`.
- Run `./meals_server`. This should create and migrate a SQLite DB (`meals.db`) in the same directory.
- Navigate to `localhost:8000` in your browser.
