package controllers

import (
	"bytes"
	"database/sql"
	"emergencia_go/config"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PacienteHoyResponse struct {
	DNI                string `json:"dni"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	Paciente           string `json:"paciente"`
	NroCuenta          int    `json:"nro_cuenta"`
	Seguro             string `json:"seguro"`
	FechaIngreso       string `json:"fecha_ingreso"`
	HoraIngreso        string `json:"hora_ingreso"`
	ServicioEmergencia string `json:"servicio_emergencia"`
	MedicoIngreso      string `json:"medico_ingreso"`
}

type TicketResponse struct {
	DNI                string `json:"dni"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	Paciente           string `json:"paciente"`
	NroCuenta          int    `json:"nro_cuenta"`
	FechaIngreso       string `json:"fecha_ingreso"`
	HoraIngreso        string `json:"hora_ingreso"`
	Servicio           string `json:"servicio"`
	Medico             string `json:"medico"`
}

// GET /api/emergencia/pacientes-hoy
// GET /api/emergencia/pacientes-hoy
func GetPacientesHoy(c *fiber.Ctx) error {
	db := config.DBSQLServer

	// Usamos CONVERT y WITH (NOLOCK) para respuesta ultrarrápida
	query := `SELECT 
	            ISNULL(p.NroDocumento, '') AS DNI,
	            ISNULL(p.NroHistoriaClinica, '') AS NroHistoriaClinica,
	            UPPER(p.ApellidoPaterno + ' ' + p.ApellidoMaterno + ' ' + p.PrimerNombre + ISNULL(' ' + p.SegundoNombre, '')) AS Paciente,
	            a.IdCuentaAtencion AS NroCuenta,
	            ISNULL(f.Descripcion, 'PARTICULAR') AS Seguro, 
	            CONVERT(VARCHAR(10), a.FechaIngreso, 103) AS FechaIngreso,
	            ISNULL(a.HoraIngreso, '') AS HoraIngreso,
	            ISNULL(s.Nombre, '') AS ServicioEmergencia,
	            UPPER(ISNULL(e.ApellidoPaterno + ' ' + e.ApellidoMaterno + ' ' + e.Nombres, 'MEDICO DE TURNO')) AS MedicoIngreso
	        FROM [dbo].[Atenciones] a WITH (NOLOCK)
	        INNER JOIN [dbo].[Pacientes] p WITH (NOLOCK) ON a.IdPaciente = p.IdPaciente
	        INNER JOIN [dbo].[Servicios] s WITH (NOLOCK) ON a.IdServicioIngreso = s.IdServicio
	        INNER JOIN [dbo].[AtencionesEmergencia] ae WITH (NOLOCK) ON a.IdAtencion = ae.IdAtencion
	        INNER JOIN [dbo].[Medicos] m WITH (NOLOCK) ON a.IdMedicoIngreso = m.IdMedico
	        INNER JOIN [dbo].[Empleados] e WITH (NOLOCK) ON m.IdEmpleado = e.IdEmpleado
	        INNER JOIN [dbo].[FuentesFinanciamiento] f WITH (NOLOCK) ON a.IdFuenteFinanciamiento = f.idFuenteFinanciamiento
	        WHERE a.FechaIngreso >= CAST(GETDATE() AS DATE) 
	          AND a.FechaIngreso < DATEADD(DAY, 1, CAST(GETDATE() AS DATE))
	        ORDER BY a.HoraIngreso DESC`

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	pacientes := make([]PacienteHoyResponse, 0, 50)
	for rows.Next() {
		var p PacienteHoyResponse
		if err := rows.Scan(&p.DNI, &p.NroHistoriaClinica, &p.Paciente, &p.NroCuenta, &p.Seguro, &p.FechaIngreso, &p.HoraIngreso, &p.ServicioEmergencia, &p.MedicoIngreso); err == nil {
			pacientes = append(pacientes, p)
		}
	}

	return c.JSON(pacientes)
}

// GET /api/emergencia/ticket/:cuenta
func GetTicketData(c *fiber.Ctx) error {
	db := config.DBSQLServer
	cuenta := c.Params("cuenta")

	query := `SELECT 
	            ISNULL(p.NroDocumento, '') AS DNI,
	            ISNULL(p.NroHistoriaClinica, '') AS NroHistoriaClinica,
	            UPPER(p.ApellidoPaterno + ' ' + p.ApellidoMaterno + ' ' + p.PrimerNombre + ISNULL(' ' + p.SegundoNombre, '')) AS Paciente,
	            a.IdCuentaAtencion AS NroCuenta,
	            FORMAT(a.FechaIngreso, 'dd/MM/yyyy') AS FechaIngreso,
	            ISNULL(a.HoraIngreso, '') AS HoraIngreso,
	            ISNULL(s.Nombre, '') AS Servicio,
	            UPPER(ISNULL(e.ApellidoPaterno + ' ' + e.ApellidoMaterno + ' ' + e.Nombres, 'MEDICO DE TURNO')) AS Medico
	        FROM [dbo].[Atenciones] a
	        INNER JOIN [dbo].[Pacientes] p ON a.IdPaciente = p.IdPaciente
	        INNER JOIN [dbo].[Servicios] s ON a.IdServicioIngreso = s.IdServicio
	        INNER JOIN [dbo].[Medicos] m ON a.IdMedicoIngreso = m.IdMedico
	        INNER JOIN [dbo].[Empleados] e ON m.IdEmpleado = e.IdEmpleado
	        WHERE a.IdCuentaAtencion = @p1`

	var t TicketResponse
	err := db.QueryRow(query, cuenta).Scan(&t.DNI, &t.NroHistoriaClinica, &t.Paciente, &t.NroCuenta, &t.FechaIngreso, &t.HoraIngreso, &t.Servicio, &t.Medico)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Ticket no encontrado"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	t.Paciente = strings.TrimSpace(t.Paciente)
	t.Medico = strings.TrimSpace(t.Medico)
	return c.JSON(t)
}

type HistorialAtencionResponse struct {
	FechaIngreso       string `json:"fecha_ingreso"`
	HoraIngreso        string `json:"hora_ingreso"`
	Paciente           string `json:"paciente"`
	DNI                string `json:"dni"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	NroCuenta          int    `json:"nro_cuenta"`
	Seguro             string `json:"seguro"`
	ServicioEmergencia string `json:"servicio_emergencia"`
	MedicoIngreso      string `json:"medico_ingreso"`
}

// GET /api/emergencia/historial-atenciones?rango=0 (0: este mes, -1: hace 1 mes)
func GetHistorialAtenciones(c *fiber.Ctx) error {
	db := config.DBSQLServer

	rangoParam := c.Query("rango", "0")
	mesesAtras, err := strconv.Atoi(rangoParam)
	if err != nil {
		mesesAtras = 0
	}

	query := `SELECT 
	            CONVERT(VARCHAR(10), a.FechaIngreso, 103) AS FechaIngreso,
	            ISNULL(a.HoraIngreso, '') AS HoraIngreso,
	            UPPER(p.ApellidoPaterno + ' ' + p.ApellidoMaterno + ' ' + p.PrimerNombre + ISNULL(' ' + p.SegundoNombre, '')) AS Paciente,
	            ISNULL(p.NroDocumento, '') AS DNI,
	            ISNULL(p.NroHistoriaClinica, '') AS NroHistoriaClinica,
	            a.IdCuentaAtencion AS NroCuenta,
	            ISNULL(f.Descripcion, 'PARTICULAR') AS Seguro, 
	            ISNULL(s.Nombre, '') AS ServicioEmergencia,
	            UPPER(ISNULL(e.ApellidoPaterno + ' ' + e.ApellidoMaterno + ' ' + e.Nombres, 'MEDICO DE TURNO')) AS MedicoIngreso
	        FROM [dbo].[Atenciones] a WITH (NOLOCK)
	        INNER JOIN [dbo].[Pacientes] p WITH (NOLOCK) ON a.IdPaciente = p.IdPaciente
	        INNER JOIN [dbo].[Servicios] s WITH (NOLOCK) ON a.IdServicioIngreso = s.IdServicio
	        INNER JOIN [dbo].[AtencionesEmergencia] ae WITH (NOLOCK) ON a.IdAtencion = ae.IdAtencion
	        INNER JOIN [dbo].[Medicos] m WITH (NOLOCK) ON a.IdMedicoIngreso = m.IdMedico
	        INNER JOIN [dbo].[Empleados] e WITH (NOLOCK) ON m.IdEmpleado = e.IdEmpleado
	        INNER JOIN [dbo].[FuentesFinanciamiento] f WITH (NOLOCK) ON a.IdFuenteFinanciamiento = f.idFuenteFinanciamiento
	        WHERE a.FechaIngreso >= DATEADD(month, @p1, CAST(DATEADD(month, DATEDIFF(month, 0, GETDATE()), 0) AS DATETIME))
	        ORDER BY a.FechaIngreso DESC, a.HoraIngreso DESC`

	rows, err := db.Query(query, mesesAtras)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	historial := make([]HistorialAtencionResponse, 0, 100)
	for rows.Next() {
		var h HistorialAtencionResponse
		if err := rows.Scan(&h.FechaIngreso, &h.HoraIngreso, &h.Paciente, &h.DNI, &h.NroHistoriaClinica, &h.NroCuenta, &h.Seguro, &h.ServicioEmergencia, &h.MedicoIngreso); err == nil {
			historial = append(historial, h)
		}
	}

	return c.JSON(historial)
}

// GET /api/reporte/:periodo (hoy | ayer | mes)
func GenerarReporteExcel(c *fiber.Ctx) error {
	db := config.DBSQLServer
	periodo := strings.ToLower(c.Params("periodo"))

	now := time.Now()
	var whereClause, fechaTitulo, filename string

	switch periodo {
	case "ayer":
		whereClause = "CAST(a.FechaIngreso AS DATE) = CAST(DATEADD(day, -1, GETDATE()) AS DATE)"
		ayer := now.AddDate(0, 0, -1)
		fechaTitulo = ayer.Format("02/01/2006")
		filename = fmt.Sprintf("Atenciones_Emergencia_Ayer_%s.xls", ayer.Format("02-01-2006"))
	case "mes":
		whereClause = "MONTH(a.FechaIngreso) = MONTH(GETDATE()) AND YEAR(a.FechaIngreso) = YEAR(GETDATE())"
		fechaTitulo = fmt.Sprintf("MES DE %s %d", strings.ToUpper(now.Month().String()), now.Year())
		filename = fmt.Sprintf("Atenciones_Emergencia_Mes_%s.xls", now.Format("01-2006"))
	default: // hoy
		whereClause = "CAST(a.FechaIngreso AS DATE) = CAST(GETDATE() AS DATE)"
		fechaTitulo = now.Format("02/01/2006")
		filename = fmt.Sprintf("Atenciones_Emergencia_Hoy_%s.xls", now.Format("02-01-2006"))
	}

	query := fmt.Sprintf(`SELECT 
	            CONVERT(VARCHAR(10), a.FechaIngreso, 103) AS FechaIngreso,
	            ISNULL(a.HoraIngreso, '') AS HoraIngreso,
	            ISNULL(p.NroDocumento, '') AS DNI,
	            ISNULL(p.NroHistoriaClinica, '') AS NroHistoriaClinica,
	            UPPER(p.ApellidoPaterno + ' ' + p.ApellidoMaterno + ' ' + p.PrimerNombre + ISNULL(' ' + p.SegundoNombre, '')) AS Paciente,
	            a.IdCuentaAtencion AS NroCuenta,
	            ISNULL(f.Descripcion, 'PARTICULAR') AS Seguro,
	            ISNULL(s.Nombre, '') AS ServicioEmergencia,
	            UPPER(ISNULL(e.ApellidoPaterno + ' ' + e.ApellidoMaterno + ' ' + e.Nombres, 'MEDICO DE TURNO')) AS MedicoIngreso
	        FROM [dbo].[Atenciones] a WITH (NOLOCK)
	        INNER JOIN [dbo].[Pacientes] p WITH (NOLOCK) ON a.IdPaciente = p.IdPaciente
	        INNER JOIN [dbo].[Servicios] s WITH (NOLOCK) ON a.IdServicioIngreso = s.IdServicio
	        INNER JOIN [dbo].[AtencionesEmergencia] ae WITH (NOLOCK) ON a.IdAtencion = ae.IdAtencion
	        INNER JOIN [dbo].[FuentesFinanciamiento] f WITH (NOLOCK) ON a.IdFuenteFinanciamiento = f.idFuenteFinanciamiento
	        INNER JOIN [dbo].[Medicos] m WITH (NOLOCK) ON a.IdMedicoIngreso = m.IdMedico
	        INNER JOIN [dbo].[Empleados] e WITH (NOLOCK) ON m.IdEmpleado = e.IdEmpleado
	        WHERE %s
	        ORDER BY a.FechaIngreso ASC, a.HoraIngreso ASC`, whereClause)

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).SendString("Error consultando la base de datos: " + err.Error())
	}
	defer rows.Close()

	var buf bytes.Buffer
	// BOM UTF-8 para garantizar acentos en Microsoft Excel
	buf.WriteString("\xEF\xBB\xBF")
	buf.WriteString(`<html><head><meta charset="utf-8"></head><body>`)
	buf.WriteString(`<table border="1">`)
	buf.WriteString(`<thead>`)
	buf.WriteString(fmt.Sprintf(`<tr style="background-color: #d93025; color: white; font-weight: bold; text-align: center;">
		<th colspan="9" style="font-size: 16px; height: 35px;">REPORTE SIGH: %s</th>
	</tr>`, fechaTitulo))
	buf.WriteString(`<tr style="background-color: #202124; color: white; font-weight: bold; font-size: 12px; text-align: center;">
		<th>FECHA</th><th>HORA</th><th>DNI</th><th>H.C.</th><th>PACIENTE</th><th>NRO. CUENTA</th><th>SEGURO</th><th>SERVICIO</th><th>MEDICO</th>
	</tr>`)
	buf.WriteString(`</thead><tbody>`)

	for rows.Next() {
		var fecha, hora, dni, hc, paciente, seguro, servicio, medico string
		var nroCuenta int

		if err := rows.Scan(&fecha, &hora, &dni, &hc, &paciente, &nroCuenta, &seguro, &servicio, &medico); err == nil {
			buf.WriteString(fmt.Sprintf(`<tr>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
				<td>%s</td>
				<td style="text-align: center;">%d</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>`, fecha, hora, dni, hc, paciente, nroCuenta, seguro, servicio, medico))
		}
	}
	buf.WriteString(`</tbody></table></body></html>`)

	c.Set("Content-Type", "application/vnd.ms-excel; charset=UTF-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	return c.Send(buf.Bytes())
}
