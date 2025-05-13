# Meal Planner

This is a learning project which is currently in progress, as such some features are unfinished or may change.

## Introduction

The aim of the application is to simplify the process of planning the weekly food shop by providing a single recipe store, a way to assign meals to days and generate a shopping list.

This is not meant to be used in production (although I plan to use it personally).

## Architecture Overview

The project is split across multiple packages with the business logic located in internal/*. I have used the service/repository pattern where the repositories are solely responsible for saving to the data store and the services are responsible for ensuring the data passed to and from the repository is valid.

### internal/account

Contains methods to create and update users.

### internal/meals

Contains methods to create and update meals.

### internal/planner

Contains the methods to assign meals to a date.

### database

The database package contains implementations of the `internal/*` repositories using SQLite as a store. Also contained here are the migrations files used to build the database schema. 

### memory

This package contains in memory implementations of the repositories defined in `internal/*`, these are intended for use in tests only. Each repository has it's own contract test which is used to ensure consistent behaviour between the memory and SQLite implimentations, allowing them to be confidently used when testing other layers of the application (e.g. services and handlers).

### web

In the web package the handlers are defined, which use the services/repositories defined in `internal/*` to create a web interface for the application. Also contained in this package are html template files along with non-go assets (e.g. JS, CSS and images).

### cmd/server

This is the main web app executeable. Here the repositories/services and web server are created and the server is started. The graceful shutdown process is also handled here.

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
- Install Docker.
- Navigate to the root of the project and run `make build`.
- Run `docker compose up -d` to spin up the dev and testing MySQL servers.
- Run `./meals_server`. This should create and migrate a SQLite DB (`meals.db`) in the same directory.
- Navigate to `localhost:8000` in your browser.
