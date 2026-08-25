package controllers

import (
	"database/sql"
	"strings"

	"emergencia_go/config"

	"github.com/gofiber/fiber/v2"
)

type PacienteHCResponse struct {
	IdPaciente         int    `json:"id_paciente"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	NombreCompleto     string `json:"nombre_completo"`
}

type UpdateHCRequest struct {
	IdPaciente int    `json:"id_paciente"`
	NuevoHC    string `json:"nuevo_hc"`
}

// GET /api/archivos/paciente/:dni
func BuscarPacientePorDNI(c *fiber.Ctx) error {
	db := config.DBSQLServer
	termino := strings.TrimSpace(c.Params("dni"))

	if termino == "" {
		return c.Status(400).JSON(fiber.Map{"error": "No se proporcionó término de búsqueda"})
	}

	query := `SELECT TOP 1 
	            IdPaciente, 
	            ISNULL(NroHistoriaClinica, '') AS NroHistoriaClinica, 
	            UPPER(ISNULL(ApellidoPaterno,'') + ' ' + ISNULL(ApellidoMaterno,'') + ' ' + ISNULL(PrimerNombre,'') + ' ' + ISNULL(SegundoNombre, '')) AS NombreCompleto
	        FROM [dbo].[Pacientes] WITH (NOLOCK)
	        WHERE NroDocumento = @p1 OR NroHistoriaClinica = @p2`

	var idPaciente int
	var hc, nombre sql.NullString

	err := db.QueryRow(query, termino, termino).Scan(&idPaciente, &hc, &nombre)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Paciente no encontrado"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Error al consultar base de datos: " + err.Error()})
	}

	nroHC := strings.TrimSpace(hc.String)
	if nroHC == "" {
		nroHC = "SIN ASIGNAR / EN BLANCO"
	}

	resp := PacienteHCResponse{
		IdPaciente:         idPaciente,
		NroHistoriaClinica: nroHC,
		NombreCompleto:     strings.Join(strings.Fields(nombre.String), " "),
	}

	return c.JSON(resp)
}

// POST /api/archivos/actualizar-hc
func ActualizarHistoriaClinica(c *fiber.Ctx) error {
	db := config.DBSQLServer

	var req UpdateHCRequest
	if err := c.BodyParser(&req); err != nil || req.IdPaciente <= 0 || strings.TrimSpace(req.NuevoHC) == "" {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Datos inválidos"})
	}

	req.NuevoHC = strings.TrimSpace(req.NuevoHC)

	query := `UPDATE [dbo].[Pacientes] 
	          SET NroHistoriaClinica = @p1 
	          WHERE IdPaciente = @p2`

	res, err := db.Exec(query, req.NuevoHC, req.IdPaciente)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Error al actualizar: " + err.Error()})
	}

	filas, _ := res.RowsAffected()
	if filas == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No se encontró el paciente para actualizar"})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Historia clínica actualizada correctamente",
	})
}

type CitaResponse struct {
	FechaSolicitud     string `json:"fecha_solicitud"`
	FechaCita          string `json:"fecha_cita"`
	HoraCita           string `json:"hora_cita"`
	DNI                string `json:"dni"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	Paciente           string `json:"paciente"`
	NroCuenta          int    `json:"nro_cuenta"`
	Servicio           string `json:"servicio"`
	Seguro             string `json:"seguro"`
	Medico             string `json:"medico"`
}

