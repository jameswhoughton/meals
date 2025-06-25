# Meal Planner

This is a learning project which is currently in progress, as such some features are unfinished or may change.

## Introduction

The aim of the application is to simplify the process of planning the weekly food shop by providing a single recipe store, a way to assign meals to days and generate a shopping list.

## Architecture Overview

The project follows a layered architecture:

- **Repositories** Interact with the storage medium.
- **Services** Contain all the the business logic.
- **Handlers** Format data to and from the service layer.

### internal/account

Contains methods to create and update users.

### internal/meals

Contains methods to create and update meals.

### internal/planner

Contains the methods to assign meals to a date.

### database

The database package contains implementations of the `internal/*` repositories using MySQL as a store. Also contained here are the migrations files used to build the database schema. 

### memory

This package contains in memory implementations of the repositories defined in `internal/*`, these are intended for use in tests only. Each repository has it's own contract test which is used to ensure consistent behaviour between the memory and MySQL implimentations, allowing them to be confidently used when testing other layers of the application (e.g. services and handlers).

### web

In the web package the handlers are defined, which use the services/repositories defined in `internal/*` to create a web interface for the application. Also contained in this package are html template files along with non-go assets (e.g. JS, CSS and images).

### frontend

Contains dependencies to build TailwindCSS styles.

### cmd/server

This is the main web app executeable. Here the repositories/services and web server are created and the server is started. The graceful shutdown process is also handled here.

### cmd/seeder

This is a seeder script to populate the application with test data (see `cmd/seeder/main.go` for more details).

## Status

- [x] New user can register
- [ ] Existing user can:
    - [x] Log in
    - [x] Log out
    - [x] Udate their account information
    - [ ] Reset their password
    - [x] Add a meal
    - [x] Edit a meal
    - [x] Delete a meal
    - [x] Assign a meal to a day
    - [x] Generate an ingredients list
    - [ ] Export/Import meals

## Dev Dependencies

- Go
- Node
- Make
- Docker

## Build and run

- Run `docker compose up -d --build` to spin up the dev and testing MySQL servers and build the app image.
- Navigate to `localhost:8005` in your browser.

## Active development

- Install [wgo](https://github.com/bokwoon95/wgo)
- Run `docker compose up -d`.
- copy `.env-example` to `.env`.
- Run `make watch` this will start the server and restart on any file changes.
- Run `make watch-tw` this will start the tailwind server and automatically restart on any file changes.
- Navigate to `localhost:8000` in your browser.

## (Optional) Seed the DB

If desired, the DB can be seeded with one or more fake users (each with their own selection of meals).

- run `go run ./cmd/seeder -user-count=N` where `N` is the number of users to create.
