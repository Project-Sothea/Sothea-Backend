# Project Sothea Backend
### Last Updated: 8 Apr, 2026
## Overview

This is the backend for the patient management system for Project Sothea, and is to be set up in conjunction with the frontend.  
The backend is written in Go, and uses a PostgreSQL database to store patient data. The backend provides a RESTful API for the frontend to interact with, and is responsible for handling requests to create, read, update, and delete patient data.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Golang](https://golang.org/) - The Go programming language.
- [PostgreSQL](https://www.postgresql.org/) - An open-source relational database system.
- [Docker](https://www.docker.com/) - A platform for building, shipping, and running applications in containers.
- [pgAdmin](https://www.pgadmin.org/) - A comprehensive database management tool for PostgreSQL. Good to have for database management.

### Installation and Setup
1. Clone the repository to your local machine: `git clone https://github.com/Project-Sothea/Sothea-Backend.git`
 
2. In the project folder, build the project with `go build -o sothea-backend` 

3. Set up the required docker containers for the database (see below).

4. Copy `.env.example` to `.env` and fill in the required values (`PORT`, `DATABASE_URL`, `SECRET_KEY`).

5. Run the Go binary with `./sothea-backend`, starting up the server.
 
6. The server should now be accessible on `http://localhost:9090` (or whichever port you configured).

7. You can now make requests to the server using a tool like Postman or curl.
 
8. To stop the server, enter `Ctrl + C` in the terminal, then run `docker stop sothea-db` to stop the database container.

### Setting Up Docker
To facilitate easy setup of the patients database with preloaded data, we've opted to use Docker with a PostgreSQL image. To set up the database, follow the steps below:
1. Make sure the Docker daemon is running in the background.

2. Build the Docker image for the Postgres database: `docker build -t sothea-db .`

3. Run the Postgres database container with `docker run --rm --name sothea-db -d -p 5432:5432 sothea-db`

4. To stop the container, run `docker stop sothea-db`

Running with a volume:
In this project's root directory, run the Postgres database container with `docker run --name sothea-db -d -p 5432:5432 -v $(pwd)/data:/var/lib/postgresql/data sothea-db`
For Windows Powershell: `docker run --name sothea-db -d -p 5432:5432 -v ${PWD}/data:/var/lib/postgresql/data sothea-db`

### Common Issues
- Database role not found / Authentication Failed
This usually happens if there are already pre-existing Postgres instances running on port 5432. To resolve this, stop check the processes running on port 5432, and stop the existing Postgres processes.
If on Windows, do `Win` + `R`, then type `services.msc`, search for the Postgres service and stop it.

### API Endpoints
All endpoints are prefixed with `/api`. API endpoints are detailed below:

#### Login
Authenticate a user and return an access token.

```plaintext
POST /api/login
```

If successful, returns `200` and the following response attributes:

| Attribute | Type   | Description          |
|-----------|--------|----------------------|
| `token`   | string | Guaranteed to exist. |

Unsuccessful responses include:  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/login' \
--header 'Content-Type: application/json' \
--data '{
    "username": "admin",
    "password": "admin"
}'
```

#### IsValidToken
Check if the authorization token in a request is valid.

```plaintext
GET /api/login/is-valid-token
```

If successful, returns `200`.

Unsuccessful responses include:  
`401` - Unauthorized.

Example request:

```shell
curl --url 'http://localhost:9090/api/login/is-valid-token' \
--header 'Authorization: Bearer <your_access_token>'
```

#### ListUsers
List all registered users.

```plaintext
GET /api/users
```

If successful, returns `200` and an array of user objects.

Unsuccessful responses include:  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/users' \
--header 'Authorization: Bearer <your_access_token>'
```

#### GetPatientVisit
Get an existing patient's visit by their ID and Visit ID.

```plaintext
GET /api/patient/:id/:vid
```

If successful, returns `200` and the following response attributes:

| Attribute             | Type   | Description          |
|-----------------------|--------|----------------------|
| `patient_details`     | object | Guaranteed to exist. |
| `admin`               | object | Guaranteed to exist. |
| `past_medical_history`| object | May not exist.       |
| `social_history`      | object | May not exist.       |
| `vital_statistics`    | object | May not exist.       |
| `height_and_weight`   | object | May not exist.       |
| `visual_acuity`       | object | May not exist.       |
| `fall_risk`           | object | May not exist.       |
| `dental`              | object | May not exist.       |
| `physiotherapy`       | object | May not exist.       |
| `doctors_consultation`| object | May not exist.       |

Unsuccessful responses include:  
`404` - Patient not found.  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient/1/1' \
--header 'Authorization: Bearer <your_access_token>'
```

Example response:

```json
{
  "patient_details": {
    "id": 1,
    "name": "John Doe",
    "family_group": "S001",
    "khmer_name": "១២៣៤ ៥៦៧៨៩០ឥឲ",
    "dob": "1994-01-10T00:00:00Z",
    "gender": "M",
    "village": "SO",
    "contact_no": "12345678",
    "drug_allergies": "panadol"
  },
  "admin": {
    "id": 1,
    "vid": 1,
    "reg_date": "2024-01-10T00:00:00Z",
    "queue_no": "1A",
    "pregnant": false,
    "last_menstrual_period": null,
    "sent_to_id": false
  },
  "past_medical_history": { "id": 1, "vid": 1, "tuberculosis": true, "diabetes": false, "..." : "..." },
  "social_history": null,
  "vital_statistics": null,
  "height_and_weight": null,
  "visual_acuity": null,
  "fall_risk": null,
  "dental": null,
  "physiotherapy": null,
  "doctors_consultation": null
}
```

#### GetPatientPhoto
Get the photo for a patient.

```plaintext
GET /api/patient/:id/photo
```

If successful, returns `200` with the raw image bytes and the detected MIME type (`image/jpeg`, `image/png`, etc.).

Unsuccessful responses include:  
`404` - Photo not found.  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient/1/photo' \
--header 'Authorization: Bearer <your_access_token>' \
--output photo.jpg
```

#### CreatePatient
Create a new patient (demographics only, no visit).

```plaintext
POST /api/patient
```

Accepts either `application/json` or `multipart/form-data` (with `patient_details` JSON field and an optional `photo` file).

If successful, returns `200` and the following response attributes:

| Attribute | Type    | Description                       |
|-----------|---------|-----------------------------------|
| `id`      | integer | Integer id of new patient created |

Unsuccessful responses include:  
`400` - Missing or invalid patient data  
`401` - Unauthorized.  
`413` - Photo too large (max 5 MiB).  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient' \
--header 'Authorization: Bearer <your_access_token>'\
--header 'Content-Type: application/json' \
--data '{
    "name": "John Doe",
    "family_group": "S001",
    "khmer_name": "តតតតតតត",
    "dob": "1994-01-10T00:00:00Z",
    "gender": "M",
    "village": "SO",
    "contact_no": "12345678",
    "drug_allergies": "panadol"
}'
```

Example response:
```json
{
 "id": 7
}
```

#### CreatePatientWithVisit
Create a new patient and their first visit atomically.

```plaintext
POST /api/patient-with-visit
```

Accepts either `application/json` (with `patient_details` and `admin` top-level keys) or `multipart/form-data`
(with `patient_details` JSON field, `admin` JSON field, and an optional `photo` file).

If successful, returns `200` and the following response attributes:

| Attribute | Type    | Description                             |
|-----------|---------|-----------------------------------------|
| `id`      | integer | Integer id of the new patient created   |
| `vid`     | integer | Integer visit id of the new visit       |

Unsuccessful responses include:  
`400` - Missing or invalid patient/admin data  
`401` - Unauthorized.  
`413` - Photo too large (max 5 MiB).  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient-with-visit' \
--header 'Authorization: Bearer <your_access_token>'\
--header 'Content-Type: application/json' \
--data '{
    "patient_details": {
        "name": "John Doe",
        "family_group": "S001",
        "khmer_name": "តតតតតតត",
        "dob": "1994-01-10T00:00:00Z",
        "gender": "M",
        "village": "SO",
        "contact_no": "12345678",
        "drug_allergies": "panadol"
    },
    "admin": {
        "reg_date": "2024-01-10T00:00:00Z",
        "queue_no": "1A",
        "pregnant": false,
        "last_menstrual_period": null,
        "sent_to_id": false
    }
}'
```

Example response:
```json
{
 "id": 7,
 "vid": 1
}
```

#### UpdatePatient
Update demographic data for an existing patient (not visit-specific fields).

```plaintext
PUT /api/patient/:id
```

Accepts either `application/json` or `multipart/form-data` (with `patient_details` JSON field and an optional `photo` file).

If successful, returns `200`.

Unsuccessful responses include:  
`404` - Patient not found.  
`400` - Missing or invalid patient data  
`401` - Unauthorized.  
`413` - Photo too large (max 5 MiB).  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient/1' \
--request PUT \
--header 'Authorization: Bearer <your_access_token>'\
--header 'Content-Type: application/json' \
--data '{
    "name": "John Doe Updated",
    "family_group": "S001",
    "khmer_name": "តតតតតតត",
    "dob": "1994-01-10T00:00:00Z",
    "gender": "M",
    "village": "SO",
    "contact_no": "12345678",
    "drug_allergies": "panadol"
}'
```

#### DeletePatient
Delete a patient and all of their associated visits and data.

```plaintext
DELETE /api/patient/:id
```

If successful, returns `200`.

Unsuccessful responses include:  
`404` - Patient not found.  
`400` - Bad Request URL  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient/1' \
--request DELETE \
--header 'Authorization: Bearer <your_access_token>'
```

#### CreatePatientVisit
Create a new visit for an existing patient.

```plaintext
POST /api/patient/:id
```

If successful, returns `200` and the following
response attributes:

| Attribute | Type    | Description                           |
|-----------|---------|---------------------------------------|
| `vid`     | integer | Integer visit id of new visit created |

Unsuccessful responses include:
`404` - Patient not found.  
`400` - Json Marshalling Error (Attempts to marshal the JSON request body into a struct failed)  
`400` - Invalid Parameters (e.g. A required field is not present)  
`400` - Empty Request Body  
`400` - Bad Request URL  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/api/patient/1' \
--header 'Authorization: Bearer <your_access_token>'\
--header 'Content-Type: application/json' \
--data '{
    "reg_date": "2024-01-10T00:00:00Z",
    "queue_no": "1A",
    "pregnant": false,
    "last_menstrual_period": null,
    "sent_to_id": false
}'
```

Example response:
```json
{
 "vid": 5
}
```

#### DeletePatientVisit
Deletes a specified visit of an existing patient.  
To avoid accidentally deleting entire patients, only deleting visits one at a time is allowed.

```plaintext
DELETE /api/patient/:id/:vid
```

If successful, returns `200`

Unsuccessful responses include:  
`404` - Patient Visit not found.  
`400` - Bad Request URL  
`401` - Unauthorized.  
`500` - Internal server error.  

Example request:

```shell
curl --url --request DELETE 'http://localhost:9090/api/patient/1/1' \
--header 'Authorization: Bearer <your_access_token>'
```

#### UpdatePatientVisit
Update a visit of an existing patient. Only fields included in the request body are updated.

```plaintext
PATCH /api/patient/:id/:vid
```

If successful, returns `200`

Unsuccessful responses include:
`404` - Patient visit not found.  
`400` - Empty Request Body
`400` - Json Marshalling Error (Attempts to marshal the JSON request body into a struct failed)
`400` - Invalid Parameters (e.g. A required field is not present)
`400` - Bad Request URL
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --location --request PATCH 'http://localhost:9090/api/patient/1/1' \
--header 'Authorization: Bearer <your_access_token>' \
--header 'Content-Type: application/json' \
--data '{
    "admin": {
        "reg_date": "2024-01-10T00:00:00Z",
        "queue_no": "3B",
        "pregnant": false,
        "last_menstrual_period": null,
        "sent_to_id": false
    },
    "past_medical_history": {
        "tuberculosis": true,
        "diabetes": false,
        "hypertension": true,
        "hyperlipidemia": false,
        "chronic_joint_pains": false,
        "chronic_muscle_aches": true,
        "sexually_transmitted_disease": true,
        "specified_stds": "TRICHOMONAS",
        "others": "None"
    },
    "doctors_consultation": {
        "well": true,
        "msk": false,
        "consultation_notes": "CHEST PAIN",
        "diagnosis": "ACUTE BRONCHITIS",
        "treatment": "REST, HYDRATION, COUGH SYRUP",
        "referral_needed": false,
        "referral_loc": null,
        "remarks": "MONITOR FOR RESOLUTION"
    }
}'
```

#### GetPatientMeta
Retrieve metadata for a specific patient, allowing further requests to be made to retrieve individual patient visit data.

```plaintext
GET /api/patient-meta/:id
```

If successful, returns `200`

| Attribute      | Type    | Description                              |
|----------------|---------|------------------------------------------|
| `id`           | integer | Integer id of patient                    |
| `vid`          | integer | Integer visit id of the latest visit     |
| `family_group` | string  | Family group identifier                  |
| `reg_date`     | string  | Registration date of the latest visit    |
| `queue_no`     | string  | Queue number of the latest visit         |
| `name`         | string  | Name of patient                          |
| `khmer_name`   | string  | Khmer name of patient                    |
| `visits`       | object  | Mapping of visit ids to registration dates |

Unsuccessful responses include:
`404` - Patient not found.  
`400` - Bad Request URL
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --location 'http://localhost:9090/api/patient-meta/1' \
--header 'Authorization: Bearer <your_access_token>'
```

Example response:
```json
{
    "id": 1,
    "vid": 1,
    "family_group": "S001",
    "reg_date": "2024-01-10T00:00:00Z",
    "queue_no": "1A",
    "name": "John Doe",
    "khmer_name": "១២៣៤ ៥៦៧៨៩០ឥឲ",
    "visits": {
        "1": "2024-01-10T00:00:00Z",
        "2": "2023-07-01T00:00:00Z",
        "3": "2023-07-02T00:00:00Z",
        "4": "2023-07-03T00:00:00Z"
    }
}
```

#### GetAllPatientVisitMeta
Retrieve and return patient visit metadata for all patients on a specific date.
Pass `default` as the date to get the latest visit for each patient.

```plaintext
GET /api/all-patient-visit-meta/:date
```

The `:date` parameter accepts either `default` (latest visits) or a date in `YYYY-MM-DD` format.

If successful, returns `200`, and an array of patient visit metadata objects.

Unsuccessful responses include:
`400` - Bad Request URL
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --location 'http://localhost:9090/api/all-patient-visit-meta/default' \
--header 'Authorization: Bearer <your_access_token>'
```

Example response:
```json
[
 {
  "id": 1,
  "vid": 2,
  "family_group": "Family 1",
  "reg_date": "2025-07-01T00:00:00Z",
  "queue_no": "Q123",
  "name": "John Doe",
  "khmer_name": "ខេមរ",
  "gender": "M",
  "village": "Village 1",
  "contact_no": "123456789",
  "drug_allergies": "None",
  "sent_to_id": false,
  "referral_needed": false,
  "has_prescription_with_drug": false,
  "all_prescription_drugs_packed": false,
  "prescription_dispensed": false
 },
 {
  "id": 2,
  "vid": 2,
  "family_group": "B009",
  "reg_date": "2024-12-03T00:00:00Z",
  "queue_no": "Q125",
  "name": "Walter White",
  "khmer_name": "អាលីស ស្ម៊ីត",
  "gender": "M",
  "village": "ABQ",
  "contact_no": "555666777",
  "drug_allergies": "None",
  "sent_to_id": false,
  "referral_needed": false,
  "has_prescription_with_drug": false,
  "all_prescription_drugs_packed": false,
  "prescription_dispensed": false
 }
]
```

#### Pharmacy Endpoints
Manage drug inventory, batches, and batch locations. All pharmacy routes require authentication.

```plaintext
GET    /api/pharmacy/drugs               - List drugs (optional ?q=<search> query param)
POST   /api/pharmacy/drugs               - Create a drug
GET    /api/pharmacy/drugs/:drugId       - Get a drug with its total stock
PATCH  /api/pharmacy/drugs/:drugId       - Update a drug
DELETE /api/pharmacy/drugs/:drugId       - Delete a drug

GET    /api/pharmacy/drugs/:drugId/batches       - List batches for a drug
POST   /api/pharmacy/drugs/:drugId/batches       - Create a batch (with optional locations)

GET    /api/pharmacy/batches             - List all batches across all drugs
GET    /api/pharmacy/batches/:batchId    - Get a specific batch
PATCH  /api/pharmacy/batches/:batchId    - Update a batch
DELETE /api/pharmacy/batches/:batchId    - Delete a batch

GET    /api/pharmacy/batches/:batchId/locations  - List locations for a batch
POST   /api/pharmacy/batches/:batchId/locations  - Add a location to a batch

PATCH  /api/pharmacy/locations/:locationId       - Update a batch location
DELETE /api/pharmacy/locations/:locationId       - Delete a batch location
```

#### Prescription Endpoints
Manage the full prescription lifecycle. All prescription routes require authentication.

```plaintext
GET    /api/prescriptions                          - List prescriptions (optional ?patient_id= and ?vid= filters)
POST   /api/prescriptions                          - Create a prescription
GET    /api/prescriptions/:id                      - Get a prescription with all its lines
PATCH  /api/prescriptions/:id                      - Update a prescription header
DELETE /api/prescriptions/:id                      - Delete a prescription

POST   /api/prescriptions/:id/lines                - Add a line (drug) to a prescription
PATCH  /api/prescriptions/lines/:lineId            - Update a prescription line
DELETE /api/prescriptions/lines/:lineId            - Remove a line

GET    /api/prescriptions/lines/:lineId/allocations - List batch allocations for a line
PUT    /api/prescriptions/lines/:lineId/allocations - Replace all allocations for a line

POST   /api/prescriptions/lines/:lineId/pack       - Mark a line as packed
POST   /api/prescriptions/lines/:lineId/unpack     - Unpack a line

POST   /api/prescriptions/:id/dispense             - Dispense a prescription (finalise)
```
