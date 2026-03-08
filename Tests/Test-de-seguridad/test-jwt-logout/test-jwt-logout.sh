#!/bin/bash
# Test de invalidación de JWT en logout 
#
# Verifica que tras el logout el token queda inutilizable en el servidor,
# incluso si alguien conserva el valor de la cookie.
#
# Requiere credenciales válidas de una cuenta BODEGA.

BACKEND_URL="http://localhost:8080"

# ── CONFIGURAR ESTOS VALORES ────────────────────────────────────────────────
EMAIL="carli@mail.com"
PASSWORD="Carlitos123"
# ────────────────────────────────────────────────────────────────────────────

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "=============================================="
echo " Test Invalidación JWT en Logout"
echo "=============================================="
echo ""

if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
  echo -e "${RED}✗ Configurar EMAIL y PASSWORD antes de ejecutar${NC}"
  exit 1
fi

# Verificar backend
echo -n "Backend activo... "
CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$BACKEND_URL/health")
if [ "$CODE" = "000" ]; then
  echo -e "${RED}✗ No responde${NC}"; exit 1
fi
echo -e "${GREEN}OK${NC}"
echo ""

# ── TEST 1: Login y capturar token ───────────────────────────────────────────
echo -e "${YELLOW}[TEST 1]${NC} Login y capturar auth_token"

LOGIN_RESP=$(curl -s -c /tmp/coviar_cookies.txt -w "\n%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

LOGIN_CODE=$(echo "$LOGIN_RESP" | tail -1)

if [ "$LOGIN_CODE" != "200" ]; then
  echo -e "  ${RED}✗ Login fallido — HTTP $LOGIN_CODE${NC}"
  exit 1
fi

# Extraer token de la cookie
TOKEN=$(grep "auth_token" /tmp/coviar_cookies.txt | awk '{print $7}')

if [ -z "$TOKEN" ]; then
  echo -e "  ${RED}✗ No se pudo capturar el auth_token${NC}"
  exit 1
fi
echo -e "  ${GREEN}✓ Login exitoso — token capturado (${#TOKEN} caracteres)${NC}"
echo ""

# ── TEST 2: Verificar que el token es aceptado antes del logout ──────────────
# El endpoint puede devolver 200, 400 u otro código por lógica de negocio.
# Lo importante es que NO devuelva 401 (token rechazado).
echo -e "${YELLOW}[TEST 2]${NC} Token aceptado antes del logout             → no 401"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  "$BACKEND_URL/api/autoevaluaciones/historial" \
  -H "Cookie: auth_token=$TOKEN")

if [ "$CODE" != "401" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — HTTP $CODE (token aceptado por el middleware)"
else
  echo -e "  ${RED}✗ FALLO${NC} — HTTP 401: token rechazado antes del logout"
fi
echo ""

# ── TEST 3: Hacer logout ──────────────────────────────────────────────────────
echo -e "${YELLOW}[TEST 3]${NC} Hacer logout"

LOGOUT_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/logout" \
  -H "Cookie: auth_token=$TOKEN")

if [ "$LOGOUT_CODE" = "200" ]; then
  echo -e "  ${GREEN}✓ Logout exitoso — HTTP $LOGOUT_CODE${NC}"
else
  echo -e "  ${RED}✗ Logout fallido — HTTP $LOGOUT_CODE${NC}"
fi
echo ""

# ── TEST 4: Usar el mismo token tras el logout → debe ser rechazado ───────────
echo -e "${YELLOW}[TEST 4]${NC} Mismo token reutilizado post-logout         → rechazar (401)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  "$BACKEND_URL/api/autoevaluaciones/historial" \
  -H "Cookie: auth_token=$TOKEN")

if [ "$CODE" = "401" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — HTTP $CODE (token revocado correctamente)"
else
  echo -e "  ${RED}✗ FALLO${NC} — Se esperaba HTTP 401, se recibió HTTP $CODE"
  echo "     El token sigue siendo válido después del logout — invalidación no funciona"
fi
echo ""

# ── TEST 5: Nuevo login genera token fresco que es aceptado por el middleware ─
echo -e "${YELLOW}[TEST 5]${NC} Nuevo token post-logout                     → aceptado (no 401)"

curl -s -c /tmp/coviar_cookies2.txt -o /dev/null --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"

TOKEN2=$(grep "auth_token" /tmp/coviar_cookies2.txt | awk '{print $7}')

if [ -z "$TOKEN2" ]; then
  echo -e "  ${RED}✗ No se pudo obtener nuevo token${NC}"
else
  CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
    "$BACKEND_URL/api/autoevaluaciones/historial" \
    -H "Cookie: auth_token=$TOKEN2")

  if [ "$CODE" != "401" ]; then
    echo -e "  ${GREEN}✓ CORRECTO${NC} — HTTP $CODE (nuevo token aceptado)"
  else
    echo -e "  ${RED}✗ FALLO${NC} — HTTP 401: nuevo token rechazado incorrectamente"
  fi
fi
echo ""

# Limpiar archivos temporales
rm -f /tmp/coviar_cookies.txt /tmp/coviar_cookies2.txt

echo "=============================================="
echo " Tests completados"
echo "=============================================="
