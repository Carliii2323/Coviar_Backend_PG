#!/bin/bash
# Test de protección contra fuerza bruta 

BACKEND_URL="http://localhost:8080"
EMAIL="carli@mail.com"        # email de una cuenta existente
WRONG_PASS="password_incorrecto_999"
MAX_ATTEMPTS=5

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

echo "=============================================="
echo " Test Brute Force — Endpoint: POST /api/login"
echo "=============================================="
echo ""

# Verificar backend
echo -n "Backend activo... "
CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$BACKEND_URL/")
if [ "$CODE" = "000" ]; then
  echo -e "${RED}✗ No responde${NC}"; exit 1
fi
echo -e "${GREEN}OK (HTTP $CODE)${NC}"
echo ""

# ── Helper ──────────────────────────────────────────────────────────────────
do_login() {
  local pass="$1"
  curl -s -w "\n%{http_code}" --max-time 5 \
    -X POST "$BACKEND_URL/api/login" \
    -H "Content-Type: application/json" \
    -d "{\"email_login\":\"$EMAIL\",\"password\":\"$pass\"}"
}

# ── TEST 1: 5 intentos fallidos consecutivos ─────────────────────────────────
echo -e "${YELLOW}[TEST 1]${NC} Enviando $MAX_ATTEMPTS intentos fallidos con contraseña incorrecta..."
echo ""

for i in $(seq 1 $MAX_ATTEMPTS); do
  RESP=$(do_login "$WRONG_PASS")
  CODE=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | head -1)

  if [ "$CODE" = "401" ]; then
    echo -e "  Intento $i: ${GREEN}401 Unauthorized${NC} (credenciales inválidas — correcto)"
  elif [ "$CODE" = "429" ]; then
    echo -e "  Intento $i: ${YELLOW}429 Too Many Requests${NC} — bloqueado antes de los $MAX_ATTEMPTS intentos"
    break
  else
    echo -e "  Intento $i: ${RED}HTTP $CODE${NC} — $BODY"
  fi
done
echo ""

# ── TEST 2: El 6to intento debe devolver 429 ────────────────────────────────
echo -e "${YELLOW}[TEST 2]${NC} 6to intento — debe devolver 429 Too Many Requests..."
RESP=$(do_login "$WRONG_PASS")
CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -1)

if [ "$CODE" = "429" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — Bloqueado (HTTP 429): $BODY"
else
  echo -e "  ${RED}✗ FALLO${NC} — Se esperaba 429, se recibió HTTP $CODE: $BODY"
fi
echo ""

# ── TEST 3: Un intento más para confirmar bloqueo persistente ────────────────
echo -e "${YELLOW}[TEST 3]${NC} Reintento adicional — debe seguir bloqueado (429)..."
RESP=$(do_login "$WRONG_PASS")
CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -1)

if [ "$CODE" = "429" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — Sigue bloqueado (HTTP 429)"
else
  echo -e "  ${RED}✗ FALLO${NC} — HTTP $CODE: $BODY"
fi
echo ""

# ── Nota sobre el desbloqueo ─────────────────────────────────────────────────
echo -e "${CYAN}Nota:${NC} La IP queda bloqueada 15 minutos."
echo "      Para resetear, reiniciar el backend (el contador es en memoria)."
echo ""
echo "=============================================="
echo " Tests completados"
echo "=============================================="
