# Tests de Seguridad — CoviarPG

Cada script prueba una vulnerabilidad específica del backend. Todos los tests usan `curl` y deben ejecutarse con el servidor corriendo en `http://localhost:8080`.

---

## Validación de archivos PDF

### ¿Cuál es la vulnerabilidad?
Si el backend acepta cualquier archivo que tenga extensión `.pdf` sin verificar su contenido real, un atacante puede subir un ejecutable, script malicioso o archivo ZIP renombrado como `.pdf`. Esto puede llevar a ejecución de código en el servidor o exposición de archivos internos.

La protección correcta es validar los **magic bytes** del archivo (los primeros bytes del contenido), no solo la extensión.

### ¿Cómo se hace el test?
El script crea 4 archivos de prueba y los intenta subir:

| Archivo | Contenido real | Esperado |
|---|---|---|
| `fake_exe.pdf` | Header EXE (`MZ...`) | ❌ Rechazado |
| `fake_text.pdf` | Texto plano | ❌ Rechazado |
| `fake_zip.pdf` | Header ZIP (`PK`) | ❌ Rechazado |
| `real.pdf` | PDF real (`%PDF-`) | ✅ Aceptado |
| PDF > 2MB | PDF válido pero grande | ❌ Rechazado |

```bash
cd Backend/Tests/Test-de-seguridad/test-PDF
# Configurar TOKEN y ID_AUTOEVALUACION en el script
bash test-validacion-pdf.sh
```

---

## Protección contra fuerza bruta en login

### ¿Cuál es la vulnerabilidad?
Sin límite de intentos, un atacante puede probar millones de combinaciones de contraseña contra un usuario conocido (ataque de diccionario o fuerza bruta). Con suficiente tiempo, puede adivinar la contraseña correcta sin que el sistema lo detecte.

La protección es bloquear temporalmente la IP después de N intentos fallidos consecutivos.

### ¿Cómo se hace el test?
El script hace 6 intentos de login con credenciales incorrectas al mismo endpoint y verifica que el 6° intento devuelva HTTP `429 Too Many Requests`.

```bash
cd Backend/Tests/Test-de-seguridad/test-brute-force
# Configurar EMAIL en el script (cuenta existente en la BD)
bash test-brute-force.sh
```

**Resultado esperado:** Los primeros intentos devuelven `401`, el 6° devuelve `429`.

---

## Ownership de Bodega (IDOR)

### ¿Cuál es la vulnerabilidad?
**IDOR** (Insecure Direct Object Reference): si el backend no verifica que el usuario autenticado es dueño del recurso que está accediendo, una bodega podría leer, modificar o cancelar las autoevaluaciones de otra bodega simplemente conociendo su ID.

Ejemplo: Bodega B hace `GET /api/autoevaluaciones/historial?id_bodega=22` para ver los datos de Bodega A.

### ¿Cómo se hace el test?
El script usa dos tokens distintos (Bodega A y Bodega B) e intenta que Bodega B acceda a datos de Bodega A.

```bash
cd Backend/Tests/Test-de-seguridad/test-ownership-bodega
# Configurar TOKEN_BODEGA_A, TOKEN_BODEGA_B, ID_BODEGA_A en el script
bash test-ownership-bodega.sh
```

**Resultado esperado:** Bodega A puede acceder a sus propios datos (`200`). Bodega B recibe `403 Forbidden` al intentar acceder a datos de Bodega A.

---

## Rate Limiting global y de escritura

### ¿Cuál es la vulnerabilidad?
Sin límite de requests, un atacante puede saturar el servidor con miles de peticiones (DoS), o abusar de endpoints de escritura para insertar datos masivamente (spam de autoevaluaciones, sobrecarga de BD).

El sistema tiene dos capas de protección:
- **Límite de escritura:** 20 POST/PUT/DELETE por minuto por IP
- **Límite global:** 200 requests por minuto por IP

### ¿Cómo se hace el test?
El script envía 25 requests POST seguidos para superar el límite de escritura (espera `429` en el request 21+), luego espera 65 segundos para que la ventana se resetee y envía 210 GET para superar el límite global.

```bash
cd Backend/Tests/Test-de-seguridad/test-rate-limit
# Configurar TOKEN en el script
bash test-rate-limit.sh
# ⚠️ Demora ~2 minutos por el reseteo de ventana
```

**Resultado esperado:** HTTP `429` al superar ambos límites.

---

## Control de acceso por rol (Admin vs Bodega)

### ¿Cuál es la vulnerabilidad?
Si los endpoints de administración no verifican el rol del JWT, cualquier usuario autenticado (aunque sea una bodega normal) puede acceder al panel de administración, ver todas las bodegas, eliminar cuentas, o ver reportes de otras bodegas.

### ¿Cómo se hace el test?
El script prueba 7 escenarios con tres estados: sin token, con token de BODEGA, y con token de ADMINISTRADOR_APP.

