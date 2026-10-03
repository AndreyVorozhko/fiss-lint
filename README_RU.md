# fiss-lint

**Русский** | [English](README.md)

`fiss-lint` — быстрый консольный статический анализатор и линтер проектов на соответствие спецификации стандарта [File-based Intellectual Space Standard (FISS) v1.0.0](https://fiss.vorozhko.ru/v1.0.0/llms.txt) (или официального [GitHub-зеркала стандарта](https://github.com/AndreyVorozhko/fiss)).

Утилита проверяет физическую структуру каталога `FISS/`, целостность и связность навигационного графа, корректность условий чтения, синтаксис производных материалов, достижимость областей знаний и валидность файлов входа для ИИ-агентов (`AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `QWEN.md`).

---

## Возможности

- **Полное покрытие детерминированных правил FISS v1.0.0**: 18 специализированных правил валидации (`FISS-R001` — `FISS-R018`), охватывающих структуру, навигацию, переопределения, ссылки, разделение областей и точки входа агентов.
- **Машиночитаемый формат**: поддержка `--format json` с инъекцией семантической версии в блок `summary` для интеграции в CI/CD пайплайны и генерации отчетов.
- **Строгий режим (`--strict`)**: автоматическая эскалация предупреждений (`Warning`) до ненулевого кода завершения `1`.
- **Нулевые зависимости**: утилита написана на чистом Go и скомпилирована без CGO (`CGO_ENABLED=0`).
- **Кроссплатформенность**: готовые бинарные сборки для Linux, macOS и Windows (архитектуры `amd64` и `arm64`).

---

## Установка

### 1. Быстрая установка скриптом (Linux, macOS, Windows)

Установка одной командой скачивает скомпилированный релизный бинарник под вашу операционную систему и процессорную архитектуру без необходимости устанавливать Go:

**Linux и macOS (POSIX sh/bash через `curl`):**
```bash
curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
```

Установка конкретной версии (например, `v1.0.0`):
```bash
curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash -s -- v1.0.0
```

По умолчанию утилита устанавливается в `~/.local/bin` (или в `/usr/local/bin` при наличии прав суперпользователя). Каталог установки можно переопределить через переменную `INSTALL_DIR`:
```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.ps1 | iex
```

Установка конкретной версии:
```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.ps1))) -Version v1.0.0
```

Скрипт скачивает исполняемый файл `fiss-lint.exe` в `%LOCALAPPDATA%\Programs\fiss-lint` (или `~/bin`) и автоматически добавляет директорию в переменную окружения `User PATH`.

### 2. Установка через Go Toolchain

Если в системе установлен Go 1.21 или новее:

```bash
go install github.com/AndreyVorozhko/fiss-lint/cmd/fiss-lint@latest
```

### 3. Готовые скомпилированные релизы

Вы можете напрямую загрузить скомпилированный бинарный файл для вашей системы со страницы [GitHub Releases](https://github.com/AndreyVorozhko/fiss-lint/releases):
- Linux: `fiss-lint-linux-amd64`, `fiss-lint-linux-arm64`
- macOS: `fiss-lint-darwin-amd64`, `fiss-lint-darwin-arm64`
- Windows: `fiss-lint-windows-amd64.exe`, `fiss-lint-windows-arm64.exe`

Сделайте файл исполняемым (`chmod +x fiss-lint-*`) и поместите в директорию из переменной `$PATH`.

### 4. Сборка из исходников

```bash
git clone https://github.com/AndreyVorozhko/fiss-lint.git
cd fiss-lint
make build
```

Собранный исполняемый файл будет доступен в `bin/fiss-lint`.

Для сборки кроссплатформенной матрицы бинарников выполните:
```bash
make build-all
```

---

## Использование

```bash
fiss-lint [flags] [path]
```

Если целевой путь `[path]` не указан, проверка запускается в текущем рабочем каталоге (`.`).

### Флаги командной строки

| Флаг | Описание | Значение по умолчанию |
|---|---|---|
| `-h`, `--help` | Показать справку по утилите и сводный реестр правил FISS | `false` |
| `-v`, `--version` | Показать версию, коммит, дату сборки и отметку Built with FISS | `false` |
| `--format string` | Формат диагностического вывода: `text` (построчный) или `json` | `text` |
| `--strict` | Завершать работу с кодом `1` при обнаружении любых предупреждений (`Warning`) | `false` |

### Коды завершения (Exit Codes)

| Код | Описание |
|---|---|
| `0` | Проверка успешно пройдена (ошибок нет; предупреждения допустимы без `--strict`). |
| `1` | Обнаружены нарушения FISS (ошибки, либо предупреждения в режиме `--strict`), либо произошла ошибка чтения файловой системы. |
| `2` | Ошибка в переданных аргументах командной строки (например, неподдерживаемый формат `--format`). |

---

## Форматы вывода

### Текстовый формат (`--format text`)

По умолчанию диагностический вывод формируется в лаконичном построчном виде:

```text
[ERROR] [FISS-R001] FISS: directory FISS/ not found
[WARNING] [FISS-R011] FISS/INDEX.md: recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found
```

Если проект полностью валиден, утилита ничего не выводит в стандартный поток вывода и завершается с кодом `0`.

### Машиночитаемый JSON (`--format json`)

При передаче флага `--format json` формируется структурированный JSON-документ:

```json
{
  "issues": [
    {
      "rule_id": "FISS-R011",
      "severity": "WARNING",
      "file": "FISS/INDEX.md",
      "line": 0,
      "message": "recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found"
    }
  ],
  "summary": {
    "version": "v1.0.0",
    "errors": 0,
    "warnings": 1,
    "total": 1
  }
}
```

Если нарушений не обнаружено, массив `issues` пуст, а счетчики равны `0`:

```json
{
  "issues": [],
  "summary": {
    "version": "v1.0.0",
    "errors": 0,
    "warnings": 0,
    "total": 0
  }
}
```

---

## Реестр детерминированных правил (FISS-R001 — FISS-R018)

Линтер реализует 18 детерминированных правил стандарта FISS v1.0.0:

| Код правила | Уровень | Категория | Нормативное требование стандарта |
|---|---|---|---|
| `FISS-R001` | Error | Структура корня | В корне проекта обязан присутствовать каталог `FISS/`. |
| `FISS-R002` | Error | Структура корня | Каталог `FISS/` обязан содержать файлы `INDEX.md` и `BOOTSTRAP.md`. |
| `FISS-R003` | Error | Навигация | Корневой `FISS/INDEX.md` обязан содержать ссылку на `BOOTSTRAP.md`. |
| `FISS-R004` | Error | Навигация | Ссылка на `BOOTSTRAP.md` обязана содержать условие чтения, требующее ознакомления перед началом работы. |
| `FISS-R005` | Error | Синтаксис | Каждая навигационная запись в любом `INDEX.md` обязана быть строгим двухстрочным элементом (`- [Title](target.md)` + `  Read when: condition`). Маркер не подлежит локализации. |
| `FISS-R006` | Error | Целостность ссылок | Каждая относительная ссылка в индексе обязана разрешаться в существующий физический `.md` файл или `INDEX.md` составной области. Ссылки на каталоги без `INDEX.md` запрещены. |
| `FISS-R007` | Error | Топология | Составная область (каталог, на который ссылается навигация) обязана содержать собственный `INDEX.md`. |
| `FISS-R008` | Error | Топология | Каждая используемая область обязана быть достижима из корневого `FISS/INDEX.md` через дерево индексов. Изолированные (orphan) файлы запрещены. |
| `FISS-R009` | Error | Переопределения | Если существует каталог `FISS/overrides/`, он обязан содержать `INDEX.md`, и `FISS/INDEX.md` обязан ссылаться на него с условием чтения перед применением навыков. |
| `FISS-R010` | Error | Точка входа агентов | Если в корне проекта существует входной файл инструкций для агентов (`AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `QWEN.md`), он обязан направлять агентов к `FISS/INDEX.md`. |
| `FISS-R011` | Warning | Ссылка на стандарт | В `FISS/INDEX.md` рекомендуется указывать ссылку на официальный сайт стандарта (`https://fiss.vorozhko.ru`) или официальное GitHub-зеркало (`https://github.com/AndreyVorozhko/fiss`). |
| `FISS-R012` | Error | Разделение знаний | Область `FISS/knowledge/` обязана быть строго разделена на `subject/` и/или `project/`. Размещение свободных файлов непосредственно в `knowledge/` запрещено. |
| `FISS-R013` | Error | Разделение человека | Область `FISS/human/` обязана быть строго разделена на `knowledge/` и/или `hmm/`. Размещение свободных файлов непосредственно в `human/` запрещено. |
| `FISS-R014` | Error | Реестры состояния | Реестры состояния в `FISS/state/` (ADR, риски, открытые вопросы) обязаны оформляться как составные области с `INDEX.md` или одиночные файлы, а не разрозненные файлы в `state/`. |
| `FISS-R015` | Error | Производные знания | Производные знания обязаны содержать точный маркер `Derived from:` со ссылками на существующие файлы-источники. |
| `FISS-R016` | Error | Производность HMM | Все материалы ментальных моделей человека в `FISS/human/hmm/` (и `FISS/human/hmm.md`) обязаны быть явно помечены как производные (`Derived from:`) и ссылаться на источники. |
| `FISS-R017` | Error | Маршрутизация переопределений | Навигация в `FISS/overrides/` обязана быть структурирована по предмету правила, а не по названию инструмента или навыка. |
| `FISS-R018` | Error | Валидность handoff | Запись передачи контекста (handoff) обязана декларировать корректный статус `fiss synchronization` (`synchronized`, `pending` или `unresolved`) и указывать ссылку на задачу. |

