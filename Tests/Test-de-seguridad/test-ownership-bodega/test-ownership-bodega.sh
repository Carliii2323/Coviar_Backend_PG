#!/bin/bash
# Test de verificación de ownership en endpoints de bodega 
#
# Requiere DOS tokens:
#   TOKEN_A: token de una bodega (ej: bodega ID 1)
#   TOKEN_B: token de OTRA bodega distinta (ej: bodega ID 2)
#
# Para obtener los tokens:
#   1. Logueate con cada cuenta en la app
#   2. DevTools (F12) → Application → Cookies → copiá auth_token

BACKEND_URL="http://localhost:8080"

# ── CONFIGURAR ESTOS VALORES ────────────────────────────────────────────────
TOKEN_A="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJjYXJsaUBtYWlsLmNvbSIsInRpcG9fY3VlbnRhIjoiQk9ERUdBIiwiYm9kZWdhX2lkIjoyNCwiaXNzIjoiY292aWFyLWJhY2tlbmQiLCJleHAiOjE3NzI4OTM0MjAsIm5iZiI6MTc3MjgwNzAyMCwiaWF0IjoxNzcyODA3MDIwfQ.qCduEbnRciYHgfKpA1h4V6jVbRycwwFj1Obx7iJrOXw"   # token de la bodega propietaria
TOKEN_B="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNiwiZW1haWwiOiJkaWF6dHV0aW45MEBnbWFpbC5jb20iLCJ0aXBvX2N1ZW50YSI6IkJPREVHQSIsImJvZGVnYV9pZCI6MjYsImlzcyI6ImNvdmlhci1iYWNrZW5kIiwiZXhwIjoxNzcyODkzMzk2LCJuYmYiOjE3NzI4MDY5OTYsImlhdCI6MTc3MjgwNjk5Nn0.bkyFdnz0Q16A22h_yQAKJ21ec8ZafsPbNlc2WhvNktk"   # token de OTRA bodega

ID_BODEGA_A=24   # ID de la bodega a la que pertenece TOKEN_A
# ────────────────────────────────────────────────────────────────────────────

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "=============================================="
echo " Test Ownership Bodega"
echo " Bodega bajo prueba: ID=$ID_BODEGA_A"
echo "=============================================="
echo ""

# Verificar backend
echo -n "Backend activo... "
CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$BACKEND_URL/")
if [ "$CODE" = "000" ]; then
  echo -e "${RED}✗ No responde${NC}"; exit 1
fi
echo -e "${GREEN}OK${NC}"
echo ""

# ── Helper ──────────────────────────────────────────────────────────────────
run_test() {
  local num="$1" desc="$2" token="$3" method="$4" endpoint="$5" expect="$6"
  echo -e "${YELLOW}[TEST $num]${NC} $desc"

  RESP=$(curl -s -w "\n%{http_code}" --max-time 5 \
    -X "$method" "$BACKEND_URL$endpoint" \
    -H "Cookie: auth_token=$token" \
    -H "Content-Type: application/json")

  CODE=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | head -1)

  if [ "$CODE" = "$expect" ]; then
    echo -e "  ${GREEN}✓ CORRECTO${NC} — HTTP $CODE"
  else
    echo -e "  ${RED}✗ FALLO${NC} — Se esperaba HTTP $expect, se recibió HTTP $CODE"
    echo "     Body: $BODY"
  fi
  echo ""
}

# ── TEST 1: Bodega A accede a su propia bodega → debe poder (200) ─────────
run_test 1 "Bodega A lee sus propios datos                → permitir (200)" \
  "$TOKEN_A" "GET" "/api/bodegas/$ID_BODEGA_A" "200"

# ── TEST 2: Bodega B intenta leer la bodega A → debe ser rechazado (403) ──
run_test 2 "Bodega B lee datos de otra bodega             → rechazar (403)" \
  "$TOKEN_B" "GET" "/api/bodegas/$ID_BODEGA_A" "403"

# ── TEST 3: Bodega B intenta actualizar la bodega A → debe ser rechazado ──
run_test 3 "Bodega B modifica datos de otra bodega        → rechazar (403)" \
  "$TOKEN_B" "PUT" "/api/bodegas/$ID_BODEGA_A" "403"

# ── TEST 4: Bodega B lee resultados de la bodega A → debe ser rechazado ───
run_test 4 "Bodega B lee resultados de otra bodega        → rechazar (403)" \
  "$TOKEN_B" "GET" "/api/bodegas/$ID_BODEGA_A/resultados-autoevaluacion" "403"

# ── TEST 5: Sin token → debe ser rechazado (401) ──────────────────────────
run_test 5 "Sin autenticación accede a bodega             → rechazar (401)" \
  "" "GET" "/api/bodegas/$ID_BODEGA_A" "401"

echo "=============================================="
echo " Tests completados"
echo "=============================================="
