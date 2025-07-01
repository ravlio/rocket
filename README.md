# Rocket State Service - Solution Documentation

## Introduction

This document provides an overview of the design, implementation, and operational aspects of the Rocket State Service, developed as a solution to the Backend Engineer Challenge. The service aims to consume telemetry messages from various rockets, aggregate their state, and expose this information via a REST API.

## How to Run the Service

### Steps to Run

1.  **Clone the Repository:**
    ```bash
    git clone <your-repository-link>
    cd rocket-state-service
    ```

2.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Run Locally (Development):**
    ```bash
    go run ./cmd/rocket-state-service/main.go
    ```
    The service will start on `http://localhost:8088` by default

## API Documentation

The service exposes a REST API compliant with OpenAPI 3.0. The full API specification is available in `api/openapi.yaml`.

### Endpoints

* **POST `/messages`**
    * **Summary:** Ingests a new rocket telemetry message.
    * **Request Body:** `application/json` (see `TelemetryMessage` schema in `api/openapi.yaml`).
    * **Responses:**
        * `202 Accepted`: Message successfully received and accepted for processing.
        * `400 Bad Request`: Invalid message format or content (e.g., missing required fields, invalid UUID, unknown message type).
        * `500 Internal Server Error`: An unexpected error occurred during message processing.

* **GET `/v1/rockets`**
    * **Summary:** Returns a list of all rockets currently tracked by the system, along with their aggregated states.
    * **Query Parameters:**
        * `sortBy` (optional, string): Field to sort the list by. Allowed values: `id`, `type`, `speed`, `mission`, `lastUpdateTime`.
        * `sortOrder` (optional, string): Sort order. Allowed values: `asc` (default), `desc`.
    * **Responses:**
        * `200 OK`: A JSON array of `RocketState` objects.
        * `500 Internal Server Error`: An unexpected error occurred.

* **GET `/v1/rockets/{id}`**
    * **Summary:** Returns the current aggregated state of a specific rocket.
    * **Path Parameters:**
        * `id` (required, string, format: uuid): The unique identifier (channel) of the rocket.
    * **Responses:**
        * `200 OK`: A `RocketState` object.
        * `404 Not Found`: Rocket with the specified ID was not found.
        * `400 Bad Request`: Invalid UUID format for the `id` parameter.
        * `500 Internal Server Error`: An unexpected error occurred.

## Design Choices and Trade-offs

### 1. In-Memory Data Store (`InMemoryRocketStore`)

* **Choice:** For simplicity and rapid development, an in-memory map (`map[uuid.UUID]rocket.State`) is used as the data store.
* **Trade-offs:**
    * **Pros:** Extremely fast for read/write operations, easy to set up, no external dependencies (like a database).
    * **Cons:**
        * **No Persistence:** All rocket states are lost if the service restarts. This is a significant limitation for a real-world application.
        * **Limited Scalability:** Does not scale horizontally. All data resides on a single instance.
        * **Not Production-Ready:** Not suitable for production environments where data durability and high availability are required.
* **Alternative (More Complex) Solution:** For a production-grade solution, a persistent database (e.g., PostgreSQL, RocksDB, Cassandra, or a managed service like AWS DynamoDB) would be used. This would involve implementing a `PostgresRocketStore` (or similar) that interacts with a database driver, handles connection pooling, migrations, etc.

### 2. Simplified Out-of-Order / At-Least-Once Message Handling

* **Choice:** The `RocketService.ProcessMessage` method currently uses a very simplified logic: it only processes a message if its `messageNumber` is strictly greater than the `LastProcessedMessageNumber` stored for that rocket. Duplicate messages with the same `messageNumber` are ignored.
* **Trade-offs:**
    * **Pros:** Simple to implement, works for messages arriving mostly in order or where only the latest message is truly critical.
    * **Cons:**
        * **Incorrect State for Out-of-Order (Older) Messages:** If a message with `messageNumber=2` arrives *after* a message with `messageNumber=3` has already been processed, `messageNumber=2` will be **ignored**. This means the rocket's state will **not be correctly aggregated** if an older, but valid, message arrives late. For example, if message #2 changed the mission, that change would be missed.
        * **No Full Event History:** The service does not store a complete history of all messages for a rocket. It only stores the current aggregated state.
* **Alternative (More Complex) Solution:**
    * **Event Sourcing / Event Log:** Store *all* incoming messages (events) for each rocket in a persistent, ordered log (e.g., Kafka, a database table). When a new message arrives (especially an out-of-order one), the service would:
        1.  Persist the new message.
        2.  Retrieve all messages for that rocket from the log up to the highest `messageNumber` seen.
        3.  **Re-aggregate** the rocket's state by applying all messages in their strict `messageNumber` (and `messageTime` for tie-breaking) order. This ensures the state is always eventually consistent and correct.
    * **Snapshotting:** Combine event sourcing with periodic snapshots of the rocket's state to optimize re-aggregation time for long-lived rockets.
    * This approach is significantly more complex but guarantees correct state aggregation regardless of message arrival order.

## Conclusion

This Rocket State Service provides a functional REST API for tracking rocket states. The design prioritizes clarity and testability through modular components. Key trade-offs were made for simplicity (in-memory store, simplified out-of-order handling) but are explicitly acknowledged, with more robust alternatives described for a production environment. The solution is thoroughly verified with automated unit and integration tests.# rocket
