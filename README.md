# private-markets-api

A small Go REST API for managing private market funds, investors, and investments with PostgreSQL persistence.

## Setup

1. Prerequisites:
- Install and/or have docker running on your machine

2. Copy or verify environment values in `.env`:

```env
POSTGRES_DB=markets_db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
DATABASE_URL=postgres://postgres:postgres@postgres:5432/markets_db
```

3. Start the application:

```bash
docker-compose up
```

## API Reference

Base URL: `http://localhost:8080/api/v1/private-markets`

### Funds

- `GET /funds`
  - List all funds.

- `POST /funds`
  - Create a new fund.
  - Request body:
    ```json
    {
      "name": "Growth Fund I",
      "vintage_year": 2025,
      "target_size_usd": 150000000,
      "status": "fundraising"
    }
    ```
  - Valid `status` values: `fundraising`, `investing`, `closed`
  - Returns `409 Conflict` if a fund with the same `name` already exists.

- `PUT /funds`
  - Update an existing fund.
  - Request body:
    ```json
    {
      "id": "<fund-id>",
      "name": "Growth Fund I Updated",
      "vintage_year": 2026,
      "target_size_usd": 200000000,
      "status": "investing"
    }
    ```

- `GET /funds/{id}`
  - Fetch a single fund by UUID.

### Investors

- `GET /investors`
  - List all investors.

- `POST /investors`
  - Create a new investor.
  - Request body:
    ```json
    {
      "name": "Example Investor",
      "investor_type": "individual",
      "email": "investor@example.com"
    }
    ```
  - Valid `investor_type` values: `individual`, `institutional`, `family office`
  - Email addresses must be unique and valid.
  - Returns `409 Conflict` if the email already exists.

### Investments

- `GET /funds/{fundID}/investments`
  - List investments for a given fund.

- `POST /funds/{fundID}/investments`
  - Create an investment for a fund.
  - Request body:
    ```json
    {
      "investor_id": "<investor-id>",
      "amount_usd": 50000,
      "investment_date": "2026-04-19"
    }
    ```
  - `investment_date` must be in `YYYY-MM-DD` format.

## Validation and Behavior

- Fund `name` is unique in the database.
- Investor `email` is unique in the database.
- `vintage_year` must be between 1900 and 2100 (to keep values within a reasonable historical and near-future range).
- `target_size_usd` and `amount_usd` must be positive.
- Invalid UUIDs return `404 Not Found`.
- Missing required fields return `400 Bad Request`.
- Duplicate resources return `409 Conflict`.

## Database

The schema is defined in `db/init.sql` and includes:

- `fund_statuses`
- `investor_types`
- `funds`
- `investors`
- `investments`

The PostgreSQL service uses Docker Compose to initialize the schema and seed data on first run.

## Assumptions and Design Decisions

- The API is aligned with the official schema and models documented by the interview API spec: `https://storage.googleapis.com/interview-api-doc-funds.wearebusy.engineering/index.html`
- Unique constraints are enforced at the database level for fund names and investor emails.
- Single SQL statements are used for create/update operations, so explicit transactions are not required for the current flows.
- UUID validation is performed in service code before database access to avoid Postgres `22P02` errors.
- Investment dates are validated using the strict `YYYY-MM-DD` format.

## AI Tooling Note

This project was developed with help from GitHub Copilot, using iterative instructions to:

- define clean service/repository/handler layers,
- use specific packages for logging, http router etc and structure the code,
- wire up database access with `pgx` and `pgxpool`,
- implement validation and error mapping/handling,
- translate DB constraint failures into proper HTTP responses.
