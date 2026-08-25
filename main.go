package main

import (
	"emergencia_go/config"
	"emergencia_go/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 1. Inicializar conexiones
	config.ConnectDBs()

	// 2. Instanciar Fiber
	app := fiber.New(fiber.Config{
		AppName: "SIGH V2 - Emergencia Go API",
	})

	// 3. Middlewares
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// 4. Ruta de prueba
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "ok",
			"message": "API Conectada",
		})
	})

	// 5. REGISTRO DE TODAS LAS RUTAS
	routes.SetupRoutes(app)

	// 6. Iniciar Servidor
	log.Println("Servidor Go corriendo en http://0.0.0.0:8080")
	log.Fatal(app.Listen(":8080"))
}
