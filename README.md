# Project Sothea Backend
### Last Updated: 4 Nov, 2024
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

4. Run the Go binary with `./sothea-backend --mode=dev`, starting it up in development mode.
 
5. The server should now be accessible on `http://localhost:9090`

6. You can now make requests to the server using a tool like Postman or curl.
 
7. To stop the server, enter `Ctrl + C` in the terminal, then run `docker stop sothea-db` to stop the database container.

### Setting Up Docker
To facilitate easy setup of the patients database with preloaded data, we've opted to use Docker with a PostgreSQL image. To set up the database, follow the steps below:
1. Make sure the Docker daemon is running in the background.

2. Build the Docker image for the Postgres database: `docker build -t sothea-db .`

3. Run the Postgres database container with `docker run --rm --name sothea-db -d -p 5432:5432 sothea-db`

4. To stop the container, run `docker stop sothea-db`

Running with a volume:
In this project's root directory, run the Postgres database container with `docker run --name sothea-db -d -p 5432:5432 -v $(pwd)/data:/var/lib/postgresql/data sothea-db`
For Windows Powershell: `docker run --name sothea-db -d -p 5432:5432 -v ${PWD}/data:/var/lib/postgresql/data sothea-db`

### Modes of Operation
When running the Go binary, you can specify the mode of operation using the `--mode` flag. The available modes are:  
- `dev` - Development mode, using config.json for configuration. This mode will run the server on port 9090.
- `prod` - Production mode, using prod.json for configuration. This mode will run the server on "192.168.0.100:9090", a static IP address that we use on our production server on the deployed network.   
Do ensure that the frontend is appropriately configured to make requests to the correct backend address.

### Common Issues
- Database role not found / Authentication Failed
This usually happens if there are already pre-existing Postgres instances running on port 5432. To resolve this, stop check the processes running on port 5432, and stop the existing Postgres processes.
If on Windows, do `Win` + `R`, then type `services.msc`, search for the Postgres service and stop it.

### API Endpoints
API endpoints are detailed below:

#### Login
Authenticate a user and return an access token.

```plaintext
POST /login
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
curl --url 'http://localhost:9090/login' \
--header 'Content-Type: application/json' \
--data '{
    "username": "admin",
    "password": "admin"
}'
```

#### GetPatientVisit
Get an existing patient's visit by their ID and Visit ID.

```plaintext
GET /patient/:id/:vid
```

If successful, returns `200` and the following response attributes:

| Attribute             | Type   | Description          |
|-----------------------|--------|----------------------|
| `admin`               | object | Guaranteed to exist. |
| `pastmedicalhistory`  | object | May not exist.       |
| `socialhistory`       | object | May not exist.       |
| `vitalstatistics`     | object | May not exist.       |
| `heightandweight`     | object | May not exist.       |
| `visualacuity`        | object | May not exist.       |
| `fallrisk`            | object | May not exist.       |
| `dental`              | object | May not exist.       |
| `doctorsconsultation` | object | May not exist.       |

