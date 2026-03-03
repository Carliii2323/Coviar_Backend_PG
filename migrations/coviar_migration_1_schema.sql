-- ============================================================
-- COVIAR 3.0 - MIGRACIÓN A POSTGRESQL LOCAL
-- Archivo 1 de 2: ESTRUCTURA (Schema)
-- Ejecutar primero, como usuario coviar_user o postgres
-- Base de datos destino: coviar_db
-- ============================================================

-- Extensiones
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- TIPOS ENUMERADOS
-- ============================================================
CREATE TYPE estado_autoevaluacion  AS ENUM ('PENDIENTE', 'COMPLETADA', 'CANCELADA');
CREATE TYPE tipo_cuenta             AS ENUM ('BODEGA', 'ADMINISTRADOR_APP');
CREATE TYPE estado_evidencia_tipo   AS ENUM ('COMPLETA', 'PARCIAL', 'SIN_EVIDENCIA');

-- Secuencia para restaurar_contrasenas (replica nombre original de Supabase)
CREATE SEQUENCE IF NOT EXISTS password_reset_tokens_id_seq;

-- ============================================================
-- TABLAS (orden respeta dependencias de claves foráneas)
-- ============================================================

-- 1. Provincias
CREATE TABLE IF NOT EXISTS provincias (
    id_provincia  INTEGER GENERATED ALWAYS AS IDENTITY,
    nombre        TEXT NOT NULL UNIQUE,
    CONSTRAINT provincias_pkey PRIMARY KEY (id_provincia)
);

-- 2. Departamentos
CREATE TABLE IF NOT EXISTS departamentos (
    id_departamento INTEGER GENERATED ALWAYS AS IDENTITY,
    nombre          TEXT NOT NULL,
    id_provincia    INTEGER NOT NULL,
    CONSTRAINT departamentos_pkey    PRIMARY KEY (id_departamento),
    CONSTRAINT departamentos_prov_fk FOREIGN KEY (id_provincia)
        REFERENCES provincias(id_provincia)
);

-- 3. Localidades
CREATE TABLE IF NOT EXISTS localidades (
    id_localidad    INTEGER GENERATED ALWAYS AS IDENTITY,
    nombre          TEXT NOT NULL,
    id_departamento INTEGER NOT NULL,
    CONSTRAINT localidades_pkey     PRIMARY KEY (id_localidad),
    CONSTRAINT localidades_depto_fk FOREIGN KEY (id_departamento)
        REFERENCES departamentos(id_departamento)
);

-- 4. Segmentos
CREATE TABLE IF NOT EXISTS segmentos (
    id_segmento  INTEGER GENERATED ALWAYS AS IDENTITY,
    nombre       TEXT NOT NULL,
    min_turistas INTEGER NOT NULL,
    max_turistas INTEGER,
    CONSTRAINT segmentos_pkey PRIMARY KEY (id_segmento)
);

-- 5. Niveles de sostenibilidad
CREATE TABLE IF NOT EXISTS niveles_sostenibilidad (
    id_nivel_sostenibilidad INTEGER GENERATED ALWAYS AS IDENTITY,
    id_segmento             INTEGER NOT NULL,
    nombre                  TEXT NOT NULL,
    min_puntaje             INTEGER NOT NULL,
    max_puntaje             INTEGER NOT NULL,
    CONSTRAINT niveles_sostenibilidad_pkey        PRIMARY KEY (id_nivel_sostenibilidad),
    CONSTRAINT niveles_sostenibilidad_segmento_fk FOREIGN KEY (id_segmento)
        REFERENCES segmentos(id_segmento)
);

-- 6. Capítulos
CREATE TABLE IF NOT EXISTS capitulos (
    id_capitulo INTEGER GENERATED ALWAYS AS IDENTITY,
    nombre      TEXT NOT NULL,
    descripcion TEXT NOT NULL,
    orden       INTEGER NOT NULL,
    CONSTRAINT capitulos_pkey PRIMARY KEY (id_capitulo)
);

