package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_httpDelivery "github.com/jieqiboh/sothea_backend/controllers"
	_patientPostgresRepository "github.com/jieqiboh/sothea_backend/repository/postgres"
	_useCase "github.com/jieqiboh/sothea_backend/usecases"
	"github.com/spf13/viper"
)

func main() {
	// Define a flag to determine the mode
	mode := flag.String("mode", "dev", "Mode of the application: dev or prod")

	// Parse the flags
	flag.Parse()

	// Determine the mode and print a message
	switch *mode {
	case "dev":
		gin.SetMode(gin.DebugMode)
		fmt.Println("Running in development mode")
		viper.SetConfigFile(`config.json`)
	case "prod":
		gin.SetMode(gin.ReleaseMode)
		fmt.Println("Running in production mode")
		viper.SetConfigFile(`prod.json`)
	default:
		fmt.Println("Unknown mode. Please use 'dev' or 'prod'.")
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	address := viper.GetString(`server.address`)
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbName := viper.GetString(`database.name`)
	dbPassword := viper.GetString(`database.password`)
	dbSslMode := viper.GetString(`database.sslmode`)
	secretKey := []byte(viper.GetString(`jwt.secretkey`))

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPassword, dbName, dbSslMode)

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
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "*"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PATCH"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	patientRepo := _patientPostgresRepository.NewPostgresPatientRepository(db)
	// Set up login routes
	loginUseCase := _useCase.NewLoginUseCase(patientRepo, 5*time.Second, secretKey)
	_httpDelivery.NewLoginHandler(router, loginUseCase, secretKey)

	// Set up patient routes
	patientUseCase := _useCase.NewPatientUsecase(patientRepo, 2*time.Second)
	_httpDelivery.NewPatientHandler(router, patientUseCase, secretKey)

	pharmacyRepo := _patientPostgresRepository.NewPostgresPharmacyRepository(db)
	pharmacyUseCase := _useCase.NewPharmacyUsecase(pharmacyRepo, 2*time.Second)
	_httpDelivery.NewPharmacyHandler(router, pharmacyUseCase, secretKey)

	prescriptionRepo := _patientPostgresRepository.NewPostgresPrescriptionRepository(db)
	prescriptionUseCase := _useCase.NewPrescriptionUsecase(prescriptionRepo, 2*time.Second)
	_httpDelivery.NewPrescriptionHandler(router, prescriptionUseCase, secretKey)

	router.Run(address)
}