Unsuccessful responses include:  
`404` - Patient not found.  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/patient/1/1' \
--header 'Authorization: Bearer <your_access_token>'
```

Example response:

```json
{
  "admin": {
    "id": 1,
    "vid": 1,
    "familyGroup": "S001",
    "regDate": "2024-01-10T00:00:00Z",
    "queueNo": "1A",
    "name": "John Doe",
    "khmerName": "១២៣៤ ៥៦៧៨៩០ឥឲ",
    "dob": "1994-01-10T00:00:00Z",
    "age": 30,
    "gender": "M",
    "village": "SO",
    "contactNo": "12345678",
    "pregnant": false,
    "lastMenstrualPeriod": null,
    "drugAllergies": "panadol",
    "sentToId": false,
    "photo": ""
  },
  "pastmedicalhistory": {
    "id": 1,
    "vid": 1,
    "tuberculosis": true,
    "diabetes": false,
    "hypertension": true,
    "hyperlipidemia": false,
    "chronicJointPains": false,
    "chronicMuscleAches": true,
    "sexuallyTransmittedDisease": true,
    "specifiedSTDs": "TRICHOMONAS",
    "others": "None"
  },
  "socialhistory": {
    "id": 1,
    "vid": 1,
    "pastSmokingHistory": true,
    "numberOfYears": 15,
    "currentSmokingHistory": false,
    "cigarettesPerDay": null,
    "alcoholHistory": true,
    "howRegular": "A"
  },
  "vitalstatistics": {
    "id": 1,
    "vid": 1,
    "temperature": 36.5,
    "spO2": 98,
    "systolicBP1": 120,
    "diastolicBP1": 80,
    "systolicBP2": 122,
    "diastolicBP2": 78,
    "averageSystolicBP": 121,
    "averageDiastolicBP": 79,
    "hr1": 72,
    "hr2": 71,
    "averageHR": 71.5,
    "randomBloodGlucoseMmolL": 5.4
  },
  "heightandweight": {
    "id": 1,
    "vid": 1,
    "height": 170,
    "weight": 70,
    "bmi": 24.2,
    "bmiAnalysis": "normal weight",
    "paedsHeight": 90,
    "paedsWeight": 80
  },
  "visualacuity": {
    "id": 1,
    "vid": 1,
    "lEyeVision": 20,
    "rEyeVision": 20,
    "additionalIntervention": "VISUAL FIELD TEST REQUIRED"
  },
  "fallrisk": {
    "id": 1,
    "vid": 1,
    "fallHistory": "a",
    "cognitiveStatus": "b",
    "continenceProblems": "e",
    "safetyAwareness": "d",
    "unsteadiness": "c"
  },
  "dental": {
    "id": 1,
    "vid": 1,
    "cleanTeethFreq": 2,
    "sugarConsumeFreq": 3,
    "pastYearDecay": true,
    "brushTeethPain": true,
    "drinkOtherWater": false,
    "dentalNotes": "None",
    "referralNeeded": true,
    "referralLoc": "Dentist",
    "tooth11": true,
    "tooth12": false,
    "tooth13": true,
    "tooth14": false,
    "tooth15": true,
    "tooth16": false,
    "tooth17": true,
    "tooth18": false,
    "tooth21": true,
    "tooth22": false,
    "tooth23": true,
    "tooth24": false,
    "tooth25": true,
    "tooth26": false,
    "tooth27": true,
    "tooth28": false,
    "tooth31": true,
    "tooth32": true,
    "tooth33": false,
    "tooth34": true,
    "tooth35": false,
    "tooth36": true,
    "tooth37": false,
    "tooth38": true,
    "tooth41": false,
    "tooth42": true,
    "tooth43": false,
    "tooth44": true,
    "tooth45": false,
    "tooth46": true,
    "tooth47": false,
    "tooth48": true
  },
  "doctorsconsultation": {
    "id": 1,
    "vid": 1,
    "well": true,
    "msk": false,
    "cvs": false,
    "respi": true,
    "gu": true,
    "git": false,
    "eye": true,
    "derm": false,
    "others": "LEUKAEMIA",
    "consultationNotes": "CHEST PAIN, SHORTNESS OF BREATH, COUGH",
    "diagnosis": "ACUTE BRONCHITIS",
    "treatment": "REST, HYDRATION, COUGH SYRUP",
    "referralNeeded": false,
    "referralLoc": null,
    "remarks": "MONITOR FOR RESOLUTION"
  }
}
```

#### CreatePatient
Create a new patient.

```plaintext
POST /patient
```

If successful, returns `200` and the following
response attributes:

| Attribute | Type    | Description                       |
|-----------|---------|-----------------------------------|
| `id`      | integer | Integer id of new patient created |

Unsuccessful responses include:  
`400` - Missing Admin Category  
`400` - Json Marshalling Error (Attempts to marshal the JSON request body into a struct failed)  
`400` - Invalid Parameters (e.g. A required field is not present)  
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --url 'http://localhost:9090/patient/' \
--header 'Authorization: Bearer <your_access_token>'\
--header 'Content-Type: application/json' \
--data '
{
    "familyGroup": "S001",
    "regDate": "2024-01-10T00:00:00Z",
    "queueNo": "1A",
    "name": "Patient's Name Here",
    "khmerName": "តតតតតតត",
    "dob": "1994-01-10T00:00:00Z",
    "age": 30,
    "gender": "M",
    "village": "SO",
    "contactNo": "12345678",
    "pregnant": false,
    "lastMenstrualPeriod": null,
    "drugAllergies": "panadol",
    "sentToID": false,
    "photo": "<photo_encoded_as_base64_string>"
}'
```

Example response:
```json
{
 "id": 7
}
```

#### CreatePatientVisit
Create a new visit for an existing patient.

