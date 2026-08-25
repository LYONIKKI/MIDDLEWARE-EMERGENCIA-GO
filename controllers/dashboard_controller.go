package controllers

import (
	"database/sql"
	"emergencia_go/config"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DashboardResponse struct {
	TotalMes         int      `json:"total_mes"`
	TopMesServicio   string   `json:"top_mes_servicio"`
	TopMesTotal      int      `json:"top_mes_total"`
	CantidadHoyTotal int      `json:"cantidad_hoy_total"`
	TopHoyServicio   string   `json:"top_hoy_servicio"`
	GraficaLabels    []string `json:"grafica_labels"`
	GraficaData      []int    `json:"grafica_data"`
}

func GetDashboardData(c *fiber.Ctx) error {
	db := config.DBSQLServer

	mesParam := c.Query("mes")
	mes := int(time.Now().Month())
	if m, err := strconv.Atoi(mesParam); err == nil && m >= 1 && m <= 12 {
		mes = m
	}
	anio := time.Now().Year()

	var resp DashboardResponse
	resp.GraficaLabels = make([]string, 0)
	resp.GraficaData = make([]int, 0)

	// 1. Gráfica por Día
	queryGrafica := `SELECT FORMAT(FechaIngreso, 'dd/MM') AS dia, COUNT(*) AS total 
	                 FROM [dbo].[Atenciones] a 
	                 INNER JOIN [dbo].[AtencionesEmergencia] ae ON a.IdAtencion = ae.IdAtencion 
	                 WHERE MONTH(FechaIngreso) = @p1 AND YEAR(FechaIngreso) = @p2
	                 GROUP BY FORMAT(FechaIngreso, 'dd/MM'), CAST(FechaIngreso AS DATE) 
	                 ORDER BY CAST(FechaIngreso AS DATE) ASC`
	rows, err := db.Query(queryGrafica, mes, anio)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dia string
			var total int
			if err := rows.Scan(&dia, &total); err == nil {
				resp.GraficaLabels = append(resp.GraficaLabels, dia)
				resp.GraficaData = append(resp.GraficaData, total)
			}
		}
	}

	// 2. Total Pacientes Mes
	queryMes := `SELECT COUNT(DISTINCT IdPaciente) 
	             FROM [dbo].[Atenciones] a 
	             INNER JOIN [dbo].[AtencionesEmergencia] ae ON a.IdAtencion = ae.IdAtencion 
	             WHERE MONTH(FechaIngreso) = @p1 AND YEAR(FechaIngreso) = @p2`
	_ = db.QueryRow(queryMes, mes, anio).Scan(&resp.TotalMes)

	// 3. Servicio Top Mes
	queryTopMes := `SELECT TOP 1 s.Nombre, COUNT(a.IdAtencion) AS Total 
	                FROM [dbo].[Atenciones] a 
	                INNER JOIN [dbo].[Servicios] s ON a.IdServicioIngreso = s.IdServicio 
	                INNER JOIN [dbo].[AtencionesEmergencia] ae ON a.IdAtencion = ae.IdAtencion 
	                WHERE MONTH(a.FechaIngreso) = @p1 AND YEAR(a.FechaIngreso) = @p2
	                GROUP BY s.Nombre ORDER BY Total DESC`
	var topServ sql.NullString
	var topTot sql.NullInt64
	if err := db.QueryRow(queryTopMes, mes, anio).Scan(&topServ, &topTot); err == nil {
		resp.TopMesServicio = topServ.String
		resp.TopMesTotal = int(topTot.Int64)
	} else {
		resp.TopMesServicio = "Sin datos"
	}

	// 4. Ingresos Totales Hoy y Servicio Top Hoy
	queryHoy := `SELECT COUNT(*) 
	             FROM [dbo].[Atenciones] a 
	             INNER JOIN [dbo].[AtencionesEmergencia] ae ON a.IdAtencion = ae.IdAtencion 
	             WHERE CAST(a.FechaIngreso AS DATE) = CAST(GETDATE() AS DATE)`
	_ = db.QueryRow(queryHoy).Scan(&resp.CantidadHoyTotal)

	queryTopHoy := `SELECT TOP 1 s.Nombre 
	                FROM [dbo].[Atenciones] a 
	                INNER JOIN [dbo].[Servicios] s ON a.IdServicioIngreso = s.IdServicio 
	                INNER JOIN [dbo].[AtencionesEmergencia] ae ON a.IdAtencion = ae.IdAtencion 
	                WHERE CAST(a.FechaIngreso AS DATE) = CAST(GETDATE() AS DATE) 
	                GROUP BY s.Nombre ORDER BY COUNT(a.IdAtencion) DESC`
	var topHoyServ sql.NullString
	if err := db.QueryRow(queryTopHoy).Scan(&topHoyServ); err == nil {
		resp.TopHoyServicio = topHoyServ.String
	} else {
		resp.TopHoyServicio = "N/A"
	}

	return c.JSON(resp)
}
