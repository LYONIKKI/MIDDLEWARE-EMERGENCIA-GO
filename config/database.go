package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

var (
	DBPostgres  *sql.DB // Conexión a autenticación (PostgreSQL)
	DBSQLServer *sql.DB // Conexión a SIGH (SQL Server)
)

// getEnv obtiene una variable de entorno o usa un valor por defecto si no existe
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func ConnectDBs() {
	var err error

	// 1. Configuración de PostgreSQL leída desde variables de entorno
	pgHost := getEnv("DB_PG_HOST", "127.0.0.1")
	pgPort := getEnv("DB_PG_PORT", "5432")
	pgUser := getEnv("DB_PG_USER", "postgres")
	pgPass := getEnv("DB_PG_PASS", "secret")
	pgName := getEnv("DB_PG_NAME", "auth_db")

	pgConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPass, pgName)

	DBPostgres, err = sql.Open("postgres", pgConnStr)
	if err != nil || DBPostgres.Ping() != nil {
		log.Printf("[WARN] No se pudo conectar a Postgres: %v", err)
	}

	// 2. Configuración de SQL Server leída desde variables de entorno
	sqlHost := getEnv("DB_SQLSERVER_HOST", "127.0.0.1")
	sqlUser := getEnv("DB_SQLSERVER_USER", "sa")
	sqlPass := getEnv("DB_SQLSERVER_PASS", "secret")
	sqlName := getEnv("DB_SQLSERVER_NAME", "SIGH")

	sqlConnStr := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=disable",
		sqlHost, sqlUser, sqlPass, sqlName)

	DBSQLServer, err = sql.Open("sqlserver", sqlConnStr)
	if err != nil || DBSQLServer.Ping() != nil {
		log.Fatalf("Error conectando a SQL Server: %v", err)
	}

	// Optimización de Pool de conexiones
	DBSQLServer.SetMaxOpenConns(25)
	DBSQLServer.SetMaxIdleConns(10)
	DBSQLServer.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Conexiones exitosas a PostgreSQL (Auth) y SQL Server (SIGH)")
}
