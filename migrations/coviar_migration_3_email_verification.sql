-- ============================================================
-- MIGRACIÓN 3: Verificación de correo electrónico
-- ============================================================

-- 1. Columna email_verificado en cuentas
ALTER TABLE cuentas
    ADD COLUMN IF NOT EXISTS email_verificado BOOLEAN NOT NULL DEFAULT FALSE;

-- 2. Tabla para almacenar códigos de verificación de 6 dígitos
--
--    intentos_fallidos: se incrementa por cada código incorrecto ingresado.
--    Al llegar a 20 el código queda bloqueado y el usuario debe solicitar uno nuevo.
--    El contador se resetea automáticamente al generar un código nuevo (DELETE + INSERT).
CREATE TABLE IF NOT EXISTS verificacion_correo (
    id                SERIAL,
    cuenta_id         INTEGER      NOT NULL,
    codigo            CHAR(6)      NOT NULL CHECK (codigo ~ '^[0-9]{6}$'),
    expires_at        TIMESTAMPTZ  NOT NULL,
    verificado        BOOLEAN      NOT NULL DEFAULT FALSE,
    intentos_fallidos INTEGER      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT verificacion_correo_pkey       PRIMARY KEY (id),
    CONSTRAINT verificacion_correo_cuenta_fk  FOREIGN KEY (cuenta_id)
        REFERENCES cuentas(id_cuenta) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_verificacion_correo_cuenta
    ON verificacion_correo(cuenta_id);

-- 3. Tabla para control de reenvíos con esperas progresivas
--
--    Secuencia de esperas entre reenvíos:
--      1er reenvío → espera 1 min antes del 2do
--      2do         → espera 3 min antes del 3ro
--      3ro         → espera 5 min antes del 4to
--      4to         → espera 10 min antes del 5to
--      5to         → espera 30 min antes del 6to
--      6to intento → bloqueado 24 horas
--
CREATE TABLE IF NOT EXISTS verificacion_reenvios (
    cuenta_id         INTEGER      NOT NULL,
    intentos          INTEGER      NOT NULL DEFAULT 0,
    proximo_reenvio   TIMESTAMPTZ,          -- NULL = sin restricción activa
    bloqueado_hasta   TIMESTAMPTZ,          -- NULL = no bloqueado

    CONSTRAINT verificacion_reenvios_pkey       PRIMARY KEY (cuenta_id),
    CONSTRAINT verificacion_reenvios_cuenta_fk  FOREIGN KEY (cuenta_id)
        REFERENCES cuentas(id_cuenta) ON DELETE CASCADE
);