```plaintext
POST /patient/:id
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
curl --url 'http://localhost:9090/patient/1' \
--header 'Authorization: Bearer <your_access_token>'\
--header 'Content-Type: application/json' \
--data '{
    "familyGroup": "S001",
    "regDate": "2024-01-10T00:00:00Z",
    "queueNo": "1A",
    "name": "Patient's Name Here",
    "khmerName": "តតតតតតត",
    "dob": "1994-01-10T00:00:00Z",
    "age": 30,
    "gender": "M",
    "village": "SO",
    "contactNo": "12345678",
    "pregnant": false,
    "lastMenstrualPeriod": null,
    "drugAllergies": "panadol",
    "sentToID": false,
    "photo": "<photo_encoded_as_base64_string>"
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
DELETE /patient/:id/:vid
```

If successful, returns `200`

Unsuccessful responses include:  
`404` - Patient Visit not found.  
`400` - Bad Request URL  
`401` - Unauthorized.  
`500` - Internal server error.  

Example request:

```shell
curl --url --request DELETE 'http://localhost:9090/patient/1/1' \
--header 'Authorization: Bearer <your_access_token>'
```

#### UpdatePatientVisit
Update a visit of an existing patient.

```plaintext
PATCH /patient/:id/:vid
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
curl --location --request PATCH 'http://localhost:9090/patient/1/1' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjI4MTAzNzIsInVzZXJuYW1lIjoiYWRtaW4ifQ.ap0GiJ3fnlxHFYlRDQ2Bk21KVhZyTTH4500tZoWA4rc' \
--header 'Content-Type: application/json' \
--data '{
    "admin": {
        "familyGroup": "S001",
        "regDate": "2024-01-10T00:00:00Z",
        "queueNo": "3B",
        "name": "Patient'\''s Name Here",
        "khmerName": "តតតតតតត",
        "dob": "1994-01-10T00:00:00Z",
        "age": 30,
        "gender": "M",
        "village": "SO",
        "contactNo": "12345678",
        "pregnant": false,
        "lastMenstrualPeriod": null,
        "drugAllergies": "panadol",
        "sentToID": false,
        "photo": "<photo_encoded_as_base64_string>"
    },
    "pastMedicalHistory": {
        "tuberculosis": true,
        "diabetes": false,
        "hypertension": true,
        "hyperlipidemia": false,
        "chronicJointPains": false,
        "chronicMuscleAches": true,
        "sexuallyTransmittedDisease": true,
        "specifiedSTDs": "TRICHOMONAS",
        "others": "None"
    },
    "socialHistory": {
        "pastSmokingHistory": true,
        "numberOfYears": 15,
        "currentSmokingHistory": false,
        "cigarettesPerDay": null,
        "alcoholHistory": true,
        "howRegular": "A"
    },
    "vitalStatistics": {
        "temperature": 36.5,
        "spO2": 98,
        "systolicBP1": 120,
        "diastolicBP1": 80,
        "systolicBP2": 122,
        "diastolicBP2": 78,
        "averageSystolicBP": 121,
        "averageDiastolicBP": 79,
        "hr1": 72,
        "hr2": 71,
        "averageHR": 71.5,
        "randomBloodGlucoseMmolL": 5.4
    },
    "heightAndWeight": {
        "height": 170,
        "weight": 70,
        "bmi": 24.2,
        "bmiAnalysis": "normal weight",
        "paedsHeight": 90,
        "paedsWeight": 80
    },
    "visualAcuity": {
        "lEyeVision": 20,
        "rEyeVision": 20,
        "additionalIntervention": "VISUAL FIELD TEST REQUIRED"
    },
    "fallrisk": {
        "fallHistory": "c",
        "cognitiveStatus": "b",
        "continenceProblems": "e",
        "safetyAwareness": "d",
        "unsteadiness": "c"
    },
    "doctorsConsultation": {
        "well": true,
        "msk": false,
        "cvs": false,
        "respi": true,
        "gu": true,
        "git": false,
        "eye": true,
        "derm": false,
        "others": "TRICHOMONAS VAGINALIS",
        "consultationNotes": "CHEST PAIN, SHORTNESS OF BREATH, COUGH",
        "diagnosis": "ACUTE BRONCHITIS",
        "treatment": "REST, HYDRATION, COUGH SYRUP",
        "referralNeeded": false,
        "referralLoc": null,
        "remarks": "MONITOR FOR RESOLUTION"
    }
}'
```

#### GetPatientMeta
Retrieve metadata for a specific patient, allowing further requests to be made to retrieve individual patient visit data.

```plaintext
GET /patient-meta/:id
```

If successful, returns `200`

