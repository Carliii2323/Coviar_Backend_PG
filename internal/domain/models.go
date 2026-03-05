package domain

import "time"

// ============================================
// MODELOS DE BODEGA
// ============================================

type Bodega struct {
	ID                 int       `json:"id_bodega,omitempty"`
	RazonSocial        string    `json:"razon_social"`
	NombreFantasia     string    `json:"nombre_fantasia"`
	CUIT               string    `json:"cuit"`              // char(11), check: ^[0-9]{11}$
	InvBod             *string   `json:"inv_bod,omitempty"` // char(6)
	InvVin             *string   `json:"inv_vin,omitempty"` // char(6)
	Calle              string    `json:"calle"`
	Numeracion         string    `json:"numeracion"` // default 'S/N'
	IDLocalidad        int       `json:"id_localidad"`
	Telefono           string    `json:"telefono"`            // check: ^[0-9]+$
	EmailInstitucional string    `json:"email_institucional"` // check: like '%@%'
	FechaRegistro      time.Time `json:"fecha_registro,omitempty"`
}

type BodegaRequest struct {
	RazonSocial        string  `json:"razon_social"`
	NombreFantasia     string  `json:"nombre_fantasia"`
	CUIT               string  `json:"cuit"`
	InvBod             *string `json:"inv_bod,omitempty"`
	InvVin             *string `json:"inv_vin,omitempty"`
	Calle              string  `json:"calle"`
	Numeracion         string  `json:"numeracion"`
	IDLocalidad        int     `json:"id_localidad"`
	Telefono           string  `json:"telefono"`
	EmailInstitucional string  `json:"email_institucional"`
}

type BodegaUpdateDTO struct {
	Telefono           string `json:"telefono"`
	EmailInstitucional string `json:"email_institucional"`
	NombreFantasia     string `json:"nombre_fantasia"`
}

// BodegaAdminItem es el DTO para listar bodegas en el panel admin,
// incluyendo el segmento y nivel de sostenibilidad de su última evaluación completada.
type BodegaAdminItem struct {
	ID                    int        `json:"id_bodega"`
	RazonSocial           string     `json:"razon_social"`
	NombreFantasia        string     `json:"nombre_fantasia"`
	CUIT                  string     `json:"cuit"`
	InvBod                *string    `json:"inv_bod,omitempty"`
	InvVin                *string    `json:"inv_vin,omitempty"`
	Calle                 string     `json:"calle"`
	Numeracion            string     `json:"numeracion"`
	IDLocalidad           int        `json:"id_localidad"`
	Telefono              string     `json:"telefono"`
	EmailInstitucional    string     `json:"email_institucional"`
	FechaRegistro         time.Time  `json:"fecha_registro,omitempty"`
	Segmento              *string    `json:"segmento"`
	NivelSostenibilidad   *string    `json:"nivel_sostenibilidad"`
	Localidad             *string    `json:"localidad"`
	Departamento          *string    `json:"departamento"`
	Provincia             *string    `json:"provincia"`
	EmailCuenta           *string    `json:"email_cuenta"`
	FechaUltimaEvaluacion *time.Time `json:"fecha_ultima_evaluacion"`
	ResponsableActivo     *string    `json:"responsable_activo"`
}

// ============================================
// MODELOS DE CUENTA
// ============================================

type TipoCuenta string

const (
	TipoCuentaBodega           TipoCuenta = "BODEGA"
	TipoCuentaAdministradorApp TipoCuenta = "ADMINISTRADOR_APP"
)

type Cuenta struct {
	ID            int        `json:"id_cuenta,omitempty"`
	Tipo          TipoCuenta `json:"tipo"`                // ENUM: BODEGA, ADMINISTRADOR_APP
	IDBodega      *int       `json:"id_bodega,omitempty"` // nullable, depende de tipo
	EmailLogin    string     `json:"email_login"`
	PasswordHash  string     `json:"-"`
	FechaRegistro time.Time  `json:"fecha_registro,omitempty"`
}

type CuentaRequest struct {
	EmailLogin string `json:"email_login"`
	Password   string `json:"password"`
}

// ============================================
// MODELOS DE RESPONSABLE
// ============================================

type Responsable struct {
	ID            int        `json:"id_responsable,omitempty"`
	IDCuenta      int        `json:"id_cuenta"`
	Nombre        string     `json:"nombre"`
	Apellido      string     `json:"apellido"`
	Cargo         string     `json:"cargo"`
	DNI           string     `json:"dni"` // varchar(8), check: ^[0-9]{7,8}$
	Activo        bool       `json:"activo"`
	FechaRegistro time.Time  `json:"fecha_registro,omitempty"`
	FechaBaja     *time.Time `json:"fecha_baja,omitempty"`
}

