#!/usr/bin/env bash
set -euo pipefail

echo "=== 1. Проверка протокола обнаружения fiss-lint ==="
if command -v fiss-lint >/dev/null 2>&1; then
    FISS_LINT="fiss-lint"
elif [ -x "./bin/fiss-lint" ]; then
    FISS_LINT="./bin/fiss-lint"
elif [ -x "/workspace/bin/fiss-lint" ]; then
    FISS_LINT="/workspace/bin/fiss-lint"
else
    echo "ERROR: fiss-lint binary not found"
    exit 1
fi
echo "Утилита обнаружена: $FISS_LINT"
"$FISS_LINT" --version

echo "=== 2. Проверка реального пространства (fiss-validate / fiss-maintain) ==="
echo "Запуск в формате JSON..."
JSON_OUTPUT=$("$FISS_LINT" --format json .)
echo "$JSON_OUTPUT"
TOTAL_ISSUES=$(echo "$JSON_OUTPUT" | grep -o '"total": [0-9]*' | awk '{print $2}')
if [ "$TOTAL_ISSUES" -ne 0 ]; then
    echo "ERROR: Ожидалось 0 проблем, получено $TOTAL_ISSUES"
    exit 1
fi
echo "JSON-проверка пройдена успешно (0 проблем)."

echo "Запуск в строгом режиме (--strict)..."
"$FISS_LINT" --strict .
echo "Строгий режим успешно пройден (exit code 0)."

echo "=== 3. Проверка негативных сценариев на фикстурах ==="
echo "Тест: testdata/invalid_link_not_found..."
set +e
"$FISS_LINT" --strict testdata/invalid_link_not_found >/dev/null 2>&1
EXIT_CODE=$?
set -e
if [ "$EXIT_CODE" -eq 0 ]; then
    echo "ERROR: Ожидался ненулевой код возврата для невалидной фикстуры"
    exit 1
fi
echo "Негативный сценарий успешно отловлен (exit code $EXIT_CODE)."

echo "=== 4. Проверка отсутствия дублирования в навыках ==="
SKILLS_DIR="${HOME}/Projects/ai-skills/skills/fiss"
if [ -d "$SKILLS_DIR" ]; then
    # Проверка fiss-validate на отсутствие старых ручных проверок Stage 2
    if grep -q "check_index_link_resolution" "$SKILLS_DIR/fiss-validate/SKILL.md"; then
        echo "ERROR: В fiss-validate/SKILL.md найден устаревший ручной метод check_index_link_resolution"
        exit 1
    fi
    # Проверка вызова fiss-lint
    if ! grep -q "fiss-lint --format json" "$SKILLS_DIR/fiss-validate/SKILL.md"; then
        echo "ERROR: В fiss-validate/SKILL.md отсутствует вызов fiss-lint --format json"
        exit 1
    fi
    if ! grep -q "fiss-lint --strict" "$SKILLS_DIR/fiss-maintain/SKILL.md"; then
        echo "ERROR: В fiss-maintain/SKILL.md отсутствует вызов fiss-lint --strict"
        exit 1
    fi
    echo "Проверка навыков в ai-skills пройдена: механическое дублирование удалено, вызов линтера зафиксирован."
fi

echo "=== ВСЕ ПРОВЕРКИ СКВОЗНОЙ ВЕРИФИКАЦИИ УСПЕШНО ПРОЙДЕНЫ ==="
