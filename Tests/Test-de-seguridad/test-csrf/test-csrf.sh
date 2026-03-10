#!/bin/bash
# Test de protección CSRF 
#
# Verifica que peticiones mutantes (POST/PUT/DELETE) con un Origin no permitido
# sean rechazadas con HTTP 403, y que las que vienen de orígenes permitidos
# (o sin Origin) sean aceptadas.

BACKEND_URL="http://localhost:8080"

# ── CONFIGURAR ESTOS VALORES ────────────────────────────────────────────────
EMAIL="carli@mail.com"
PASSWORD="NuevaPass123"
# ────────────────────────────────────────────────────────────────────────────

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
PASS=0; FAIL=0

check() {
  local desc="$1" result="$2" expected="$3"
  if [ "$result" = "$expected" ]; then
    echo -e "  ${GREEN}✓ CORRECTO${NC} — $desc"
    ((PASS++))
  else
    echo -e "  ${RED}✗ FALLO${NC}   — $desc (esperado: $expected, obtenido: $result)"
    ((FAIL++))
  fi
}

echo "=============================================="
echo " Test Protección CSRF"
echo "=============================================="
echo ""

# Verificar backend
echo -n "Backend activo... "
CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$BACKEND_URL/health")
if [ "$CODE" = "000" ]; then
  echo -e "${RED}✗ No responde${NC}"; exit 1
fi
echo -e "${GREEN}OK${NC}"
echo ""

# ── TEST 1: POST con Origin malicioso → 403 ───────────────────────────────
echo -e "${YELLOW}[TEST 1]${NC} POST con Origin malicioso → rechazar (403)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -H "Origin: https://evil-attacker.com" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

check "Origin malicioso bloqueado" "$CODE" "403"
echo ""

# ── TEST 2: POST sin Origin → permitido (login normal desde curl/server) ──
echo -e "${YELLOW}[TEST 2]${NC} POST sin Origin → permitido (curl/server-to-server)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

# Login puede retornar 200 (ok) o 400 (datos), lo importante es que no sea 403
if [ "$CODE" != "403" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — Sin Origin permitido, HTTP $CODE"
  ((PASS++))
else
  echo -e "  ${RED}✗ FALLO${NC}   — Sin Origin fue bloqueado incorrectamente (403)"
  ((FAIL++))
fi
echo ""

# ── TEST 3: POST con Origin permitido (localhost) → permitido ─────────────
echo -e "${YELLOW}[TEST 3]${NC} POST con Origin permitido (localhost:3000) → aceptado"

LOGIN_RESP=$(curl -s -c /tmp/csrf_test_cookies.txt -w "\n%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:3000" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

LOGIN_CODE=$(echo "$LOGIN_RESP" | tail -1)
TOKEN=$(grep "auth_token" /tmp/csrf_test_cookies.txt | awk '{print $7}')

if [ "$LOGIN_CODE" != "403" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — Origin localhost:3000 aceptado, HTTP $LOGIN_CODE"
  ((PASS++))
else
  echo -e "  ${RED}✗ FALLO${NC}   — Origin permitido fue bloqueado (403)"
  ((FAIL++))
fi
echo ""

# ── TEST 4: PUT con Origin malicioso → 403 ────────────────────────────────
echo -e "${YELLOW}[TEST 4]${NC} PUT con Origin malicioso → rechazar (403)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X PUT "$BACKEND_URL/api/cuentas/password" \
  -H "Content-Type: application/json" \
  -H "Origin: https://attacker.io" \
  -H "Cookie: auth_token=${TOKEN:-dummy}" \
  -d "{\"currentPassword\":\"test\",\"newPassword\":\"test2\"}")

check "PUT con Origin malicioso bloqueado" "$CODE" "403"
echo ""

# ── TEST 5: DELETE con Origin malicioso → 403 ─────────────────────────────
echo -e "${YELLOW}[TEST 5]${NC} DELETE con Origin malicioso → rechazar (403)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X DELETE "$BACKEND_URL/api/autoevaluaciones/1/respuestas/1/evidencia" \
  -H "Origin: http://csrf-attack.example.com" \
  -H "Cookie: auth_token=${TOKEN:-dummy}")

check "DELETE con Origin malicioso bloqueado" "$CODE" "403"
echo ""

# ── TEST 6: GET con Origin malicioso → permitido (GET no muta estado) ─────
echo -e "${YELLOW}[TEST 6]${NC} GET con Origin malicioso → no bloqueado (método seguro)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X GET "$BACKEND_URL/health" \
  -H "Origin: https://evil-attacker.com")

if [ "$CODE" != "403" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — GET no bloqueado por CSRF (HTTP $CODE)"
  ((PASS++))
else
  echo -e "  ${RED}✗ FALLO${NC}   — GET fue bloqueado incorrectamente"
  ((FAIL++))
fi
echo ""

# ── TEST 7: POST con Referer malicioso → 403 ──────────────────────────────
echo -e "${YELLOW}[TEST 7]${NC} POST sin Origin pero con Referer malicioso → rechazar (403)"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -H "Referer: https://evil-attacker.com/csrf-page" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

check "Referer malicioso bloqueado" "$CODE" "403"
echo ""

# ── TEST 8: POST con Referer permitido → aceptado ─────────────────────────
echo -e "${YELLOW}[TEST 8]${NC} POST sin Origin pero con Referer permitido → aceptado"

CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X POST "$BACKEND_URL/api/login" \
  -H "Content-Type: application/json" \
  -H "Referer: http://localhost:3000/login" \
  -d "{\"email_login\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

if [ "$CODE" != "403" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — Referer permitido aceptado, HTTP $CODE"
  ((PASS++))
else
  echo -e "  ${RED}✗ FALLO${NC}   — Referer permitido fue bloqueado"
  ((FAIL++))
fi
echo ""

# Limpiar
rm -f /tmp/csrf_test_cookies.txt

# ── Resumen ───────────────────────────────────────────────────────────────
echo "=============================================="
echo " Resultados: ${PASS} pasados, ${FAIL} fallidos"
echo "=============================================="
if [ "$FAIL" -eq 0 ]; then
  echo -e " ${GREEN}✓ Protección CSRF funcionando correctamente${NC}"
else
  echo -e " ${RED}✗ Protección CSRF incompleta — Revisar los fallos indicados${NC}"
fi
echo ""
