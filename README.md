# fiss-lint

`fiss-lint` — быстрый консольный валидатор и линтер проектов на соответствие спецификации [File-based Intellectual Space Standard (FISS) v1.0.0](https://fiss.vorozhko.ru/v1.0.0/llms.txt).

Утилита проверяет структуру каталога `FISS/`, целостность навигационного графа, корректность условий чтения, ссылочную связность файлов и файлы входа для ИИ-агентов (`AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `QWEN.md`).

---

## Возможности

- **Проверка правил FISS v1.0.0**: 18 специализированных правил валидации (структура, навигация, переопределения, ссылки, агенты).
- **Машиночитаемый формат**: поддержка `--format json` для интеграции в CI/CD пайплайны и генерации отчетов.
- **Строгий режим (`--strict`)**: эскалация предупреждений (`Warning`) до кода завершения `1`.
- **Нулевые зависимости**: утилита написана на чистом Go и скомпилирована без CGO (`CGO_ENABLED=0`).
- **Кроссплатформенность**: бинарные сборки для Linux, macOS и Windows (amd64, arm64).

---

## Установка и сборка

### Сборка из исходников

Требуется Go 1.21 или новее:

```bash
git clone https://github.com/AndreyVorozhko/fiss-lint.git
cd fiss-lint
make build
```

Собранный исполняемый файл будет доступен в `bin/fiss-lint`.

Для сборки под все поддерживаемые платформы:

```bash
make build-all
```

Бинарники будут созданы в папке `bin/`:
- `bin/fiss-lint-linux-amd64`, `bin/fiss-lint-linux-arm64`
- `bin/fiss-lint-darwin-amd64`, `bin/fiss-lint-darwin-arm64`
- `bin/fiss-lint-windows-amd64.exe`, `bin/fiss-lint-windows-arm64.exe`

---

## Использование

```bash
fiss-lint [flags] [path]
```

Если путь `[path]` не указан, проверка запускается в текущем рабочем каталоге (`.`).

### Флаги командной строки

| Флаг | Описание | Значение по умолчанию |
|---|---|---|
| `-h`, `--help` | Показать справку по утилите и сводку правил FISS | `false` |
| `-v`, `--version` | Показать версию, коммит и дату сборки | `false` |
| `--format string` | Формат вывода: `text` (терминальный построчный) или `json` | `text` |
| `--strict` | Завершать работу с кодом `1` при обнаружении любых предупреждений (`Warning`) | `false` |

### Коды завершения (Exit Codes)

| Код | Описание |
|---|---|
| `0` | Проверка успешно пройдена (ошибок нет; предупреждения допустимы без `--strict`). |
| `1` | Обнаружены нарушения FISS (ошибки, либо предупреждения при указании флага `--strict`), либо произошла ошибка чтения файловой системы. |
| `2` | Ошибка в переданных аргументах командной строки (например, недопустимый формат `--format`). |

---

## Форматы вывода

### Текстовый формат (`--format text`)

По умолчанию вывод формируется в лаконичном диагностическом формате:

```text
[ERROR] [FISS-R001] FISS: directory FISS/ not found
[WARNING] [FISS-R011] FISS/INDEX.md: recommended link to official standard https://fiss.vorozhko.ru not found
```

Если проект полностью валиден, утилита ничего не выводит в терминал и возвращает код `0`.

### Машиночитаемый JSON (`--format json`)

При указании `--format json` формируется структурированный JSON с отступами:

```json
{
  "issues": [
    {
      "rule_id": "FISS-R011",
      "severity": "WARNING",
      "file": "FISS/INDEX.md",
      "line": 0,
      "message": "recommended link to official standard https://fiss.vorozhko.ru not found"
    }
  ],
  "summary": {
    "errors": 0,
    "warnings": 1,
    "total": 1
  }
}
```

Если нарушений не найдено, поле `issues` содержит пустой массив `[]`, а `summary` — нулевые счетчики:

```json
{
  "issues": [],
  "summary": {
    "errors": 0,
    "warnings": 0,
    "total": 0
  }
}
```

---

## Интеграция в CI/CD

### GitHub Actions

Пример рабочего процесса `.github/workflows/fiss-lint.yml`:

```yaml
name: FISS Validation

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  lint:
    name: Lint FISS Intellectual Space
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          check-latest: true

      - name: Build fiss-lint
        run: make build

      - name: Run fiss-lint in strict mode
        run: ./bin/fiss-lint --strict .

      - name: Generate JSON report artifact
        if: always()
        run: ./bin/fiss-lint --format json . > fiss-report.json || true

      - name: Upload report artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: fiss-validation-report
          path: fiss-report.json
```

---

## Интеграция с Git Hooks

### pre-commit framework

Для использования с [pre-commit](https://pre-commit.com) добавьте следующий хук в `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: fiss-lint
        name: fiss-lint
        entry: fiss-lint --strict .
        language: system
        pass_filenames: false
        always_run: true
```

### Git pre-commit Hook вручную

Создайте файл `.git/hooks/pre-commit` (или настройте `core.hooksPath`):

```bash
#!/bin/sh
set -e

fiss-lint --strict .
```

Сделайте файл исполняемым:

```bash
chmod +x .git/hooks/pre-commit
```

---

## Лицензия

Проект распространяется на условиях лицензии MIT.
