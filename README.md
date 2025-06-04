# Meal Planner

This is a learning project which is currently in progress, as such some features are unfinished or may change.

## Introduction

The aim of the application is to simplify the process of planning the weekly food shop by providing a single recipe store, a way to assign meals to days and generate a shopping list.

## Architecture Overview

The project is split across multiple packages with the business logic located in internal/*. I have used the service/repository pattern where the repositories are solely responsible for saving to the data store and the services are responsible for ensuring the data passed to and from the repository is valid.

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

### cmd/server

This is the main web app executeable. Here the repositories/services and web server are created and the server is started. The graceful shutdown process is also handled here.

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
- Run `make watch` this will start the server and restart on any file changes.
- Run `make watch-tw` this will start the tailwind server and automatically restart on any file changes.
- Navigate to `localhost:8000` in your browser.
