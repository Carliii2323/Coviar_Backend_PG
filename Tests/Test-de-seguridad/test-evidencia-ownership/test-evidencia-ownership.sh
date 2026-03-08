#!/bin/bash
# Test de verificación de ownership en endpoints de evidencia 
#
# Requiere DOS tokens:
#   TOKEN_A: token de la bodega dueña de la autoevaluación
#   TOKEN_B: token de OTRA bodega distinta
#
# Para obtener los tokens:
#   1. Logueate con cada cuenta en la app
#   2. DevTools (F12) → Application → Cookies → copiá auth_token

BACKEND_URL="http://localhost:8080"

# ── CONFIGURAR ESTOS VALORES ────────────────────────────────────────────────
TOKEN_A="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNiwiZW1haWwiOiJkaWF6dHV0aW45MEBnbWFpbC5jb20iLCJ0aXBvX2N1ZW50YSI6IkJPREVHQSIsImJvZGVnYV9pZCI6MjYsImlzcyI6ImNvdmlhci1iYWNrZW5kIiwiZXhwIjoxNzcyODk3NjU3LCJuYmYiOjE3NzI4MTEyNTcsImlhdCI6MTc3MjgxMTI1N30.wDueQ2VzmTJWIdd8DoqWoqLKcHHFZjYWx7L-FGII7NE"   # token de la bodega dueña de la autoevaluación
TOKEN_B="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJjYXJsaUBtYWlsLmNvbSIsInRpcG9fY3VlbnRhIjoiQk9ERUdBIiwiYm9kZWdhX2lkIjoyNCwiaXNzIjoiY292aWFyLWJhY2tlbmQiLCJleHAiOjE3NzI4OTgxOTgsIm5iZiI6MTc3MjgxMTc5OCwiaWF0IjoxNzcyODExNzk4fQ.Bci-DLvEAPpHqq9u8HEud7VP2972fOJK8vwVtvR2dvY"   # token de OTRA bodega distinta

ID_AUTOEVALUACION=94   # ID de una autoevaluación que pertenece a TOKEN_A
ID_RESPUESTA=7440      # ID de una respuesta que pertenece a esa autoevaluación
# ────────────────────────────────────────────────────────────────────────────

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "=============================================="
echo " Test Ownership Evidencias"
echo " Autoevaluación bajo prueba: ID=$ID_AUTOEVALUACION"
echo "=============================================="
echo ""

# Verificar configuración
if [ -z "$TOKEN_A" ] || [ -z "$TOKEN_B" ] || [ "$ID_AUTOEVALUACION" = "0" ] || [ "$ID_RESPUESTA" = "0" ]; then
  echo -e "${RED}✗ Configurar TOKEN_A, TOKEN_B, ID_AUTOEVALUACION e ID_RESPUESTA antes de ejecutar${NC}"
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

# ── Helper ──────────────────────────────────────────────────────────────────
run_test() {
  local num="$1" desc="$2" token="$3" method="$4" endpoint="$5" expect="$6"
  echo -e "${YELLOW}[TEST $num]${NC} $desc"

  if [ -z "$token" ]; then
    RESP=$(curl -s -w "\n%{http_code}" --max-time 5 \
      -X "$method" "$BACKEND_URL$endpoint" \
      -H "Content-Type: application/json")
  else
    RESP=$(curl -s -w "\n%{http_code}" --max-time 5 \
      -X "$method" "$BACKEND_URL$endpoint" \
      -H "Cookie: auth_token=$token" \
      -H "Content-Type: application/json")
  fi

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

AE_URL="/api/autoevaluaciones/$ID_AUTOEVALUACION"
RESP_URL="$AE_URL/respuestas/$ID_RESPUESTA"

# ── TEST 1: Bodega A ve su propia lista de evidencias → permitir (200) ─────
run_test 1 "Bodega A lista evidencias propias               → permitir (200)" \
  "$TOKEN_A" "GET" "$AE_URL/evidencias" "200"

# ── TEST 2: Bodega B intenta listar evidencias ajenas → rechazar (403) ─────
run_test 2 "Bodega B lista evidencias de otra bodega        → rechazar (403)" \
  "$TOKEN_B" "GET" "$AE_URL/evidencias" "403"

# ── TEST 3: Bodega A ve su propia evidencia → permitir (200) ───────────────
run_test 3 "Bodega A obtiene evidencia propia               → permitir (200)" \
  "$TOKEN_A" "GET" "$RESP_URL/evidencia" "200"

# ── TEST 4: Bodega B intenta ver evidencia ajena → rechazar (403) ──────────
run_test 4 "Bodega B obtiene evidencia de otra bodega       → rechazar (403)" \
  "$TOKEN_B" "GET" "$RESP_URL/evidencia" "403"

# ── TEST 5: Bodega B intenta descargar evidencia ajena → rechazar (403) ────
run_test 5 "Bodega B descarga evidencia de otra bodega      → rechazar (403)" \
  "$TOKEN_B" "GET" "$RESP_URL/evidencia/descargar" "403"

# ── TEST 6: Bodega B intenta descargar ZIP ajeno → rechazar (403) ──────────
run_test 6 "Bodega B descarga ZIP de otra bodega            → rechazar (403)" \
  "$TOKEN_B" "GET" "$AE_URL/evidencias/descargar" "403"

# ── TEST 7: Bodega B intenta eliminar evidencia ajena → rechazar (403) ─────
run_test 7 "Bodega B elimina evidencia de otra bodega       → rechazar (403)" \
  "$TOKEN_B" "DELETE" "$RESP_URL/evidencia" "403"

# ── TEST 8: Sin token → rechazar (401) ────────────────────────────────────
run_test 8 "Sin autenticación accede a evidencias           → rechazar (401)" \
  "" "GET" "$AE_URL/evidencias" "401"

echo "=============================================="
echo " Tests completados"
echo "=============================================="
