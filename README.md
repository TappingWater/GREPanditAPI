# GREPanditAPI Backend

![GREPandit Logo](./public/icon-192x192.png)

Official Domain: [GREPandit](https://grepandit.com)

Front end Logic: [GREPanditAPI](https://github.com/TappingWater/GREPanditAPI)

## Overview

This REST API, built with Golang, leverages the Echo framework and serves as the
backend for the GREPandit frontend.

-   **Language**: Golang
-   **Framework**: Echo-Go

## Installation

To set up the backend for GREPandit:

1. **Clone the API Repository**:

    ```bash
    git clone [YOUR API REPO LINK HERE]
    ```

2. **Initialize Go Modules**: Navigate to your API directory and run:

    ```bash
    go mod init
    ```

3. **Environment Configuration**: The application requires a `.env` file for
   proper configuration. Create a `.env` file in the root directory of the API
   with the following content:

    ```
    APP_ENV=YOUR_ENVIRONMENT_HERE (e.g., dev)
    DB_HOST=YOUR_DB_HOST_HERE
    DB_PORT=YOUR_DB_PORT_HERE
    DB_USER=YOUR_DB_USER_HERE
    DB_PASSWORD=YOUR_DB_PASSWORD_HERE
    DB_NAME=YOUR_DB_NAME_HERE
    DB_SSLMODE=YOUR_DB_SSLMODE_HERE (e.g., disable)
    AWS_COGNITO_URL=YOUR_COGNITO_URL_HERE
    ```

    _Note_: For development purposes, use `dev` for the `APP_ENV` value. Ensure
    that you're using your configurations and not sharing sensitive values
    publicly.

4. **PostgreSQL Instance**: The application requires a running instance of
   PostgreSQL. Ensure that you have PostgreSQL set up and running, and that the
   connection details in the `.env` file are correct.
5. **Run Development Server**:
    ```bash
     go run application.go
    ```

## Structure

The code structure promotes separation of concerns by organizing into distinct
layers:

-   **Handlers**: Manage HTTP requests and responses
-   **Services**: Incorporate business logic
-   **Models**: Define data structures
-   **Database**: Oversee database connections and schema
-   **Middleware**: Extend additional functionality or processing

### Entry Point

`application.go` is the primary entry point. It initializes the Echo framework,
configures middleware and the database, registers routes and handlers, and
starts the server.

### Directories

-   **internal/handlers/**: Contains HTTP handlers for API endpoints.
-   **internal/models/**: Houses data structures representing domain models.
-   **internal/services/**: Stores the application's business logic.
-   **internal/database/**: Manages database connections and table/index
    initialization.
-   **internal/middleware/**: Contains custom middleware.

### Dependency Management

-   **go.mod & go.sum**: Auto-generated by Go to manage project dependencies.

# Endpoints

The API has several endpoints organized into different groups:

## VerbalQuestion Endpoints

-   **Base URL**: `/vbquestions`

| Method | Endpoint    | Description                         |
| ------ | ----------- | ----------------------------------- |
| POST   | `/`         | Create a new verbal question        |
| GET    | `/:id`      | Retrieve a specific verbal question |
| GET    | `/adaptive` | Fetch adaptive questions            |
| GET    | `/vocab`    | Fetch questions based on vocabulary |
| POST   | `/random`   | Fetch random questions              |
| GET    | `/`         | Retrieve all verbal questions       |

## Word Endpoints

-   **Base URL**: `/words`

| Method | Endpoint      | Description             |
| ------ | ------------- | ----------------------- |
| POST   | `/`           | Create a new word       |
| PATCH  | `/marked`     | Mark words              |
| GET    | `/marked`     | Retrieve marked words   |
| GET    | `/:id`        | Fetch word by ID        |
| GET    | `/word/:word` | Fetch word by word text |

## User Endpoints

-   **Base URL**: `/users`

| Method | Endpoint             | Description                               |
| ------ | -------------------- | ----------------------------------------- |
| POST   | `/`                  | Create a new user                         |
| GET    | `/`                  | Retrieve user details                     |
| POST   | `/marked-words`      | Add marked words                          |
| POST   | `/marked-questions`  | Add marked verbal questions               |
| DELETE | `/marked-words`      | Remove marked words                       |
| DELETE | `/marked-questions`  | Remove marked verbal questions            |
| GET    | `/marked-words`      | Get marked words by user token            |
| GET    | `/marked-questions`  | Get marked verbal questions by user token |
| GET    | `/problematic-words` | Get problematic words by user token       |

## UserVerbalStat Endpoints

-   **Base URL**: `/verbal-stats`

| Method | Endpoint | Description                         |
| ------ | -------- | ----------------------------------- |
| POST   | `/`      | Create user verbal stats            |
| GET    | `/`      | Retrieve verbal stats by user token |

## Authentication

Authentication is implemented using middleware that checks AWS Cognito with a
JWKS key.

## Data Models

### UserVerbalStat

```go
type UserVerbalStat struct {
	ID         int          `json:"id"`
	UserToken  string       `json:"u_id"`
	QuestionID int          `json:"question_id"`
	Correct    bool         `json:"correct"`
	Answers    []string     `json:"answers"`
	Duration   int          `json:"duration"`
	Date       time.Time    `json:"time"`
	Competence Competence   `json:"competence"`
	FramedAs   FramedAs     `json:"framed_as"`
	Type       QuestionType `json:"type"`
	Difficulty Difficulty   `json:"difficulty"`
	Vocabulary []Word       `json:"vocabulary"`
}
```

### UserMarkedWord

```go
type UserMarkedWord struct {
	ID        int    `json:"id"`
	UserToken string `json:"user_token"`
	WordID    int    `json:"word_id"`
	Word      Word   `json:"word"`
}
```

### UserMarkedVerbalQuestion

```go
type UserMarkedVerbalQuestion struct {
	ID               int    `json:"id"`
	UserToken        string `json:"user_token"`
	VerbalQuestionID int    `json:"verbal_question_id"`
}
```

### User

```go
type User struct {
	ID            int            `json:"id"`
	Token         string         `json:"token"`
	Email         string         `json:"email"`
	VerbalAbility map[string]int `json:"verbal_ability"`
}
```

### VerbalQuestion

```go
type VerbalQuestion struct {
	ID           int               `json:"id"`
	Competence   Competence        `json:"competence"`
	FramedAs     FramedAs          `json:"framed_as"`
	Type         QuestionType      `json:"type"`
	Paragraph    string            `json:"paragraph"`
	Question     string            `json:"question"`
	Options      []Option          `json:"options"`
	Difficulty   Difficulty        `json:"difficulty"`
	Vocabulary   []Word            `json:"vocabulary"`
	VocabWordMap map[string]string `json:"wordmap"`
}
```

### VerbalQuestionRequest

```go
type VerbalQuestionRequest struct {
	ID         int          `json:"id"`
	Competence Competence   `json:"competence"`
	FramedAs   FramedAs     `json:"framed_as"`
	Type       QuestionType `json:"type"`
	Paragraph  string       `json:"paragraph"`
	Question   string       `json:"question"`
	Options    []Option     `json:"options"`
	Difficulty Difficulty   `json:"difficulty"`
	Vocabulary []string     `json:"vocabulary"`
}
```

### RandomQuestionsRequest

```go
type RandomQuestionsRequest struct {
	Limit        int          `json:"limit"`
	QuestionType QuestionType `json:"type,omitempty"`
	Competence   Competence   `json:"competence,omitempty"`
	FramedAs     FramedAs     `json:"framed_as,omitempty"`
	Difficulty   Difficulty   `json:"difficulty,omitempty"`
	ExcludeIDs   []int        `json:"exclude_ids,omitempty"`
}
```

### Meaning

```go
type Meaning struct {
	Meaning string `json:"meaning"`
	Type    string `json:"type"`
}
```

### Word

```go
type Word struct {
	ID       int       `json:"id"`
	Word     string    `json:"word"`
	Meanings []Meaning `json:"meanings"`
	Examples []string  `json:"examples"`
	Marked   bool      `json:"marked"`
}
```

### WordMap

```go
type WordMap struct {
	BaseForm  string `json:"base_form"`
	Variation string `json:"variation"`
}
```

### MarkWordsReq

```go
type MarkWordsReq struct {
	Words []string `json:"words"`
}
```

## License

This work is licensed under a Creative Commons Attribution-NonCommercial 4.0
International License. See [LICENSE](LICENSE.md) for more details.
