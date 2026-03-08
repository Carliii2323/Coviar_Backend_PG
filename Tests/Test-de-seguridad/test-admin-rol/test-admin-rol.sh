#!/bin/bash
# Test de verificación de rol en endpoints de administrador
#
# Requiere DOS tokens:
#   TOKEN_BODEGA:  token de una cuenta BODEGA (sin rol admin)
#   TOKEN_ADMIN:   token de una cuenta ADMINISTRADOR_APP
#
# Para obtener los tokens:
#   1. Logueate con cada cuenta en la app
#   2. DevTools (F12) → Application → Cookies → copiá auth_token

BACKEND_URL="http://localhost:8080"

# ── CONFIGURAR ESTOS VALORES ────────────────────────────────────────────────
TOKEN_BODEGA="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNiwiZW1haWwiOiJkaWF6dHV0aW45MEBnbWFpbC5jb20iLCJ0aXBvX2N1ZW50YSI6IkJPREVHQSIsImJvZGVnYV9pZCI6MjYsImlzcyI6ImNvdmlhci1iYWNrZW5kIiwiZXhwIjoxNzcyODk3NjU3LCJuYmYiOjE3NzI4MTEyNTcsImlhdCI6MTc3MjgxMTI1N30.wDueQ2VzmTJWIdd8DoqWoqLKcHHFZjYWx7L-FGII7NE"   # token de una cuenta BODEGA
TOKEN_ADMIN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozNCwiZW1haWwiOiJicmVuZGEuc2lsdmE0MzQ1QGdtYWlsLmNvbSIsInRpcG9fY3VlbnRhIjoiQURNSU5JU1RSQURPUl9BUFAiLCJib2RlZ2FfaWQiOjAsImlzcyI6ImNvdmlhci1iYWNrZW5kIiwiZXhwIjoxNzcyODk3Njc4LCJuYmYiOjE3NzI4MTEyNzgsImlhdCI6MTc3MjgxMTI3OH0.0CcbrYBM8D4E5_Ymr4KkuFM1YEE_EC4ghxHFYhSTtqE"    # token de una cuenta ADMINISTRADOR_APP
# ────────────────────────────────────────────────────────────────────────────

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "=============================================="
echo " Test Rol Administrador"
echo "=============================================="
echo ""

# Verificar que los tokens estén configurados
if [ -z "$TOKEN_BODEGA" ] || [ -z "$TOKEN_ADMIN" ]; then
  echo -e "${RED}✗ Configurar TOKEN_BODEGA y TOKEN_ADMIN en el script antes de ejecutar${NC}"
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

# ── TEST 1: Sin token → 401 ───────────────────────────────────────────────
run_test 1 "Sin autenticación accede a /admin/stats         → rechazar (401)" \
  "" "GET" "/api/admin/stats" "401"

# ── TEST 2: Bodega accede a stats → 403 ──────────────────────────────────
run_test 2 "Cuenta BODEGA accede a /admin/stats             → rechazar (403)" \
  "$TOKEN_BODEGA" "GET" "/api/admin/stats" "403"

# ── TEST 3: Bodega accede a evaluaciones → 403 ───────────────────────────
run_test 3 "Cuenta BODEGA accede a /admin/evaluaciones      → rechazar (403)" \
  "$TOKEN_BODEGA" "GET" "/api/admin/evaluaciones" "403"

# ── TEST 4: Bodega accede a listado de bodegas → 403 ─────────────────────
run_test 4 "Cuenta BODEGA accede a /admin/bodegas           → rechazar (403)" \
  "$TOKEN_BODEGA" "GET" "/api/admin/bodegas" "403"

# ── TEST 5: Admin accede a stats → 200 ───────────────────────────────────
run_test 5 "Cuenta ADMIN accede a /admin/stats              → permitir (200)" \
  "$TOKEN_ADMIN" "GET" "/api/admin/stats" "200"

# ── TEST 6: Admin accede a evaluaciones → 200 ────────────────────────────
run_test 6 "Cuenta ADMIN accede a /admin/evaluaciones       → permitir (200)" \
  "$TOKEN_ADMIN" "GET" "/api/admin/evaluaciones" "200"

# ── TEST 7: Admin accede a listado de bodegas → 200 ──────────────────────
run_test 7 "Cuenta ADMIN accede a /admin/bodegas            → permitir (200)" \
  "$TOKEN_ADMIN" "GET" "/api/admin/bodegas" "200"

echo "=============================================="
echo " Tests completados"
echo "=============================================="
