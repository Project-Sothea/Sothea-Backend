package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_httpDelivery "github.com/jieqiboh/sothea_backend/controllers"
	_postgresRepository "github.com/jieqiboh/sothea_backend/repository/postgres"
	_useCase "github.com/jieqiboh/sothea_backend/usecases"
	"github.com/spf13/viper"
)

func main() {
	viper.AutomaticEnv()

	port := viper.GetString("PORT")
	connStr := viper.GetString("DATABASE_URL")
	secretKey := []byte(viper.GetString("SECRET_KEY"))

	// Open a database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// You might want to check the connection here to handle errors
	err = db.Ping()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	router := gin.Default()
	router.Static("/app", "./dist")

	// Root router group for all public API endpoints
	api := router.Group("/api")

	patientRepo := _postgresRepository.NewPostgresPatientRepository(db)
	// Set up login routes
	loginUseCase := _useCase.NewLoginUseCase(patientRepo, 5*time.Second, secretKey)
	_httpDelivery.NewLoginHandler(api, loginUseCase, secretKey)

	// Set up patient routes
	patientUseCase := _useCase.NewPatientUsecase(patientRepo, 2*time.Second)
	_httpDelivery.NewPatientHandler(api, patientUseCase, secretKey)

	pharmacyRepo := _postgresRepository.NewPostgresPharmacyRepository(db)
	pharmacyUseCase := _useCase.NewPharmacyUsecase(pharmacyRepo, 2*time.Second)
	_httpDelivery.NewPharmacyHandler(api, pharmacyUseCase, secretKey, db)

	prescriptionRepo := _postgresRepository.NewPostgresPrescriptionRepository(db)
	prescriptionUseCase := _useCase.NewPrescriptionUsecase(prescriptionRepo, pharmacyRepo, 2*time.Second)
	_httpDelivery.NewPrescriptionHandler(api, prescriptionUseCase, secretKey, db)

	router.NoRoute(func(c *gin.Context) {
		// Only serve index.html for non-API requests
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./dist/index.html")
		}
	})

	router.Run("0.0.0.0:" + port)
}