// GET /api/archivos/citas-ce?tipo=hoy (o tipo=manana)
func GetCitasHoyManana(c *fiber.Ctx) error {
	db := config.DBSQLServer
	tipo := strings.ToLower(c.Query("tipo", "hoy"))

	var whereDate string
	if tipo == "manana" {
		whereDate = `CAST(c.Fecha AS DATE) = CAST(
			DATEADD(
				day, 
				CASE 
					WHEN DATENAME(WEEKDAY, GETDATE()) IN ('Saturday', 'Sábado') THEN 2
					ELSE 1
				END, 
				GETDATE()
			) AS DATE)`
	} else {
		whereDate = `CAST(c.Fecha AS DATE) = CAST(GETDATE() AS DATE)`
	}

	query := `SELECT 
	            ISNULL(CONVERT(VARCHAR(10), c.FechaSolicitud, 103), '') AS FechaSolicitud,
	            ISNULL(CONVERT(VARCHAR(10), c.Fecha, 103), '') AS FechaCita,
	            ISNULL(c.HoraInicio, '') AS HoraCita,
	            ISNULL(p.NroDocumento, '') AS DNI,
	            ISNULL(p.NroHistoriaClinica, '') AS NroHistoriaClinica,
	            UPPER(ISNULL(p.ApellidoPaterno,'') + ' ' + ISNULL(p.ApellidoMaterno,'') + ' ' + ISNULL(PrimerNombre,'') + ISNULL(' ' + SegundoNombre, '')) AS Paciente,
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
	        WHERE ` + whereDate + `
	        ORDER BY c.IdAtencion ASC`

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	citas := make([]CitaResponse, 0, 50)
	for rows.Next() {
		var item CitaResponse
		var medicoRaw, pacienteRaw string
		if err := rows.Scan(
			&item.FechaSolicitud,
			&item.FechaCita,
			&item.HoraCita,
			&item.DNI,
			&item.NroHistoriaClinica,
			&pacienteRaw,
			&item.NroCuenta,
			&item.Servicio,
			&item.Seguro,
			&medicoRaw,
		); err == nil {
			item.Paciente = strings.Join(strings.Fields(pacienteRaw), " ")
			item.Medico = strings.Join(strings.Fields(medicoRaw), " ")
			citas = append(citas, item)
		}
	}

	return c.JSON(citas)
}

// Estructuras de Petición / Respuesta
type SalidaHCRequest struct {
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	DNI                string `json:"dni"`
	Paciente           string `json:"paciente"`
	IdCuentaAtencion   int    `json:"id_cuenta_atencion"`
	ServicioDestino    string `json:"servicio_destino"`
	QuienRecoge        string `json:"quien_recoge"`
	UsuarioSalida      string `json:"usuario_salida"`
	Observaciones      string `json:"observaciones"`
}

type DevolucionHCRequest struct {
	IdMovimiento      int    `json:"id_movimiento"`
	QuienDevuelve     string `json:"quien_devuelve"`
	UbicacionArchivo  string `json:"ubicacion_archivo"`
	UsuarioDevolucion string `json:"usuario_devolucion"`
	Observaciones     string `json:"observaciones"`
}

type MovimientoItem struct {
	IdMovimiento       int    `json:"id_movimiento"`
	NroHistoriaClinica string `json:"nro_historia_clinica"`
	DNI                string `json:"dni"`
	Paciente           string `json:"paciente"`
	ServicioDestino    string `json:"servicio_destino"`
	FechaSalida        string `json:"fecha_salida"`
	QuienRecoge        string `json:"quien_recoge"`
	UsuarioSalida      string `json:"usuario_salida"`
	FechaDevolucion    string `json:"fecha_devolucion"`
	QuienDevuelve      string `json:"quien_devuelve"`
	UbicacionArchivo   string `json:"ubicacion_archivo"`
	UsuarioDevolucion  string `json:"usuario_devolucion"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones"`
}

// POST /api/archivos/movimiento/salida
func RegistrarSalidaHC(c *fiber.Ctx) error {
	db := config.DBSQLServer

	var req SalidaHCRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.NroHistoriaClinica) == "" || strings.TrimSpace(req.QuienRecoge) == "" {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Datos incompletos para registrar la salida"})
	}

	query := `INSERT INTO [dbo].[MovimientosHC] 
	            (NroHistoriaClinica, DNI, Paciente, IdCuentaAtencion, ServicioDestino, QuienRecoge, UsuarioSalida, Observaciones, Estado) 
	          VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, 'SALIENTE')`

	_, err := db.Exec(query,
		strings.TrimSpace(req.NroHistoriaClinica),
		strings.TrimSpace(req.DNI),
		strings.TrimSpace(req.Paciente),
		req.IdCuentaAtencion,
		strings.TrimSpace(req.ServicioDestino),
		strings.TrimSpace(req.QuienRecoge),
		strings.TrimSpace(req.UsuarioSalida),
		strings.TrimSpace(req.Observaciones),
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Error al registrar salida: " + err.Error()})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Salida de historia clínica registrada correctamente"})
}