type ResponsableRequest struct {
	Nombre   string  `json:"nombre"`
	Apellido string  `json:"apellido"`
	Cargo    string  `json:"cargo"`
	DNI      *string `json:"dni,omitempty"`
}

type ResponsableUpdateDTO struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Cargo    string `json:"cargo"`
	DNI      string `json:"dni"`
}

type EmailUpdateDTO struct {
	Email string `json:"email"`
}

// ============================================
// MODELOS DE USUARIO
// ============================================

type Usuario struct {
	IdUsuario     int        `json:"idUsuario,omitempty" db:"id_usuario"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	Nombre        string     `json:"nombre" db:"nombre"`
	Apellido      string     `json:"apellido" db:"apellido"`
	Rol           string     `json:"rol" db:"rol"`
	Activo        bool       `json:"activo" db:"activo"`
	FechaRegistro time.Time  `json:"fecha_registro" db:"fecha_registro"`
	UltimoAcceso  *time.Time `json:"ultimo_acceso" db:"ultimo_acceso"`
}

type UsuarioDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Rol      string `json:"rol"`
}

type UsuarioLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ToPublic limpia campos sensibles antes de enviar al cliente
func (u *Usuario) ToPublic() *Usuario {
	publicUser := *u
	publicUser.PasswordHash = ""
	return &publicUser
}

// ============================================
// MODELOS DE UBICACIÓN
// ============================================

type Provincia struct {
	ID     int    `json:"id_provincia,omitempty"`
	Nombre string `json:"nombre"`
}

type Departamento struct {
	ID          int    `json:"id_departamento,omitempty"`
	IDProvincia int    `json:"id_provincia"`
	Nombre      string `json:"nombre"`
}

type Localidad struct {
	ID             int    `json:"id_localidad,omitempty"`
	IDDepartamento int    `json:"id_departamento"`
	Nombre         string `json:"nombre"`
}

// ============================================
// DTOs DE REGISTRO
// ============================================

type RegistroRequest struct {
	Bodega      BodegaRequest      `json:"bodega"`
	Cuenta      CuentaRequest      `json:"cuenta"`
	Responsable ResponsableRequest `json:"responsable"`
}

type RegistroResponse struct {
	IDBodega      int    `json:"id_bodega"`
	IDCuenta      int    `json:"id_cuenta"`
	IDResponsable int    `json:"id_responsable"`
	Mensaje       string `json:"mensaje"`
}

// ============================================
// MODELOS DE AUTOEVALUACIÓN
// ============================================

type EstadoAutoevaluacion string

const (
	EstadoPendiente  EstadoAutoevaluacion = "PENDIENTE"
	EstadoCompletada EstadoAutoevaluacion = "COMPLETADA"
	EstadoCancelada  EstadoAutoevaluacion = "CANCELADA"
)

type Segmento struct {
	ID          int    `json:"id_segmento"`
	Nombre      string `json:"nombre"`
	MinTuristas int    `json:"min_turistas"`
	MaxTuristas *int   `json:"max_turistas,omitempty"`
}

type NivelSostenibilidad struct {
	ID         int    `json:"id_nivel_sostenibilidad"`
	IDSegmento int    `json:"id_segmento"`
	Nombre     string `json:"nombre"`
	MinPuntaje int    `json:"min_puntaje"`
	MaxPuntaje int    `json:"max_puntaje"`
}

type Capitulo struct {
	ID          int    `json:"id_capitulo"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Orden       int    `json:"orden"`
}

type Indicador struct {
	ID          int    `json:"id_indicador"`
	IDCapitulo  int    `json:"id_capitulo"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Orden       int    `json:"orden"`
}

type IndicadorConHabilitacion struct {
	Indicador        *Indicador        `json:"indicador"`
	NivelesRespuesta []*NivelRespuesta `json:"niveles_respuesta"`
	Habilitado       bool              `json:"habilitado"`
}

type NivelRespuesta struct {
	ID          int    `json:"id_nivel_respuesta"`
	IDIndicador int    `json:"id_indicador"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Puntos      int    `json:"puntos"`
	Posicion    int    `json:"posicion"`
}

