# Project Sothea Backend
### Last Updated: 8 Apr, 2026

## Overview

This is the backend for the patient management system for Project Sothea, and is to be set up in conjunction with the frontend.  
The backend is written in Go, and uses a PostgreSQL database to store patient data. The backend provides a RESTful API for the frontend to interact with, and is responsible for handling requests to create, read, update, and delete patient data.

---

## Prerequisites

Before you begin, ensure you have the following installed:

- [Golang](https://golang.org/) - The Go programming language.
- [PostgreSQL](https://www.postgresql.org/) - An open-source relational database system.
- [Docker](https://www.docker.com/) - A platform for building, shipping, and running applications in containers.

## Installation and Setup
1. Clone the repository to your local machine: `git clone https://github.com/Project-Sothea/Sothea-Backend.git`

2. In the project folder, build the project with `go build -o sothea-backend`

3. Set up the required docker containers for the database (see below).

4. Copy `.env.example` to `.env` and fill in the required values (`PORT`, `DATABASE_URL`, `SECRET_KEY`).

5. Copy the `dist/` folder from the frontend build into this project's root directory. The backend serves the frontend static files from `./dist` at `/`.

6. Run the Go binary with `./sothea-backend`, starting up the server.

7. The server should now be accessible on `http://localhost:3000` (or whichever port you configured in the `.env` file).

## Setting Up Docker
To facilitate easy setup of the patients database with preloaded data, we use Docker Compose with a PostgreSQL image. The database schema and seed data are automatically applied on first startup. To set up the database, follow the steps below:
1. Make sure the Docker daemon is running in the background.

2. Start the database container: `docker compose up -d`

3. To stop the container, run `docker compose down`

To also remove the database volume (full reset): `docker compose down -v`

## Configuration

Copy `.env.example` to `.env` and fill in the required values:

| Variable       | Description                                      |
|----------------|--------------------------------------------------|
| `PORT`         | Port the HTTP server listens on (e.g. `3000`)    |
| `DATABASE_URL` | PostgreSQL connection string                     |
| `SECRET_KEY`   | Secret key used to sign and verify JWT tokens    |

## Common Issues
- **Database role not found / Authentication Failed**  
  This usually happens if there are already pre-existing Postgres instances running on port 5432. Stop the existing Postgres processes.  
  If on Windows, do `Win` + `R`, then type `services.msc`, search for the Postgres service and stop it.

---

## Developer Documentation

## Entity Design

The center of the backend is the patients entity, a representation of a patient's data. A patient comprises the following categories, each with their own fields:  
Patient SQL schema: `/db/schema/patients.sql`  
Patient Golang struct schema: `/entities/models.go`

- Patient Details (demographics: name, DOB, gender, village, etc.)
- Admin (per-visit: reg date, queue number, pregnancy, etc.)
- Past Medical History
- Social History
- Vital Statistics
- Height and Weight
- Visual Acuity
- Fall Risk
- Dental
- Physiotherapy
- Doctor's Consultation

Patients go through the physical health screening stations with the admin station first, and the rest having no guaranteed order. Hence, if a patient exists, they will have a patient_details row and an admin row for each visit, but may not have the other categories present yet.  
Additionally, patients may have multiple visits, with multiple rows for each category, representing the previous years of visits.  
Every visit row in the patient database will have the following structure:

```
+------------+----------+-----------------------+  
| patient_id | visit_id | rest of categories    |  
+------------+----------+-----------------------+  
```

The patient id uniquely identifies a patient, while the visit id narrows down which visit the row is associated with. A numeric visit id is used instead of the date to allow for easier querying.

## Database and Database Schema

The database used is PostgreSQL. To interact with the db, the Go database driver used is [pgx/v5](https://github.com/jackc/pgx) with a connection pool (`pgxpool`). Type-safe SQL queries are generated from raw SQL using [sqlc](https://github.com/sqlc-dev/sqlc). The generated code lives in `repository/sqlc/` and should not be edited by hand — edit the source SQL in `db/queries/` and re-run `sqlc generate` instead.

### sqlc Configuration (`sqlc.yaml`)

Key settings in `sqlc.yaml`:

- **`emit_pointers_for_null_types: true`** — nullable database columns are generated as pointer types (`*string`, `*bool`, `*time.Time`, etc.) in the Go structs. This is what drives the pointer convention described below.
- **`emit_json_tags: true`** — JSON tags are auto-generated in snake_case on all struct fields.
- **`sql_package: pgx/v5`** — uses the pgx driver instead of `database/sql`.
- **Type overrides** — `date` and `timestamptz` columns are mapped to `time.Time` (with pointer variants for nullable columns), and `prescription_lines.dose_amount`/`duration` are mapped to `float64` instead of the default `pgtype.Numeric`.

## Directory Structure

```
.
├── README.md                    - Setup instructions and developer documentation.
├── go.mod
├── go.sum
├── .gitignore
├── .env.example                 - Template for required environment variables.
├── docker-compose.yml           - PostgreSQL container configuration.
├── sql/                         - Database initialisation scripts (run automatically on first Docker startup).
│   ├── 1_users_setup.sql
│   ├── 2_patients_setup.sql
│   ├── 3_pharmacy_setup.sql
│   └── 4_prescription_setup.sql
├── sqlc.yaml                    - sqlc code generation configuration.
├── main.go                      - Entry point; wires repos, usecases, and handlers.
├── controllers/
│   ├── middleware/
│   │   ├── auth.go              - JWT token creation, verification, and AuthRequired middleware.
│   │   └── tx.go                - Database transaction middleware (WithTx, GetTx).
│   ├── login_handler.go         - Handles login and user listing requests.
│   ├── patient_handler.go       - Handles patient CRUD and visit management.
│   ├── pharmacy_handler.go      - Handles drug, batch, and location management.
│   └── prescription_handler.go  - Handles prescription lifecycle (lines, packing, dispensing).
├── entities/
│   ├── models.go                - Aggregated view types (Patient, PatientMeta, DrugStock, Prescription, etc.)
│   └── errors.go                - Custom sentinel error definitions.
├── usecases/
│   ├── login_ucase.go           - Login logic and JWT token generation.
│   ├── patient_ucase.go         - Patient operations with context timeout.
│   ├── pharmacy_ucase.go        - Drug and batch inventory operations.
│   └── prescription_ucase.go    - Prescription workflow, FEFO allocation suggestion, dispense logic.
├── repository/
│   ├── postgres/
│   │   ├── postgres_patient.go           - PostgreSQL implementation of patient operations.
│   │   ├── postgres_pharmacy.go          - PostgreSQL implementation of pharmacy operations.
│   │   ├── postgres_prescriptions.go     - PostgreSQL implementation of prescription operations.
│   │   ├── postgres_user.go              - PostgreSQL implementation of user/auth operations.
│   │   ├── postgres_util.go              - Shared repository utilities.
│   │   ├── postgres_pharmacy_errors.go   - Custom pharmacy error types (e.g. DuplicateBatchNumberError).
│   │   └── postgres_prescription_errors.go - Custom prescription error types (e.g. InsufficientStockError).
│   └── sqlc/                    - Auto-generated type-safe SQL code. Do NOT edit manually.
│       ├── db.go
│       ├── models.go
│       ├── patient.sql.go
│       ├── pharmacy.sql.go
│       ├── prescriptions.sql.go
│       └── users.sql.go
├── db/
│   ├── schema/                  - Source SQL schema definitions.
│   │   ├── users.sql
│   │   ├── patients.sql
│   │   ├── pharmacy.sql
│   │   └── prescription.sql
│   └── queries/                 - Source SQL queries used by sqlc to generate repository/sqlc/.
│       ├── users.sql
│       ├── patient.sql
│       ├── pharmacy.sql
│       └── prescriptions.sql
├── util/
│   ├── helper.go                - Git root path helpers (GetGitRoot, MustGitPath).
│   └── media.go                 - Patient photo upload, validation, and filesystem storage.
├── uploads/                     - Patient photo storage root (uploads/patient/<id>).
└── dist/                        - Frontend static files served by the backend at `/`.
```

## Error Handling

We have defined some custom errors in `entities/errors.go`.  
They serve to make passing errors around easier, and to provide more context to the error, such as whether a Patient or PatientVisit was not found.

HTTP Error Codes Used:

- 400: Bad Request
- 401: Unauthorized
- 404: Not Found
- 409: Conflict (e.g. duplicate drug name or batch number)
- 500: Internal Server Error

## Middleware

The backend uses two middleware components in `controllers/middleware/`:

- **auth.go**: Checks for a valid JWT Bearer token in the Authorization header and stashes the `userID` and `username` in the Gin context for downstream handlers.
- **tx.go**: Wraps a handler in a database transaction, sets the `sothea.user_id` session variable for audit triggers, and commits on success or rolls back on any error or abort.

## Naming and Code Conventions

SQL Fields: snake_case  
e.g. `consultation_notes`

Golang Struct Fields: CamelCase with first letter capitalised  
e.g. `RegDate`

JSON Fields: snake_case  
e.g. `past_medical_history`, `reg_date`

## Miscellaneous Design Choices

### Using Pointers instead of Defined Null Types

We have chosen to use pointers for fields that can be null in the database, such as `drug_allergies`, `last_menstrual_period`, `no_of_years`, etc. This is handled automatically by sqlc via `emit_pointers_for_null_types: true`.  
This allows a field to represent three states: explicitly set to a value, explicitly set to null (`nil` pointer in JSON), or omitted.

For primitive types like bool, they can only take on 2 values: true, false. When JSON is unmarshalled into a struct, a missing field gets set to the zero value (false for booleans). With `binding: required`, the validator cannot distinguish between an explicitly-set false and a missing field.

The workaround is to use a pointer (`*bool`), which can be `nil` when absent, `&false` when explicitly false, and `&true` when explicitly true.

### Patient Photo Storage

Patient photos are stored on disk rather than in the database. Photos are saved to `uploads/patient/<id>` (no file extension) under the repo root. The MIME type is detected at read time from the file's magic bytes rather than from a stored extension. Files are written atomically via a `.tmp` intermediate file to prevent partial writes. The maximum allowed photo size is 5 MiB. Supported formats are JPEG, PNG, WebP, GIF, and BMP.

### Pharmacy and Prescription Modules

Beyond the patient module, the backend also manages a pharmacy inventory and prescription lifecycle:

- **Pharmacy**: Tracks drugs, batches, and batch locations. Stock quantities are maintained per batch location.
- **Prescriptions**: Each prescription belongs to a patient visit. A prescription has one or more lines (one drug per line). Lines have allocations that reserve stock from specific batch locations. A line must be fully allocated and then packed before the prescription can be dispensed.
- **DB Triggers**: Stock reservation/release is handled by database triggers on the `prescription_batch_items` table, keeping stock accurate without manual bookkeeping in Go code.
