# Marketplace Backend

This is a REST API for a minimal marketplace, built with Go, following clean architecture principles.

## How to Run (Local Developer)

1.  Ensure you have Docker and Docker Compose installed.

2.  Copy the example environment file. The first time you run `make docker-up`, this will be done for you. If you need to reset it, you can run:
    ```bash
    cp .env.example .env
    ```

3.  Fill in the required secrets in the `.env` file, especially `JWT_SECRET`.

4.  Start the application and its dependencies:
    ```bash
    make docker-up
    ```

5.  Check the health of the application:
    ```bash
    curl http://localhost:8080/v1/health
    ```

6.  Run tests:
    ```bash
    make test
    ```

7.  Stop the application and remove the database volume:
    ```bash
    make docker-down
    ```