---

## Интеграция с Git-хуками

### Хук `pre-push` (рекомендуется)

Хук `pre-push` гарантирует, что невалидное интеллектуальное пространство не попадет в удаленный репозиторий. Создайте исполняемый файл `.git/hooks/pre-push`:

```bash
#!/bin/sh
set -e

echo "Running fiss-lint before push..."
fiss-lint --strict .
```

Сделайте файл исполняемым:
```bash
chmod +x .git/hooks/pre-push
```

### Хук `pre-commit` вручную

Для проверки перед каждым локальным коммитом создайте `.git/hooks/pre-commit`:

```bash
#!/bin/sh
set -e

fiss-lint --strict .
```

```bash
chmod +x .git/hooks/pre-commit
```

### Фреймворк `pre-commit`

Для использования с [pre-commit framework](https://pre-commit.com) добавьте в `.pre-commit-config.yaml`:

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

---

## Интеграция в CI/CD

### GitHub Actions

Пример рабочего процесса в `.github/workflows/fiss-lint.yml`:

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
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Install fiss-lint
        run: |
          curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Run fiss-lint in strict mode
        run: fiss-lint --strict .

      - name: Generate JSON report artifact
        if: always()
        run: fiss-lint --format json . > fiss-report.json || true

      - name: Upload validation report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: fiss-validation-report
          path: fiss-report.json
```

### GitLab CI

Пример конфигурации в `.gitlab-ci.yml`:

```yaml
stages:
  - lint

fiss_lint:
  stage: lint
  image: alpine:latest
  before_script:
    - apk add --no-cache bash curl
    - curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
  script:
    - fiss-lint --strict .
  artifacts:
    when: always
    reports:
      junit: fiss-report.xml
    paths:
      - fiss-report.json
```

---

## Разделение труда в экосистеме FISS

В экосистеме FISS инструменты и агентные навыки имеют строго разграниченную ответственность:

| Инструмент | Роль | Изменяет файлы? | Зона ответственности |
|---|---|:---:|---|
| **`fiss-lint`** | CLI-линтер | **Нет** | Быстрые детерминированные механические проверки структуры, AST-парсинг, ссылочная целостность, синтаксис навигации и обязательные маркеры. |
| **`fiss-validate`** | AI-навык аудита | **Нет** | Качественный семантический аудит на соответствие 7 архитектурным принципам стандарта (7C: Compact, Context-aware, Context-first, Classified, Canonical, Continuous, Composable). |
| **`fiss-init`** | AI-навык инициализации | **Да** | Первичное развёртывание пространства FISS «с нуля» (0-to-1): создание корневых файлов (`INDEX.md`, `BOOTSTRAP.md`), входных точек агентов и регламента непрерывности. |
| **`fiss-maintain`** | AI-навык ведения | **Да** | Ведение пространства в ходе задач (Day 2): управление трёхфазным шлюзом передачи контекста (`Lock` -> `Prepare` -> `Release`), интеграция результатов по 7 направлениям, обновление навигации. |

---

## Лицензия

Проект распространяется на условиях лицензии MIT.

---

Built with ღ and [FISS](https://fiss.vorozhko.ru)