| Attribute     | Type    | Description                           |
|---------------|---------|---------------------------------------|
| `id`          | integer | Integer id of patient                 |
| `vid`         | integer | Integer visit id                      |
| `familyGroup` | string  | Integer visit id of new visit created |
| `regDate`     | string  | Registration date of visit            |
| `queueNo`     | string  | Queue number given for visit          |
| `name`        | string  | Name of patient                       |
| `khmerName`   | string  | Khmer name of patient                 |
| `visits`      | object  | Mapping of visit ids to regDate       |

Unsuccessful responses include:
`404` - Patient visit not found.  
`400` - Bad Request URL
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --location 'http://localhost:9090/patient-meta/1' \
--header 'Authorization: Bearer <your_access_token>'
```

Example response:
```json
{
    "id": 1,
    "vid": 1,
    "familyGroup": "S001",
    "regDate": "2024-01-10T00:00:00Z",
    "queueNo": "1A",
    "name": "John Doe",
    "khmerName": "១២៣៤ ៥៦៧៨៩០ឥឲ",
    "visits": {
        "1": "2024-01-10T00:00:00Z",
        "2": "2023-07-01T00:00:00Z",
        "3": "2023-07-02T00:00:00Z",
        "4": "2023-07-03T00:00:00Z"
    }
}
```

#### GetAllPatientVisitMeta
Retrieve and return patient visit metadata for all patients on a specific date

```plaintext
GET /all-patient-visit-meta/:date
```

If successful, returns `200`, and an array of patient visit metadata objects.

Unsuccessful responses include:
`400` - Bad Request URL
`401` - Unauthorized.  
`500` - Internal server error.

Example request:

```shell
curl --location 'http://localhost:9090/all-patient-visit-meta/default' \
--header 'Authorization: Bearer <your_access_token>'
```

Example response:
```json
[
 {
  "id": 1,
  "vid": 2,
  "familyGroup": "Family 1",
  "regDate": "2025-07-01T00:00:00Z",
  "queueNo": "Q123",
  "name": "John Doe",
  "khmerName": "ខេមរ",
  "gender": "M",
  "village": "Village 1",
  "contactNo": "123456789",
  "drugAllergies": "None",
  "sentToId": false,
  "referralNeeded": false
 },
 {
  "id": 2,
  "vid": 2,
  "familyGroup": "B009",
  "regDate": "2024-12-03T00:00:00Z",
  "queueNo": "Q125",
  "name": "Walter White",
  "khmerName": "អាលីស ស្ម៊ីត",
  "gender": "M",
  "village": "ABQ",
  "contactNo": "555666777",
  "drugAllergies": "None",
  "sentToId": false,
  "referralNeeded": false
 },
 {
  "id": 3,
  "vid": 1,
  "familyGroup": "S002B",
  "regDate": "2024-01-10T00:00:00Z",
  "queueNo": "2B",
  "name": "Bob Smith",
  "khmerName": "១២៣៤ ៥៦៧៨៩០ឥឲ",
  "gender": "M",
  "village": "R1",
  "contactNo": "99999999",
  "drugAllergies": "aspirin",
  "sentToId": false,
  "referralNeeded": null
 },
 {
  "id": 4,
  "vid": 1,
  "familyGroup": "S003",
  "regDate": "2024-01-10T00:00:00Z",
  "queueNo": "3A",
  "name": "Bob Johnson",
  "khmerName": "១២៣៤ ៥៦៧៨៩០ឥឲ",
  "gender": "M",
  "village": "R1",
  "contactNo": "11111111",
  "drugAllergies": null,
  "sentToId": false,
  "referralNeeded": null
 },
 {
  "id": 5,
  "vid": 1,
  "familyGroup": "S004",
  "regDate": "2024-01-10T00:00:00Z",
  "queueNo": "4B",
  "name": "Alice Brown",
  "khmerName": "១២៣៤ ៥៦៧៨៩០ឥឲ",
  "gender": "F",
  "village": "R1",
  "contactNo": "17283948",
  "drugAllergies": null,
  "sentToId": false,
  "referralNeeded": null
 },
 {
  "id": 6,
  "vid": 1,
  "familyGroup": "S005A",
  "regDate": "2024-01-10T00:00:00Z",
  "queueNo": "5C",
  "name": "Charlie Davis",
  "khmerName": "១២៣៤ ៥៦៧៨៩០ឥឲ",
  "gender": "M",
  "village": "R1",
  "contactNo": "09876543",
  "drugAllergies": null,
  "sentToId": false,
  "referralNeeded": null
 }
]
```

#### Export Patients
Export all patient data to a CSV file.

```plaintext
GET /export-db

GET /export-db?includePhoto=true
```

#### Is Valid Token
Check if the authorization token in a request is valid.

```plaintext
GET /isValidToken
```
