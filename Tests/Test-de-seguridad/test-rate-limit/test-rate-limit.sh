#!/bin/bash
# Test de rate limiting global 
#
# Límites implementados:
#   - Global:   200 req/min por IP (todos los endpoints)
#   - Escritura: 20 req/min por IP (POST, PUT, DELETE)

BACKEND_URL="http://localhost:8080"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJjYXJsaUBtYWlsLmNvbSIsInRpcG9fY3VlbnRhIjoiQk9ERUdBIiwiYm9kZWdhX2lkIjoyNCwiaXNzIjoiY292aWFyLWJhY2tlbmQiLCJleHAiOjE3NzI4OTQ5NTQsIm5iZiI6MTc3MjgwODU1NCwiaWF0IjoxNzcyODA4NTU0fQ.Nn9zDhM3yT8hSj31cdvzs65N1akNTC69K6FDc1JsrOM"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

echo "=============================================="
echo " Test Rate Limiting "
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

# ── TEST 1: Límite de escritura (POST) — superar 20 req/min ─────────────────
echo -e "${YELLOW}[TEST 1]${NC} Enviando 25 POST seguidos → debe bloquear en el 21+ (429)"
echo ""

blocked=0
for i in $(seq 1 25); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 \
    -X POST "$BACKEND_URL/api/login" \
    -H "Content-Type: application/json" \
    -d '{"email_login":"test@test.com","password":"test"}')

  if [ "$CODE" = "429" ]; then
    echo -e "  Request $i: ${GREEN}429 Too Many Requests${NC} — rate limit activo"
    blocked=1
    break
  else
    echo -e "  Request $i: HTTP $CODE"
  fi
done

if [ "$blocked" = "1" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — El límite de escritura funciona"
else
  echo -e "  ${RED}✗ FALLO${NC} — Nunca se recibió 429 en 25 POSTs"
fi
echo ""

# Esperar que se resetee la ventana
echo -e "${CYAN}Esperando 65 segundos para resetear la ventana de rate limit...${NC}"
sleep 65

# ── TEST 2: Límite global (GET) — superar 200 req/min ───────────────────────
echo -e "${YELLOW}[TEST 2]${NC} Enviando 210 GET seguidos → debe bloquear en el 201+ (429)"
echo ""

blocked=0
for i in $(seq 1 210); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 \
    -H "Cookie: auth_token=$TOKEN" \
    "$BACKEND_URL/health")

  if [ "$CODE" = "429" ]; then
    echo -e "  Request $i: ${GREEN}429 Too Many Requests${NC} — rate limit global activo"
    blocked=1
    break
  fi
done

if [ "$blocked" = "1" ]; then
  echo -e "  ${GREEN}✓ CORRECTO${NC} — El límite global funciona"
else
  echo -e "  ${RED}✗ FALLO${NC} — Nunca se recibió 429 en 210 GETs"
fi
echo ""

echo "=============================================="
echo " Tests completados"
echo "=============================================="