type Autoevaluacion struct {
	ID                    int                  `json:"id_autoevaluacion"`
	FechaInicio           time.Time            `json:"fecha_inicio"`
	FechaFin              *time.Time           `json:"fecha_fin,omitempty"`
	Estado                EstadoAutoevaluacion `json:"estado"`
	IDBodega              int                  `json:"id_bodega"`
	IDSegmento            *int                 `json:"id_segmento,omitempty"`
	PuntajeFinal          *int                 `json:"puntaje_final,omitempty"`
	IDNivelSostenibilidad *int                 `json:"id_nivel_sostenibilidad,omitempty"`
	EstadoEvidencia       *EstadoEvidencia     `json:"estado_evidencia,omitempty"`
	IDResponsable         *int                 `json:"id_responsable,omitempty"`
}

type Respuesta struct {
	ID               int `json:"id_respuesta"`
	IDNivelRespuesta int `json:"id_nivel_respuesta"`
	IDIndicador      int `json:"id_indicador"`
	IDAutoevaluacion int `json:"id_autoevaluacion"`
}

type EstructuraAutoevaluacion struct {
	Capitulos []*CapituloEstructura `json:"capitulos"`
}

type CapituloEstructura struct {
	Capitulo    *Capitulo                   `json:"capitulo"`
	Indicadores []*IndicadorConHabilitacion `json:"indicadores"`
}

// DTOs
type CreateAutoevaluacionRequest struct {
	IDBodega int `json:"id_bodega"`
}

type SeleccionarSegmentoRequest struct {
	IDSegmento int `json:"id_segmento"`
}

type GuardarRespuestaRequest struct {
	IDRespuesta      int `json:"id_respuesta,omitempty"`
	IDIndicador      int `json:"id_indicador"`
	IDNivelRespuesta int `json:"id_nivel_respuesta"`
}

type GuardarRespuestasRequest struct {
	Respuestas []GuardarRespuestaRequest `json:"respuestas"`
}

type AutoevaluacionPendienteResponse struct {
	AutoevaluacionPendiente *Autoevaluacion           `json:"autoevaluacion_pendiente,omitempty"`
	Respuestas              []GuardarRespuestaRequest `json:"respuestas,omitempty"`
	Mensaje                 string                    `json:"mensaje"`
}

// ============================================
// MODELOS DE EVIDENCIA
// ============================================

type EstadoEvidencia string

const (
	EstadoSinEvidencia EstadoEvidencia = "SIN_EVIDENCIA"
	EstadoParcial      EstadoEvidencia = "PARCIAL"
	EstadoCompleta     EstadoEvidencia = "COMPLETA"
)

type Evidencia struct {
	ID          int    `json:"id_evidencia"`
	IDRespuesta int    `json:"id_respuesta"`
	Nombre      string `json:"nombre"`    // Nombre generado: {original}_ae{id_auto}_r{id_resp}_{timestamp}.pdf
	Ubicacion   string `json:"ubicacion"` // Ruta completa: evidencias/{id_bodega}/{nombre}
}

type AgregarEvidenciaRequest struct {
	IDAutoevaluacion int `json:"id_autoevaluacion"`
	IDRespuesta      int `json:"id_respuesta"`
}

type ObtenerEvidenciaResponse struct {
	Evidencia *Evidencia `json:"evidencia,omitempty"`
	Mensaje   string     `json:"mensaje,omitempty"`
}

// ============================================
// MODELOS DE ADMINISTRACIÓN
// ============================================

type AdminStatsResponse struct {
	TotalBodegas            int                    `json:"totalBodegas"`
	EvaluacionesCompletadas int                    `json:"evaluacionesCompletadas"`
	PromedioSostenibilidad  float64                `json:"promedioSostenibilidad"`
	NivelPromedio           string                 `json:"nivelPromedio"`
	DistribucionNiveles     DistribucionNiveles    `json:"distribucionNiveles"`
	DistribucionPorSegmento []SegmentoDistribucion `json:"distribucionPorSegmento"`
}

type DistribucionNiveles struct {
	Minimo int `json:"minimo"`
	Medio  int `json:"medio"`
	Alto   int `json:"alto"`
}

type SegmentoDistribucion struct {
	IDSegmento     int    `json:"id_segmento"`
	NombreSegmento string `json:"nombre_segmento"`
	Alto           int    `json:"alto"`
	Medio          int    `json:"medio"`
	Minimo         int    `json:"minimo"`
}

type EvaluacionListItem struct {
	IDAutoevaluacion int     `json:"id_autoevaluacion"`
	IDBodega         int     `json:"id_bodega"`
	NombreBodega     string  `json:"nombre_bodega"`
	RazonSocial      string  `json:"razon_social"`
	Estado           string  `json:"estado"`
	FechaInicio      string  `json:"fecha_inicio"`
	FechaFin         *string `json:"fecha_fin"`
	Responsable      string  `json:"responsable"`
}

