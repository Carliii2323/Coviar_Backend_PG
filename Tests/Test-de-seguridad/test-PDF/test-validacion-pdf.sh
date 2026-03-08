#!/bin/bash
# Test de validación de magic bytes PDF 

BACKEND_URL="http://localhost:8080"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJjYXJsaUBtYWlsLmNvbSIsInRpcG9fY3VlbnRhIjoiQk9ERUdBIiwiaXNzIjoiY292aWFyLWJhY2tlbmQiLCJleHAiOjE3NzI4ODg4MjgsIm5iZiI6MTc3MjgwMjQyOCwiaWF0IjoxNzcyODAyNDI4fQ.gToqkCEGJEKf6om9msy953CUJcWQXiVr8-WM3PdsVA0"

ID_AUTOEVALUACION=90
ID_RESPUESTA=19518

ENDPOINT="$BACKEND_URL/api/autoevaluaciones/$ID_AUTOEVALUACION/respuestas/$ID_RESPUESTA/evidencias"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "=============================================="
echo " Test Magic Bytes PDF — Endpoint:"
echo " $ENDPOINT"
echo "=============================================="
echo ""

# Verificar backend
echo -n "Backend activo... "
CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$BACKEND_URL/")
if [ "$CODE" = "000" ]; then
  echo -e "${RED}✗ No responde${NC}"; exit 1
fi
echo -e "${GREEN}OK (HTTP $CODE)${NC}"

# Verificar token - hacer una llamada autenticada simple
echo -n "Token válido... "
CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -H "Cookie: auth_token=$TOKEN" \
  "$BACKEND_URL/api/autoevaluaciones/$ID_AUTOEVALUACION/evidencias")
if [ "$CODE" = "401" ] || [ "$CODE" = "403" ]; then
  echo -e "${RED}✗ Token inválido o expirado (HTTP $CODE)${NC}"
  echo ""
  echo "  El token JWT está vencido. Para obtener uno nuevo:"
  echo "  1. Logueate en la app"
  echo "  2. Abrí DevTools (F12) → Application → Cookies"
  echo "  3. Copiá el valor de 'auth_token'"
  echo "  4. Reemplazalo en la variable TOKEN de este script"
  exit 1
elif [ "$CODE" = "000" ]; then
  echo -e "${RED}✗ Sin respuesta${NC}"; exit 1
fi
echo -e "${GREEN}OK (HTTP $CODE)${NC}"
echo ""

# ── Helper ─────────────────────────────────────────────────────────────────
run_test() {
  local num="$1" desc="$2" file="$3" expect_ok="$4"
  echo -e "${YELLOW}[TEST $num]${NC} $desc"

  # Verificar que el archivo existe
  if [ ! -f "$file" ]; then
    echo -e "  ${RED}✗ No se pudo crear el archivo de prueba: $file${NC}"; echo ""; return
  fi

  RESP=$(curl -s -w "\n%{http_code}" --max-time 10 \
    -X POST "$ENDPOINT" \
    -H "Cookie: auth_token=$TOKEN" \
    -F "file=@$file" \
    2>&1)

  HTTP_CODE=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | head -1)

  if [ "$HTTP_CODE" = "000" ]; then
    echo -e "  ${RED}✗ Sin respuesta (HTTP 000)${NC} — posible problema de ruta o timeout"
    echo "     Archivo: $file ($(wc -c < "$file") bytes)"
    # Intentar con ruta Windows explícita
    WIN_FILE=$(cygpath -w "$file" 2>/dev/null || echo "$file")
    echo "     Ruta Windows: $WIN_FILE"
    echo ""
    return
  fi

  if [ "$expect_ok" = "true" ]; then
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
      echo -e "  ${GREEN}✓ CORRECTO${NC} — Aceptado (HTTP $HTTP_CODE)"
    else
      echo -e "  ${RED}✗ FALLO${NC} — Rechazó un PDF válido (HTTP $HTTP_CODE): $BODY"
    fi
  else
    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "422" ] || [ "$HTTP_CODE" = "413" ]; then
      echo -e "  ${GREEN}✓ CORRECTO${NC} — Rechazado (HTTP $HTTP_CODE)"
    else
      echo -e "  ${RED}✗ FALLO${NC} — Debería haberse rechazado (HTTP $HTTP_CODE): $BODY"
    fi
  fi
  echo ""
}

# ── Limpiar evidencia existente antes de los tests ─────────────────────────
echo -n "Limpiando evidencia previa... "
DEL_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
  -X DELETE "$BACKEND_URL/api/autoevaluaciones/$ID_AUTOEVALUACION/respuestas/$ID_RESPUESTA/evidencia" \
  -H "Cookie: auth_token=$TOKEN")
if [ "$DEL_CODE" = "200" ] || [ "$DEL_CODE" = "204" ]; then
  echo -e "${GREEN}OK (evidencia eliminada)${NC}"
elif [ "$DEL_CODE" = "404" ]; then
  echo -e "${GREEN}OK (no había evidencia)${NC}"
else
  echo -e "${YELLOW}HTTP $DEL_CODE (continuando de todas formas)${NC}"
fi
echo ""

# ── Crear archivos en directorio actual ────────────────────────────────────
printf '\x4D\x5A\x90\x00exe-renombrado' > ejecutable_test.pdf
echo "soy texto plano, no pdf"            > texto_test.pdf
printf '\x50\x4B\x03\x04zip-renombrado'  > zip_test.pdf

cat > real_test.pdf << 'EOF'
%PDF-1.4
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >> endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer << /Size 4 /Root 1 0 R >>
startxref
190
%%EOF
EOF

# PDF grande: header válido + padding
printf '%%PDF-1.4\n' > grande_test.pdf
python3 -c "import sys; sys.stdout.buffer.write(b'0'*2200000)" >> grande_test.pdf 2>/dev/null || \
  dd if=/dev/zero bs=1024 count=2200 2>/dev/null >> grande_test.pdf

# ── Ejecutar tests ─────────────────────────────────────────────────────────
run_test 1 "EXE renombrado como .pdf (magic MZ)       → rechazar" "ejecutable_test.pdf" "false"
run_test 2 "Texto plano renombrado como .pdf           → rechazar" "texto_test.pdf"      "false"
run_test 3 "ZIP renombrado como .pdf (magic PK)        → rechazar" "zip_test.pdf"        "false"
run_test 4 "PDF real con %PDF- (magic bytes válidos)   → aceptar"  "real_test.pdf"       "true"

# Limpiar evidencia subida por test 4 para que test 5 no se bloquee
curl -s -o /dev/null -X DELETE \
  "$BACKEND_URL/api/autoevaluaciones/$ID_AUTOEVALUACION/respuestas/$ID_RESPUESTA/evidencia" \
  -H "Cookie: auth_token=$TOKEN"

run_test 5 "PDF mayor a 2MB                            → rechazar" "grande_test.pdf"     "false"

# Limpiar
rm -f ejecutable_test.pdf texto_test.pdf zip_test.pdf real_test.pdf grande_test.pdf

echo "=============================================="
echo " Tests completados"
echo "=============================================="
