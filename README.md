# Toggl Backend Unattended Programming Test
This file serves as base knowledge on how to utilize the Go source code written in here

# Summary
This source code mainly functions as a program that handles Decks of Card Game(s).
There are three main functionality:
1. `Create a Deck`
- Initializing a Deck to be used for a Card Game
2. `Open a Deck`
- Open a created Deck and see what Cards are available on it
3. `Draw Card from a Deck`
- Taking Card(s) from a Deck, with Last In First Out concept

# Getting Started
To run the application, do the following command:
```
go run main.go
```
By doing so, you can access the application on `localhost:80` address.
You can also import the provided `Postman` collection, where all of the request paths are already setup.

# Running Test
Run the following command to run all available test scripts
```
go test -v ./...
```

# API Documentation
### 1. `Create a Deck`
- Endpoint: `POST` `localhost:80/deck`

| Query Parameter | Possible Values | Default | Mandatory |
|-----------------|-----------------|---------|-----------|
| shuffle | true/false | false | false|
| cards | Combination of `A/2/3/4/5/6/7/8/9/10/J/Q/K` + `C/D/H/S`| null | false

### 2. `Open a Deck`
- Endpoint: `GET` `localhost:80/deck/:deck_id`
### 3. `Draw Card from a Deck`
- Endpoint: `GET` `localhost:80/deck/:deck_id/draw`

| Query Parameter | Possible Values | Default | Mandatory |
|-----------------|-----------------|---------|-----------|
| count | any integer | 1 | false|