-- 7. Indicadores
CREATE TABLE IF NOT EXISTS indicadores (
    id_indicador INTEGER GENERATED ALWAYS AS IDENTITY,
    id_capitulo  INTEGER NOT NULL,
    nombre       TEXT NOT NULL,
    descripcion  TEXT NOT NULL,
    orden        INTEGER NOT NULL,
    CONSTRAINT indicadores_pkey       PRIMARY KEY (id_indicador),
    CONSTRAINT indicadores_capitulo_fk FOREIGN KEY (id_capitulo)
        REFERENCES capitulos(id_capitulo)
);

-- 8. Niveles de respuesta
CREATE TABLE IF NOT EXISTS niveles_respuesta (
    id_nivel_respuesta INTEGER GENERATED ALWAYS AS IDENTITY,
    id_indicador       INTEGER NOT NULL,
    nombre             TEXT NOT NULL,
    descripcion        TEXT,
    puntos             INTEGER NOT NULL,
    posicion           INTEGER,
    CONSTRAINT niveles_respuesta_pkey        PRIMARY KEY (id_nivel_respuesta),
    CONSTRAINT niveles_respuesta_indicador_fk FOREIGN KEY (id_indicador)
        REFERENCES indicadores(id_indicador)
);

-- 9. Bodegas
CREATE TABLE IF NOT EXISTS bodegas (
    id_bodega           INTEGER GENERATED ALWAYS AS IDENTITY,
    razon_social        TEXT NOT NULL,
    nombre_fantasia     TEXT NOT NULL,
    cuit                CHAR(11) NOT NULL CHECK (cuit ~ '^[0-9]{11}$'),
    inv_bod             CHAR(6),
    inv_vin             CHAR(6),
    calle               TEXT NOT NULL,
    numeracion          TEXT NOT NULL DEFAULT 'S/N',
    id_localidad        INTEGER NOT NULL,
    telefono            TEXT NOT NULL CHECK (telefono ~ '^[0-9]+$'),
    email_institucional TEXT NOT NULL CHECK (email_institucional LIKE '%@%'),
    fecha_registro      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bodegas_pkey         PRIMARY KEY (id_bodega),
    CONSTRAINT bodegas_localidad_fk FOREIGN KEY (id_localidad)
        REFERENCES localidades(id_localidad)
);

-- 10. Cuentas
CREATE TABLE IF NOT EXISTS cuentas (
    id_cuenta      INTEGER GENERATED ALWAYS AS IDENTITY,
    tipo           tipo_cuenta NOT NULL,
    id_bodega      INTEGER UNIQUE,
    email_login    VARCHAR NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    fecha_registro TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT cuentas_pkey      PRIMARY KEY (id_cuenta),
    CONSTRAINT cuentas_bodega_fk FOREIGN KEY (id_bodega)
        REFERENCES bodegas(id_bodega)
);

-- 11. Responsables
CREATE TABLE IF NOT EXISTS responsables (
    id_responsable INTEGER GENERATED ALWAYS AS IDENTITY,
    id_cuenta      INTEGER NOT NULL,
    nombre         TEXT NOT NULL,
    apellido       TEXT NOT NULL,
    cargo          TEXT NOT NULL,
    dni            VARCHAR NOT NULL CHECK (dni ~ '^[0-9]{7,8}$'),
    activo         BOOLEAN NOT NULL DEFAULT TRUE,
    fecha_registro TIMESTAMPTZ NOT NULL DEFAULT now(),
    fecha_baja     TIMESTAMPTZ,
    CONSTRAINT responsables_pkey     PRIMARY KEY (id_responsable),
    CONSTRAINT responsables_cuenta_fk FOREIGN KEY (id_cuenta)
        REFERENCES cuentas(id_cuenta)
);

-- 12. Autoevaluaciones
CREATE TABLE IF NOT EXISTS autoevaluaciones (
    id_autoevaluacion       INTEGER GENERATED ALWAYS AS IDENTITY,
    fecha_inicio            TIMESTAMPTZ NOT NULL DEFAULT now(),
    fecha_fin               TIMESTAMPTZ,
    estado                  estado_autoevaluacion NOT NULL DEFAULT 'PENDIENTE',
    id_bodega               INTEGER NOT NULL,
    id_segmento             INTEGER,
    puntaje_final           INTEGER,
    id_nivel_sostenibilidad INTEGER,
    estado_evidencia        estado_evidencia_tipo,
    CONSTRAINT autoevaluaciones_pkey          PRIMARY KEY (id_autoevaluacion),
    CONSTRAINT autoevaluaciones_bodega_fk     FOREIGN KEY (id_bodega)
        REFERENCES bodegas(id_bodega),
    CONSTRAINT autoevaluaciones_segmento_fk   FOREIGN KEY (id_segmento)
        REFERENCES segmentos(id_segmento),
    CONSTRAINT autoevaluaciones_nivel_sost_fk FOREIGN KEY (id_nivel_sostenibilidad)
        REFERENCES niveles_sostenibilidad(id_nivel_sostenibilidad)
);

