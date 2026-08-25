package routes

import (
	"emergencia_go/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/dashboard", controllers.GetDashboardData)
	// Módulo Emergencia

	api.Get("/reporte/:periodo", controllers.GenerarReporteExcel)

	emg := api.Group("/emergencia")
	emg.Get("/pacientes-hoy", controllers.GetPacientesHoy)
	emg.Get("/ticket/:cuenta", controllers.GetTicketData)
	emg.Get("/historial-atenciones", controllers.GetHistorialAtenciones)

	arch := api.Group("/archivos")
	arch.Get("/citas-ce", controllers.GetCitasHoyManana)
	arch.Get("/paciente/:dni", controllers.BuscarPacientePorDNI)
	arch.Post("/actualizar-hc", controllers.ActualizarHistoriaClinica)

	// Trazabilidad y Movimientos
	arch.Post("/movimiento/salida", controllers.RegistrarSalidaHC)
	arch.Post("/movimiento/devolucion", controllers.RegistrarDevolucionHC)
	arch.Get("/movimiento/historial/:hc", controllers.ObtenerHistorialHC)
	arch.Get("/movimiento/ultimo-estado/:hc", controllers.ObtenerUltimoEstadoHC)

}
