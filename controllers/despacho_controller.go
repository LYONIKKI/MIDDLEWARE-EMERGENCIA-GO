package controllers

import (
	"emergencia_go/config"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ExpedienteDespacho struct {
	HI                 string `json:"hi"`
	HF                 string `json:"hf"`
	FechaSolicitud     string `json:"fecha_solicitud"`
	HoraSolicitud      string `json:"hora_solicitud"`
	FechaCita          string `json:"fecha_cita"`
	DNI                string `json:"dni"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	Paciente           string `json:"paciente"`
	NroCuenta          int    `json:"nro_cuenta"`
	Servicio           string `json:"servicio"`
	Seguro             string `json:"seguro"`
	Medico             string `json:"medico"`
}

// GetCitasDespacho descarga los expedientes del día o mañana ordenados para archivo
func GetCitasDespacho(c *fiber.Ctx) error {
	periodo := strings.TrimSpace(c.Query("fecha")) // "hoy" o "manana"
	filtroFecha := "CAST(GETDATE() AS DATE)"
	if periodo == "manana" {
		filtroFecha = "DATEADD(day, 1, CAST(GETDATE() AS DATE))"
	}

	query := fmt.Sprintf(`
		SELECT 
			ISNULL(c.HoraInicio, '') AS HI,
			ISNULL(c.HoraFin, '') AS HF,
			ISNULL(CONVERT(VARCHAR(10), c.FechaSolicitud, 103), '') AS FechaSolicitud,
			ISNULL(c.HoraSolicitud, '') AS HoraSolicitud,
			ISNULL(CONVERT(VARCHAR(10), c.Fecha, 103), '') AS FechaCita,
			ISNULL(p.NroDocumento, '') AS DNI,
			ISNULL(p.NroHistoriaClinica, '') AS NroHistoriaClinica,
			UPPER(ISNULL(p.ApellidoPaterno,'') + ' ' + ISNULL(p.ApellidoMaterno,'') + ' ' + ISNULL(p.PrimerNombre,'') + ISNULL(' ' + p.SegundoNombre, '')) AS Paciente,
			ISNULL(a.IdCuentaAtencion, 0) AS NroCuenta,
			ISNULL(s.Nombre, '') AS Servicio,
			ISNULL(f.Descripcion, 'PARTICULAR') AS Seguro, 
			UPPER(ISNULL(e.ApellidoPaterno,'') + ' ' + ISNULL(e.ApellidoMaterno,'') + ' ' + ISNULL(e.Nombres, 'MEDICO DE TURNO')) AS Medico
		FROM [dbo].[Citas] c WITH (NOLOCK)
		INNER JOIN [dbo].[Pacientes] p WITH (NOLOCK) ON c.IdPaciente = p.IdPaciente
		LEFT JOIN [dbo].[Atenciones] a WITH (NOLOCK) ON c.IdAtencion = a.IdAtencion 
		INNER JOIN [dbo].[Servicios] s WITH (NOLOCK) ON c.IdServicio = s.IdServicio
		INNER JOIN [dbo].[Medicos] m WITH (NOLOCK) ON c.IdMedico = m.IdMedico
		INNER JOIN [dbo].[Empleados] e WITH (NOLOCK) ON m.IdEmpleado = e.IdEmpleado
		LEFT JOIN [dbo].[FuentesFinanciamiento] f WITH (NOLOCK) ON a.idFuenteFinanciamiento = f.idFuenteFinanciamiento
		WHERE c.Fecha = %s
		ORDER BY c.HoraInicio ASC, c.HoraFin ASC
	`, filtroFecha)

	rows, err := config.DBSQLServer.Query(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}
	defer rows.Close()

	var lista []ExpedienteDespacho
	for rows.Next() {
		var item ExpedienteDespacho
		if err := rows.Scan(
			&item.HI, &item.HF, &item.FechaSolicitud, &item.HoraSolicitud,
			&item.FechaCita, &item.DNI, &item.NroHistoriaClinica, &item.Paciente,
			&item.NroCuenta, &item.Servicio, &item.Seguro, &item.Medico,
		); err == nil {
			lista = append(lista, item)
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"total":  len(lista),
		"data":   lista,
	})
}
