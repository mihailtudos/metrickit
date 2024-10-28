package handlers

/*
Package handlers provides HTTP handlers for the application.

This package implements the HTTP server layer for managing and serving
metric-related functionalities, including retrieving, updating, and displaying
metrics. It utilizes the Chi router for handling routes and supports middleware
for logging and response compression.

The primary components of this package include:

  - ServerHandler: A struct that encapsulates the services, logger,
    database connection, and template filesystem used by the handlers.

  - NewHandler: A function that initializes a new ServerHandler and registers
    the application's routes.

Key HTTP routes defined in this package:

1. GET /value/{metricType}/{metricName}:
  - Retrieves the value of a specific metric type and name.
  - Content-Type: text/plain
  - Returns the metric value or an error if not found.

2. GET /:
  - Displays collected metrics in an HTML format.
  - Content-Type: text/html

3. POST /update/{metricType}/{metricName}/{metricValue}:
  - Updates a metric value by type and name.
  - Content-Type: text/plain

4. POST /update/:
  - Handles metric updates in JSON format.

5. POST /updates/:
  - Handles batch updates for metrics.

6. GET /ping:
  - Checks the database connectivity.

7. Debug and profiling routes under /debug/pprof/:
  - Allows performance profiling of the application.

8. Swagger documentation routes:
  - Serves Swagger API documentation and UI for the application.

This package also includes error handling for unknown metric types and
logging of significant events during request processing, ensuring
robustness and maintainability.

Usage:

To use the handlers in your application, create a new instance of
ServerHandler using the NewHandler function and attach it to an HTTP server.
Example:

	services := // Initialize your service dependencies
	logger := // Initialize your logger
	dbConn := // Initialize your database connection
	secret := // Your secret key for authentication

	handler := handlers.NewHandler(services, logger, dbConn, secret)

	http.ListenAndServe(":8080", handler)
*/
