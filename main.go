package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"time"

	_httpDelivery "sothea-backend/controllers"
	"sothea-backend/middlewares"
	_postgresRepository "sothea-backend/repository/postgres"
	_useCase "sothea-backend/usecases"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	godotenv.Load()
	viper.AutomaticEnv()

	port := viper.GetString("PORT")
	connStr := viper.GetString("DATABASE_URL")
	secretKey := []byte(viper.GetString("SECRET_KEY"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	if len(secretKey) == 0 {
		log.Fatal("SECRET_KEY environment variable is not set")
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()
	// Test postgres connection
	if _, err := pool.Exec(ctx, "SELECT 1"); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Use(middlewares.ErrorLogger())

	if viper.GetString("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
		// only trust localhost
		router.SetTrustedProxies([]string{"127.0.0.1"})
	}

	const distDir = "./dist"
	router.StaticFile("/", filepath.Join(distDir, "index.html"))
	router.Static("/assets", "./dist/assets")

	api := router.Group("/api")

	userRepo := _postgresRepository.NewPostgresUserRepository(pool)

	timezone := viper.GetString("TIMEZONE")
	if timezone == "" {
		timezone = "Asia/Phnom_Penh"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Fatal(err)
	}
	patientRepo := _postgresRepository.NewPostgresPatientRepository(pool, loc)

	loginUseCase := _useCase.NewLoginUseCase(userRepo, 30*time.Second, secretKey)
	_httpDelivery.NewLoginHandler(api, loginUseCase, secretKey)

	patientUseCase := _useCase.NewPatientUsecase(patientRepo, 30*time.Second)
	_httpDelivery.NewPatientHandler(api, patientUseCase, secretKey)

	pharmacyRepo := _postgresRepository.NewPostgresPharmacyRepository(pool)
	pharmacyUseCase := _useCase.NewPharmacyUsecase(pharmacyRepo, 30*time.Second)
	_httpDelivery.NewPharmacyHandler(api, pharmacyUseCase, secretKey, pool)

	prescriptionRepo := _postgresRepository.NewPostgresPrescriptionRepository(pool)
	prescriptionUseCase := _useCase.NewPrescriptionUsecase(prescriptionRepo, pharmacyRepo, 30*time.Second)
	_httpDelivery.NewPrescriptionHandler(api, prescriptionUseCase, secretKey, pool)

	router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./dist/index.html")
		}
	})
	log.Printf("Server is running on port %s", port)
	router.Run("0.0.0.0:" + port)
}
