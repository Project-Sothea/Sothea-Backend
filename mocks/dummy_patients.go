package mocks

import (
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// Valid Patient JSON
var ValidPatientJson = `{
  "admin": {
    "familyGroup": "S001",
    "regDate": "2024-01-10T00:00:00Z",
	"queueNo": "8A",
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
    "others": null
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
}`

// Only the admin section of ValidPatientJson
var ValidPatientAdminJson = `{
	"familyGroup": "S001",
	"regDate": "2024-01-10T00:00:00Z",
	"queueNo": "8A",
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
}`

var admin = entities.Admin{
	FamilyGroup:         entities.PtrTo("S001"),
	RegDate:             entities.PtrTo(time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC)),
	QueueNo:             entities.PtrTo("8A"),
	Name:                entities.PtrTo("Patient's Name Here"),
	KhmerName:           entities.PtrTo("តតតតតតត"),
	Dob:                 entities.PtrTo(time.Date(1994, time.January, 10, 0, 0, 0, 0, time.UTC)),
	Age:                 entities.PtrTo(30),
	Gender:              entities.PtrTo("M"),
	Village:             entities.PtrTo("SO"),
	ContactNo:           entities.PtrTo("12345678"),
	Pregnant:            entities.PtrTo(false),
	LastMenstrualPeriod: nil,
	DrugAllergies:       entities.PtrTo("panadol"),
	SentToID:            entities.PtrTo(false),
}
var pastmedicalhistory = entities.PastMedicalHistory{
	Tuberculosis:               entities.PtrTo(true),
	Diabetes:                   entities.PtrTo(false),
	Hypertension:               entities.PtrTo(true),
	Hyperlipidemia:             entities.PtrTo(false),
	ChronicJointPains:          entities.PtrTo(false),
	ChronicMuscleAches:         entities.PtrTo(true),
	SexuallyTransmittedDisease: entities.PtrTo(true),
	SpecifiedSTDs:              entities.PtrTo("TRICHOMONAS"),
	Others:                     nil,
}
var socialhistory = entities.SocialHistory{
	PastSmokingHistory:    entities.PtrTo(true),
	NumberOfYears:         entities.PtrTo(int32(15)),
	CurrentSmokingHistory: entities.PtrTo(false),
	CigarettesPerDay:      nil,
	AlcoholHistory:        entities.PtrTo(true),
	HowRegular:            entities.PtrTo("A"),
}
var vitalstatistics = entities.VitalStatistics{
	Temperature:             entities.PtrTo(36.5),
	SpO2:                    entities.PtrTo(98.0),
	SystolicBP1:             entities.PtrTo(120.0),
	DiastolicBP1:            entities.PtrTo(80.0),
	SystolicBP2:             entities.PtrTo(122.0),
	DiastolicBP2:            entities.PtrTo(78.0),
	AverageSystolicBP:       entities.PtrTo(121.0),
	AverageDiastolicBP:      entities.PtrTo(79.0),
	HR1:                     entities.PtrTo(72.0),
	HR2:                     entities.PtrTo(71.0),
	AverageHR:               entities.PtrTo(71.5),
	RandomBloodGlucoseMmolL: entities.PtrTo(5.4),
}
var heightandweight = entities.HeightAndWeight{
	Height:      entities.PtrTo(170.0),
	Weight:      entities.PtrTo(70.0),
	BMI:         entities.PtrTo(24.2),
	BMIAnalysis: entities.PtrTo("normal weight"),
	PaedsHeight: entities.PtrTo(90.0),
	PaedsWeight: entities.PtrTo(80.0),
}
var visualacuity = entities.VisualAcuity{
	LEyeVision:             entities.PtrTo(int32(20)),
	REyeVision:             entities.PtrTo(int32(20)),
	AdditionalIntervention: entities.PtrTo("VISUAL FIELD TEST REQUIRED"),
}
var doctorsconsultation = entities.DoctorsConsultation{
	Well:              entities.PtrTo(true),
	Msk:               entities.PtrTo(false),
	Cvs:               entities.PtrTo(false),
	Respi:             entities.PtrTo(true),
	Gu:                entities.PtrTo(true),
	Git:               entities.PtrTo(false),
	Eye:               entities.PtrTo(true),
	Derm:              entities.PtrTo(false),
	Others:            entities.PtrTo("TRICHOMONAS VAGINALIS"),
	ConsultationNotes: entities.PtrTo("CHEST PAIN, SHORTNESS OF BREATH, COUGH"),
	Diagnosis:         entities.PtrTo("ACUTE BRONCHITIS"),
	Treatment:         entities.PtrTo("REST, HYDRATION, COUGH SYRUP"),
	ReferralNeeded:    entities.PtrTo(false),
	ReferralLoc:       nil,
	Remarks:           entities.PtrTo("MONITOR FOR RESOLUTION"),
}
var ValidPatient = entities.Patient{
	Admin:               &admin,
	PastMedicalHistory:  &pastmedicalhistory,
	SocialHistory:       &socialhistory,
	VitalStatistics:     &vitalstatistics,
	HeightAndWeight:     &heightandweight,
	VisualAcuity:        &visualacuity,
	DoctorsConsultation: &doctorsconsultation,
}

