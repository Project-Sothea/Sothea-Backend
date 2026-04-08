# Backend Developer Documentation

### Last Updated: 8 Apr, 2026

For future developers, this document will serve as a guide to understanding the backend and how to work with it, as well
as the numerous design choices made.

## Overview

Sothea-Backend provides a REST API for Sothea-Frontend to interact with, to do CRUD operations on patients.
It is written in Go, due to its speed, reliability and ease of use. The backend uses a PostgreSQL database to store
patient data.  
It also draws inspiration
from [Bob's Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).
Some key concepts used include Dependency Rules and use cases, which make testing and maintenance easier.
For basic deployment instructions, refer to the README.md file.

## Entity Design

The center of the backend is the patients entity, a representation of a patient's data. A patient comprises the
following categories, each with their own fields:  
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

Patients go through the physical health screening stations with the admin station first, and the rest having no
guaranteed order.
Hence, if a patient exists, they will have a patient_details row and an admin row for each visit, but may not have the
other categories present yet.  
Additionally, patients may have multiple visits, with multiple rows for each category, representing the previous years
of visits.  
Every visit row in the patient database will have the following structure:

```
+------------+----------+-----------------------+  
| patient_id | visit_id | rest of categories    |  
+------------+----------+-----------------------+  
```

The patient id uniquely identifies a patient, while the visit id narrows down which visit the row is associated with.
A numeric visit id is used instead of the date to allow for easier querying.

## Types
To view the details of the types used in the backend, refer to this [google sheet](https://docs.google.com/spreadsheets/d/1V9VuKGOoyZ5-ul5enuHliImjRJSSkVKatQeDi4MGblU/edit?gid=0#gid=0).

## Database and Database Schema

The database used is PostgreSQL. To interact with the db, the Go database driver used is
[pgx/v5](https://github.com/jackc/pgx) with a connection pool (`pgxpool`).
Type-safe SQL queries are generated from raw SQL using [sqlc](https://github.com/sqlc-dev/sqlc). The generated code
lives in `repository/sqlc/` and should not be edited by hand — edit the source SQL in `db/queries/` and re-run
`sqlc generate` instead.

We have decided against using a full ORM due to the interesting nature of the patient entity, which may or may not have
all categories filled out, save for the admin category.

Additionally, we've tried to keep the schema simple, choosing not to compute derived fields at the database or backend
level.

## Directory Structure

```
.
├── README.md                    - Setup instructions and API documentation.
├── DOCS.md                      - This file; backend developer documentation.
├── go.mod
├── go.sum
├── .gitignore
├── .env.example                 - Template for required environment variables.
├── Dockerfile                   - PostgreSQL image with initialization scripts.
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
└── uploads/                     - Patient photo storage root (uploads/patient/<id>).
```

Test files have been excluded from the directory structure for brevity.

## Error Handling

We have defined some custom errors in `entities/errors.go`.  
They serve to make passing errors around easier, and to provide more context to the error, such as whether a Patient or
PatientVisit was not found.

HTTP Error Codes Used:

- 400: Bad Request
- 401: Unauthorized
- 404: Not Found
- 409: Conflict (e.g. duplicate drug name or batch number)
- 500: Internal Server Error

## Middleware

The backend uses two middleware components in `controllers/middleware/`:

- **auth.go**: Checks for a valid JWT Bearer token in the Authorization header and stashes the `userID` and `username`
  in the Gin context for downstream handlers.
- **tx.go**: Wraps a handler in a database transaction, sets the `sothea.user_id` session variable for audit triggers,
  and commits on success or rolls back on any error or abort.

## Configuration

The backend reads configuration from environment variables at startup using Viper and godotenv.
Copy `.env.example` to `.env` and fill in the required values:

| Variable       | Description                                      |
|----------------|--------------------------------------------------|
| `PORT`         | Port the HTTP server listens on (e.g. `9090`)    |
| `DATABASE_URL` | PostgreSQL connection string                     |
| `SECRET_KEY`   | Secret key used to sign and verify JWT tokens    |

## Testing

The backend uses the standard Go testing package for testing.
Testing is done at the controller and repository levels, with the use of mocks to simulate the usecases.

For testing at the controller level, we used [vektra/mockery](https://github.com/vektra/mockery) to generate mocks for
the usecases and repositories.
For testing at the repository level, we opted for using [Dockertest](https://github.com/ory/dockertest) to spin up a
temporary PostgreSQL container for testing, and to run the tests against it.
This is to ensure that the data access layer, which is far more complex to mock, performs exactly as expected.
See why you probably shouldn't mock the
database [here](https://dominikbraun.io/blog/you-probably-shouldnt-mock-the-database/).

## Naming and Code Conventions

SQL Fields: snake_case  
e.g. `consultation_notes`

Golang Struct Fields: CamelCase with first letter capitalised  
e.g. `RegDate`

JSON Fields: snake_case  
e.g. `past_medical_history`, `reg_date`

## Miscellaneous Design Choices

### Using Pointers instead of Defined Null Types

We have chosen to use pointers for fields that can be null in the database, such as `dob`, `drug_allergies`,
`last_menstrual_period`, etc.  
This allows a field to represent three states: explicitly set to a value, explicitly set to null (`nil` pointer in JSON),
or omitted.

For primitive types like bool, they can only take on 2 values: true, false. When JSON is unmarshalled into a struct, a
missing field gets set to the zero value (false for booleans). With `binding: required`, the validator cannot
distinguish between an explicitly-set false and a missing field.

The workaround is to use a pointer (`*bool`), which can be `nil` when absent, `&false` when explicitly false, and
`&true` when explicitly true.

### Pharmacy and Prescription Modules

Beyond the patient module, the backend also manages a pharmacy inventory and prescription lifecycle:

- **Pharmacy**: Tracks drugs, batches, and batch locations. Stock quantities are maintained per batch location.
- **Prescriptions**: Each prescription belongs to a patient visit. A prescription has one or more lines (one drug per
  line). Lines have allocations that reserve stock from specific batch locations. A line must be fully allocated and then
  packed before the prescription can be dispensed.
- **DB Triggers**: Stock reservation/release is handled by database triggers on the `prescription_batch_items` table,
  keeping stock accurate without manual bookkeeping in Go code.