// ============================================
// MODELOS DE HISTORIAL Y RESULTADOS
// ============================================

type HistorialItemResponse struct {
	IDAutoevaluacion       int                      `json:"id_autoevaluacion"`
	FechaInicio            string                   `json:"fecha_inicio"`
	FechaFinalizacion      string                   `json:"fecha_finalizacion"`
	Estado                 string                   `json:"estado"`
	IDBodega               int                      `json:"id_bodega"`
	IDSegmento             *int                     `json:"id_segmento"`
	NombreSegmento         string                   `json:"nombre_segmento,omitempty"`
	PuntajeFinal           *int                     `json:"puntaje_final"`
	PuntajeMaximo          int                      `json:"puntaje_maximo"`
	Porcentaje             float64                  `json:"porcentaje"`
	IDNivelSostenibilidad  *int                     `json:"id_nivel_sostenibilidad"`
	NivelSostenibilidad    *NivelSostenibilidadInfo `json:"nivel_sostenibilidad,omitempty"`
	IndicadoresRespondidos int                      `json:"indicadores_respondidos"`
	IndicadoresTotal       int                      `json:"indicadores_total"`
}

type NivelSostenibilidadInfo struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion,omitempty"`
}

type ResponsableInfo struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Cargo    string `json:"cargo"`
	DNI      string `json:"dni"`
}

type ResultadoDetalladoResponse struct {
	Autoevaluacion HistorialItemResponse        `json:"autoevaluacion"`
	Capitulos      []ResultadoCapituloDetallado `json:"capitulos"`
	Responsable    *ResponsableInfo             `json:"responsable,omitempty"`
}

type ResultadoCapituloDetallado struct {
	IDCapitulo             int                         `json:"id_capitulo"`
	Nombre                 string                      `json:"nombre"`
	PuntajeObtenido        int                         `json:"puntaje_obtenido"`
	PuntajeMaximo          int                         `json:"puntaje_maximo"`
	Porcentaje             float64                     `json:"porcentaje"`
	IndicadoresCompletados int                         `json:"indicadores_completados"`
	IndicadoresTotal       int                         `json:"indicadores_total"`
	Indicadores            []ResultadoIndicadorDetalle `json:"indicadores"`
}

type ResultadoIndicadorDetalle struct {
	IDIndicador          int    `json:"id_indicador"`
	Nombre               string `json:"nombre"`
	Descripcion          string `json:"descripcion"`
	Orden                int    `json:"orden"`
	RespuestaNombre      string `json:"respuesta_nombre"`
	RespuestaDescripcion string `json:"respuesta_descripcion"`
	RespuestaPuntos      int    `json:"respuesta_puntos"`
	PuntajeMaximo        int    `json:"puntaje_maximo"`
	IDRespuesta          *int   `json:"id_respuesta,omitempty"`
	TieneEvidencia       bool   `json:"tiene_evidencia"`
}

// ============================================
// MODELOS PARA RESULTADOS DE BODEGA
// ============================================

type ResultadosBodegaResponse struct {
	Bodega               BodegaResumen             `json:"bodega"`
	Autoevaluacion       AutoevaluacionResumen     `json:"autoevaluacion"`
	Segmento             SegmentoResumen           `json:"segmento"`
	NivelSustentabilidad NivelResumen              `json:"nivel_sustentabilidad"`
	Capitulos            []CapituloResultadoSimple `json:"capitulos"`
}

type BodegaResumen struct {
	NombreFantasia string `json:"nombre_fantasia"`
}

type AutoevaluacionResumen struct {
	FechaFin     string `json:"fecha_fin"`
	PuntajeFinal int    `json:"puntaje_final"`
}

type SegmentoResumen struct {
	Nombre string `json:"nombre"`
}

type NivelResumen struct {
	Nombre string `json:"nombre"`
}

type CapituloResultadoSimple struct {
	Nombre      string                     `json:"nombre"`
	Orden       int                        `json:"orden"`
	Indicadores []IndicadorResultadoSimple `json:"indicadores"`
}

type IndicadorResultadoSimple struct {
	Nombre           string                    `json:"nombre"`
	Descripcion      string                    `json:"descripcion"`
	Orden            int                       `json:"orden"`
	NivelesRespuesta []NivelRespuestaResultado `json:"niveles_respuesta"`
}

type NivelRespuestaResultado struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Puntos      int    `json:"puntos"`
}