// Missing Admin Field
var MissingAdminPatientJson = `{
  "pastMedicalHistory": {
    "tuberculosis": true,
    "diabetes": false,
    "hypertension": true,
    "hyperlipidemia": false,
    "chronicJointPains": false,
    "chronicMuscleAches": true,
    "sexuallyTransmittedDisease": true,
    "specifiedSTDs": "TRICHOMONAS",
    "others": null
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
}`

var MissingAdminPatient = entities.Patient{
	PastMedicalHistory:  &pastmedicalhistory,
	SocialHistory:       &socialhistory,
	VitalStatistics:     &vitalstatistics,
	HeightAndWeight:     &heightandweight,
	VisualAcuity:        &visualacuity,
	DoctorsConsultation: &doctorsconsultation,
}

// Invalid Parameters
var InvalidParametersPatientJson = `{
  "admin": {
    "regDate": "2024-01-10T00:00:00Z",
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
  },
  "pastMedicalHistory": {
    "tuberculosis": "invalid data type here",
    "diabetes": false,
    "hypertension": true,
    "hyperlipidemia": false,
    "chronicJointPains": false,
    "chronicMuscleAches": true,
    "sexuallyTransmittedDisease": true,
    "specifiedSTDs": "TRICHOMONAS",
    "others": null
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
}`

// Invalid Parameters Admin Json
var InvalidParametersAdminJson = `{
    "regDate": "2024-01-10T00:00:00Z",
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
}`

// JSON Marshalling Error
var JSONMarshallingErrorPatientJson = `{
  "admin": {
    "familyGroup": false,
    "regDate": "2024-01-10T00:00:00Z",
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
    "others": null
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
"fallRisk": {
	"fallWorries": a,
	"fallHistory": a,
	"cognitiveStatus": b,
	"continenceProblems": c,
	"safetyAwareness": d,
	"unsteadiness": b,
	"fallRiskScore": 6
}
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
}`

// JSON Marshalling Error Admin Json
var JSONMarshallingErrorAdminJson = `{
    "familyGroup": false,
    "regDate": "2024-01-10T00:00:00Z",
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
}`

var ValidPatientMeta = entities.PatientMeta{
	ID:          1,
	VID:         1,
	FamilyGroup: entities.PtrTo("Family A"),
	RegDate:     entities.PtrTo(time.Now()), // Use current time or a specific date
	QueueNo:     entities.PtrTo("Q123"),
	Name:        entities.PtrTo("John Doe"),
	KhmerName:   entities.PtrTo("ជោគជ័យ"),
	Visits: map[int32]time.Time{
		1: time.Now().Add(-24 * time.Hour), // Example: a visit one day ago
		2: time.Now().Add(-48 * time.Hour), // Example: a visit two days ago
	},
}

var ValidPatientVisitMetaArray = []entities.PatientVisitMeta{
	{
		ID:                         1,
		VID:                        1,
		FamilyGroup:                entities.PtrTo("Family A"),
		RegDate:                    entities.PtrTo(time.Now().Add(-24 * time.Hour)), // Example date
		QueueNo:                    entities.PtrTo("Q123"),
		Name:                       entities.PtrTo("John Doe"),
		KhmerName:                  entities.PtrTo("ជោគជ័យ"),
		Gender:                     entities.PtrTo("M"),
		Village:                    entities.PtrTo("Village A"),
		ContactNo:                  entities.PtrTo("12345678"),
		DrugAllergies:              entities.PtrTo("Penicillin"),
		SentToID:                   entities.PtrTo(false),
		ReferralNeeded:             entities.PtrTo(false),
		HasPrescriptionWithDrug:    entities.PtrTo(true),
		AllPrescriptionDrugsPacked: entities.PtrTo(false),
		PrescriptionDispensed:      entities.PtrTo(false),
	},
	{
		ID:                         2,
		VID:                        2,
		FamilyGroup:                entities.PtrTo("Family B"),
		RegDate:                    entities.PtrTo(time.Now().Add(-48 * time.Hour)), // Example date
		QueueNo:                    entities.PtrTo("Q124"),
		Name:                       entities.PtrTo("Jane Smith"),
		KhmerName:                  entities.PtrTo("ស្រីសាម"),
		Gender:                     entities.PtrTo("F"),
		Village:                    entities.PtrTo("Village B"),
		ContactNo:                  entities.PtrTo("87654321"),
		DrugAllergies:              entities.PtrTo("None"),
		SentToID:                   entities.PtrTo(true),
		ReferralNeeded:             entities.PtrTo(true),
		HasPrescriptionWithDrug:    entities.PtrTo(true),
		AllPrescriptionDrugsPacked: entities.PtrTo(true),
		PrescriptionDispensed:      entities.PtrTo(true),
	},
}
