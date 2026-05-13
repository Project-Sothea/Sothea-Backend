package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"time"

	_httpDelivery "sothea-backend/controllers"
	_postgresRepository "sothea-backend/repository/postgres"
	_useCase "sothea-backend/usecases"

	"github.com/gin-contrib/cors"
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

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
	router.Run("0.0.0.0:" + port)
}
