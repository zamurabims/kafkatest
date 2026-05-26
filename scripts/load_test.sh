#!/bin/bash

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================"
echo "Trending Service Load Test"
echo "Target: $BASE_URL"
echo "========================================"

echo ""
echo ">>> Проверяем что сервис живой..."
curl -sf "$BASE_URL/health" || { echo "Service not available"; exit 1; }
echo " OK"

echo ""
echo ">>> Текущий топ перед тестом:"
curl -s "$BASE_URL/top?limit=5" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/top?limit=5"

echo ""
echo "========================================"
echo "TEST 1: GET /top — 50k запросов, 200 конкурентных"
echo "========================================"
hey -n 50000 -c 200 -q 0 "$BASE_URL/top?limit=10"

echo ""
echo "========================================"
echo "TEST 2: GET /top — 10k запросов, 500 конкурентных (пиковая нагрузка)"
echo "========================================"
hey -n 10000 -c 500 -q 0 "$BASE_URL/top?limit=10"

echo ""
echo "========================================"
echo "TEST 3: Mixed — одновременно /top и /stoplist"
echo "========================================"
hey -n 20000 -c 200 -q 0 "$BASE_URL/top?limit=5" &
hey -n 1000  -c 10  -q 0 -m POST \
    -H "Content-Type: application/json" \
    -d '{"word":"testword"}' \
    "$BASE_URL/stoplist" &
wait

echo ""
echo "Load test complete."