-- 13. Respuestas
CREATE TABLE IF NOT EXISTS respuestas (
    id_respuesta       INTEGER GENERATED ALWAYS AS IDENTITY,
    id_nivel_respuesta INTEGER NOT NULL,
    id_indicador       INTEGER NOT NULL,
    id_autoevaluacion  INTEGER NOT NULL,
    CONSTRAINT respuestas_pkey                              PRIMARY KEY (id_respuesta),
    CONSTRAINT respuestas_autoevaluacion_indicador_unique   UNIQUE (id_autoevaluacion, id_indicador),
    CONSTRAINT respuestas_nivel_respuesta_fk FOREIGN KEY (id_nivel_respuesta)
        REFERENCES niveles_respuesta(id_nivel_respuesta),
    CONSTRAINT respuestas_indicador_fk       FOREIGN KEY (id_indicador)
        REFERENCES indicadores(id_indicador),
    CONSTRAINT respuestas_autoevaluacion_fk  FOREIGN KEY (id_autoevaluacion)
        REFERENCES autoevaluaciones(id_autoevaluacion)
);

-- 14. Segmento-Indicador (tabla de relación, sin identity)
CREATE TABLE IF NOT EXISTS segmento_indicador (
    id_segmento  INTEGER NOT NULL,
    id_indicador INTEGER NOT NULL,
    CONSTRAINT segmento_indicador_pkey         PRIMARY KEY (id_segmento, id_indicador),
    CONSTRAINT segmento_indicador_segmento_fk  FOREIGN KEY (id_segmento)
        REFERENCES segmentos(id_segmento),
    CONSTRAINT segmento_indicador_indicador_fk FOREIGN KEY (id_indicador)
        REFERENCES indicadores(id_indicador)
);

-- 15. Evidencias
CREATE TABLE IF NOT EXISTS evidencias (
    id_evidencia INTEGER GENERATED ALWAYS AS IDENTITY,
    id_respuesta INTEGER NOT NULL,
    nombre       TEXT NOT NULL,
    ubicacion    TEXT NOT NULL,
    CONSTRAINT evidencias_pkey        PRIMARY KEY (id_evidencia),
    CONSTRAINT evidencias_respuesta_fk FOREIGN KEY (id_respuesta)
        REFERENCES respuestas(id_respuesta)
);

-- 16. Restaurar contraseñas
CREATE TABLE IF NOT EXISTS restaurar_contrasenas (
    id         INTEGER NOT NULL DEFAULT nextval('password_reset_tokens_id_seq'),
    user_id    INTEGER NOT NULL,
    token      VARCHAR NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT restaurar_contrasenas_pkey    PRIMARY KEY (id),
    CONSTRAINT restaurar_contrasenas_user_fk FOREIGN KEY (user_id)
        REFERENCES cuentas(id_cuenta)
);

-- ============================================================
-- ÍNDICES
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_autoevaluaciones_bodega        ON autoevaluaciones(id_bodega);
CREATE INDEX IF NOT EXISTS idx_autoevaluaciones_segmento      ON autoevaluaciones(id_segmento);
CREATE INDEX IF NOT EXISTS idx_autoevaluaciones_estado        ON autoevaluaciones(estado);
CREATE INDEX IF NOT EXISTS idx_respuestas_autoevaluacion      ON respuestas(id_autoevaluacion);
CREATE INDEX IF NOT EXISTS idx_indicadores_capitulo           ON indicadores(id_capitulo);
CREATE INDEX IF NOT EXISTS idx_niveles_respuesta_indicador    ON niveles_respuesta(id_indicador);

-- Solo un responsable activo por cuenta
CREATE UNIQUE INDEX IF NOT EXISTS un_responsable_activo_por_cuenta
    ON responsables(id_cuenta) WHERE activo = TRUE;