| Escenario | Token | Endpoint | Esperado |
|---|---|---|---|
| Sin autenticación | Ninguno | `GET /api/admin/evaluaciones` | `401` |
| Bodega accede a admin | BODEGA | `GET /api/admin/evaluaciones` | `403` |
| Bodega accede a admin | BODEGA | `GET /api/admin/bodegas` | `403` |
| Admin accede a admin | ADMIN | `GET /api/admin/evaluaciones` | `200` |
| Admin accede a admin | ADMIN | `GET /api/admin/bodegas` | `200` |

```bash
cd Backend/Tests/Test-de-seguridad/test-admin-rol
# Configurar TOKEN_BODEGA y TOKEN_ADMIN en el script
bash test-admin-rol.sh
```

---

## Ownership de Evidencias (IDOR en archivos)

### ¿Cuál es la vulnerabilidad?
Similar al test de Ownership de Bodega pero sobre archivos de evidencia. Si el backend no verifica que la evidencia pertenece a la autoevaluación del usuario autenticado, una bodega puede descargar, ver o eliminar los archivos adjuntos de otra bodega conociendo el ID de la evidencia.

Esto es especialmente grave porque las evidencias pueden contener documentos confidenciales de la bodega.

### ¿Cómo se hace el test?
El script usa Token A (dueño legítimo) y Token B (atacante) sobre los mismos IDs de evidencia.

```bash
cd Backend/Tests/Test-de-seguridad/test-evidencia-ownership
# Configurar TOKEN_A, TOKEN_B, ID_AUTOEVALUACION_A, ID_INDICADOR, ID_EVIDENCIA en el script
bash test-evidencia-ownership.sh
```

**Resultado esperado:** Token A puede operar sobre sus evidencias (`200`). Token B recibe `403` en todas las operaciones sobre evidencias ajenas.

---

## Protección CSRF

### ¿Cuál es la vulnerabilidad?
**CSRF** (Cross-Site Request Forgery): un sitio malicioso puede hacer que el navegador de un usuario autenticado envíe requests al backend sin que el usuario lo sepa, aprovechando que las cookies se envían automáticamente.

La protección es validar que el header `Origin` o `Referer` corresponda al dominio legítimo de la aplicación.

### ¿Cómo se hace el test?
El script prueba 8 combinaciones de Origin/Referer en requests de escritura (POST/PUT/DELETE) y verifica que solo los orígenes legítimos (`localhost:3000`) sean aceptados.

| Escenario | Origin | Esperado |
|---|---|---|
| Origen malicioso | `https://evil.com` | `403` |
| Origen malicioso en PUT | `https://attacker.io` | `403` |
| Sin Origin header (curl/Postman) | Ninguno | `200` (permitido) |
| Origen legítimo | `http://localhost:3000` | `200` |
| GET con origen malicioso | `https://evil.com` | `200` (GET no modifica) |

```bash
cd Backend/Tests/Test-de-seguridad/test-csrf
# Configurar TOKEN en el script
bash test-csrf.sh
```

---

## Invalidación de JWT en logout

### ¿Cuál es la vulnerabilidad?
Los JWT son stateless por diseño — si el servidor no mantiene una lista negra, un token válido sigue funcionando hasta que expire aunque el usuario haya cerrado sesión. Si alguien roba el token de la cookie antes del logout, puede seguir usándolo por hasta 24 horas.

La protección es guardar tokens revocados en una **blacklist** en BD al hacer logout.

### ¿Cómo se hace el test?
El script sigue un flujo completo de 5 pasos:

1. Login → captura el token de la cookie
2. Verifica que el token funciona (`200`)
3. Hace logout
4. Intenta usar el mismo token → debe fallar (`401`)
5. Hace login de nuevo → el nuevo token debe funcionar (`200`)

```bash
cd Backend/Tests/Test-de-seguridad/test-jwt-logout
# Configurar EMAIL y PASSWORD en el script
bash test-jwt-logout.sh
```

**Resultado esperado:** El token pre-logout devuelve `401` después del logout. Un token nuevo sí funciona.

---

## Cómo ejecutar todos los tests

```bash
# Asegurarse que el backend está corriendo
cd Backend && go run ./cmd/api

# En otra terminal, ejecutar cada test individualmente
# (configurar los tokens/IDs en cada script primero)
bash Tests/Test-de-seguridad/test-brute-force/test-brute-force.sh
bash Tests/Test-de-seguridad/test-csrf/test-csrf.sh
bash Tests/Test-de-seguridad/test-admin-rol/test-admin-rol.sh
bash Tests/Test-de-seguridad/test-ownership-bodega/test-ownership-bodega.sh
bash Tests/Test-de-seguridad/test-evidencia-ownership/test-evidencia-ownership.sh
bash Tests/Test-de-seguridad/test-jwt-logout/test-jwt-logout.sh
bash Tests/Test-de-seguridad/test-PDF/test-validacion-pdf.sh
bash Tests/Test-de-seguridad/test-rate-limit/test-rate-limit.sh
```

> Los tokens JWT deben reemplazarse con valores válidos obtenidos del navegador (DevTools → Application → Cookies → `auth_token`) o del response de login.