// POST /api/archivos/movimiento/devolucion
func RegistrarDevolucionHC(c *fiber.Ctx) error {
	db := config.DBSQLServer

	var req DevolucionHCRequest
	if err := c.BodyParser(&req); err != nil || req.IdMovimiento <= 0 || strings.TrimSpace(req.QuienDevuelve) == "" || strings.TrimSpace(req.UbicacionArchivo) == "" {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Debe especificar quién devuelve y la ubicación física en archivo"})
	}

	query := `UPDATE [dbo].[MovimientosHC] 
	          SET FechaDevolucion = GETDATE(),
	              QuienDevuelve = @p1,
	              UbicacionArchivo = @p2,
	              UsuarioDevolucion = @p3,
	              Observaciones = ISNULL(Observaciones + ' | ', '') + @p4,
	              Estado = 'DEVUELTO'
	          WHERE IdMovimiento = @p5 AND Estado = 'SALIENTE'`

	res, err := db.Exec(query,
		strings.TrimSpace(req.QuienDevuelve),
		strings.TrimSpace(req.UbicacionArchivo),
		strings.TrimSpace(req.UsuarioDevolucion),
		strings.TrimSpace(req.Observaciones),
		req.IdMovimiento,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Error al registrar devolución: " + err.Error()})
	}

	filas, _ := res.RowsAffected()
	if filas == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No se encontró el registro pendiente de devolución"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Historia clínica devuelta y ubicada en archivo correctamente"})
}

// GET /api/archivos/movimiento/historial/:hc
func ObtenerHistorialHC(c *fiber.Ctx) error {
	db := config.DBSQLServer
	hc := strings.TrimSpace(c.Params("hc"))

	query := `SELECT 
	            IdMovimiento, NroHistoriaClinica, ISNULL(DNI,''), Paciente, ServicioDestino,
	            CONVERT(VARCHAR(19), FechaSalida, 120) AS FechaSalida,
	            QuienRecoge, UsuarioSalida,
	            ISNULL(CONVERT(VARCHAR(19), FechaDevolucion, 120), '') AS FechaDevolucion,
	            ISNULL(QuienDevuelve, ''),
	            ISNULL(UbicacionArchivo, 'NO ASIGNADA'),
	            ISNULL(UsuarioDevolucion, ''),
	            Estado,
	            ISNULL(Observaciones, '')
	          FROM [dbo].[MovimientosHC] WITH (NOLOCK)
	          WHERE NroHistoriaClinica = @p1 OR DNI = @p1
	          ORDER BY FechaSalida DESC`

	rows, err := db.Query(query, hc)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	historial := make([]MovimientoItem, 0)
	for rows.Next() {
		var m MovimientoItem
		if err := rows.Scan(
			&m.IdMovimiento, &m.NroHistoriaClinica, &m.DNI, &m.Paciente, &m.ServicioDestino,
			&m.FechaSalida, &m.QuienRecoge, &m.UsuarioSalida, &m.FechaDevolucion,
			&m.QuienDevuelve, &m.UbicacionArchivo, &m.UsuarioDevolucion, &m.Estado, &m.Observaciones,
		); err == nil {
			historial = append(historial, m)
		}
	}

	return c.JSON(historial)
}

// GET /api/archivos/movimiento/ultimo-estado/:hc
func ObtenerUltimoEstadoHC(c *fiber.Ctx) error {
	db := config.DBSQLServer
	hc := strings.TrimSpace(c.Params("hc"))

	query := `SELECT TOP 1 
	            IdMovimiento, NroHistoriaClinica, ISNULL(DNI,''), Paciente, ServicioDestino,
	            CONVERT(VARCHAR(19), FechaSalida, 120) AS FechaSalida,
	            QuienRecoge, UsuarioSalida,
	            ISNULL(CONVERT(VARCHAR(19), FechaDevolucion, 120), '') AS FechaDevolucion,
	            ISNULL(QuienDevuelve, ''),
	            ISNULL(UbicacionArchivo, 'ARCHIVO GENERAL'),
	            Estado
	          FROM [dbo].[MovimientosHC] WITH (NOLOCK)
	          WHERE NroHistoriaClinica = @p1 OR DNI = @p1
	          ORDER BY IdMovimiento DESC`

	var m MovimientoItem
	err := db.QueryRow(query, hc).Scan(
		&m.IdMovimiento, &m.NroHistoriaClinica, &m.DNI, &m.Paciente, &m.ServicioDestino,
		&m.FechaSalida, &m.QuienRecoge, &m.UsuarioSalida, &m.FechaDevolucion,
		&m.QuienDevuelve, &m.UbicacionArchivo, &m.Estado,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(fiber.Map{"estado": "EN_ARCHIVO", "ubicacion_archivo": "ARCHIVO GENERAL"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(m)
}
