# Overview

Language: Golang
Framework: Echo

## Structure

Code structure promotes separation of concerns by organizing the code into distinct layers:

- Handlers: Handle HTTP requests and responses
- Services: Implement business logic
- Models: Define data structures
- Database: Manage database connections and schema
- Middleware: Provide additional functionality or processing

### cmd/server/main.go

This is the main entry point of your application. It's responsible for initializing the Echo framework, setting up middleware, configuring the database, registering routes and handlers, and starting the server.

### internal/handlers/

This directory contains the HTTP handlers for different API endpoints. Each file corresponds to a specific resource or group of related resources, e.g., user_handler.go contains handlers for user-related endpoints.

### internal/models/

This directory contains the data structures (structs) representing your application's domain models, e.g., user.go defines the User struct.

### internal/services/

This directory contains the business logic for your application. Services are responsible for interacting with the database and performing any necessary data transformations or validations. Each file corresponds to a specific resource or group of related resources, e.g., user_service.go contains methods for creating, fetching, updating, and deleting users.

### internal/database/

This directory contains the logic for connecting to the database and initializing the necessary tables or indexes. The database.go file exports a function for setting up the database connection, which can be used in main.go.

### internal/middleware/

This directory contains custom middleware for your application, e.g., custom_middleware.go. Middleware can be used for various purposes, such as logging, authentication, or request/response transformation.

### go.mod and go.sum

These files are automatically generated by Go when you run go mod init and manage your project's dependencies.

## Data Models

Verbal Question model:
	Verbal Question ID: Integer: Primary Key
	Competence: String Index
		Analyzing and drawing conclusions: ex:What is the primary purpose of this passage
		Reasoning from incomplete data: ex: Which statement can be inferred from this passage
		Identifying authors assumptions/perspective: ex: What assumptions underlies the authors' arguements
		Understanding multiple levels of meaning: ex: Which of the following best describes the tone of this passage
		Selecting important info. ex: Identify the key info in this passage
		Distinguish major/minor points. ex: Which detail from the passage supports the authors primary claim
	Framed As: String Index
		MCQ with single answer
		MCQ with multiple answers
		Select sentence from passage
	Type: String Index
		Reading Comprehension
		Text Completion
		Sentence equivalance
	Paragraph ID: Foreign Key Int
	Question: String
	Options: String[], Int[] (Viable paragraphs to choose from)
	Answer: String[]
	Explanation: String
	Difficulty: String EASY/MID/HARD

Paragraph model:
	Paragraph ID: Integer : Primary Key
	Paragraph text: String

Word model:
	Word ID: Integer: Primary key
	Word: String Index
	Meanings: String[]

Verbal stats model: (NOSQL Table for write heavy)
	User ID: Integer: Key	
	Words: String[]
	Datapoints: Integer[] 
Verbal stat datapoint: (NOSQL Table)
	Datapoint ID: Integer: Key
	Type: 
	Question ID:
	Date:
	Competence:
	Difficulty:
	Type: