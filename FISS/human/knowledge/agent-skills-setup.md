# Развёртывание агентных навыков и правил проекта

Инструкция по воспроизведению локального окружения агентных навыков (`.agents/skills/`) и правил (`.agents/rules/`) для проекта `fiss-lint` на новой машине.

---

## 1. Архитектурный контекст

Каталог `.agents/` не включён в Git-репозиторий проекта (`.gitignore`), так как содержит локальные симлинки на внешние библиотеки навыков.

Для работы ИИ-агентов в проекте используются два репозитория-донора:
1. **`ai-skills`** — репозиторий проектных навыков и утилит управления фермами симлинков (`make farm`).
2. **`skills-research`** — исследовательский репозиторий внешних навыков и инструкций.

---

## 2. Состав подключаемых компонентов

### Правила разработки (`.agents/rules/`)
Декларативные стандарты кодирования, автоматически действующие при редактировании файлов:

| Файл правила | Источник | Назначение |
| :--- | :--- | :--- |
| `go.md` | `skills-research/repos/awesome-copilot/instructions/go.instructions.md` | Стандарты идиоматичного Go (Effective Go, Google Style Guide, плоский поток, табличные тесты). |
| `markdown.md` | `skills-research/repos/awesome-copilot/instructions/markdown.instructions.md` | Нормативный синтаксис CommonMark 0.31.2 (изоляция fenced blocks, разметка списков и ссылок). |

### Навыки (`.agents/skills/`)
Процедурные сценарии для этапов жизненного цикла задачи:

* **Из репозитория `ai-skills`:**
  * Категория `fiss`: `fiss-maintain`, `fiss-validate`;
  * Категория `task-lifecycle`: `writing-plans`, `executing-plans`, `human-review-surface`, `statement-create`, `intent-exploration`, `pre-analysis`, `brainstorming`, `project-knowledge-refresh`, `subject-knowledge-refresh`, `adr-maintain`, `risk-register`, `glossary-maintain`, `open-questions-maintain`;
  * Категория `common`: `git-commit`, `clear-write`;
  * Категория `it`: `project-review`.

* **Из репозитория `skills-research`:**
  * Язык и чистота кода: `golang-pro`, `clean-code`;
  * Архитектура и дизайн: `software-architecture`, `improve-codebase-architecture`, `api-and-interface-design`, `doubt-driven-development`;
  * Тестирование и верификация: `test-driven-development`, `verification-before-completion`, `systematic-debugging`;
  * Ревью: `code-review-and-quality`, `review-and-simplify-changes`, `receiving-code-review`.

---

## 3. Пошаговая процедура развёртывания

### Предпосылки
На целевой машине должны быть склонированы три репозитория:
```bash
# Пример структуры каталогов рабочего пространства
~/Projects/fiss-lint          # Текущий репозиторий
~/Projects/ai-skills          # Репозиторий навыков
~/Projects/skills-research    # Репозиторий внешних навыков
```

### Шаг 1. Актуализация конфигурации фермы в `ai-skills`
В файле `ai-skills/farms/fiss-lint.env` проверьте и при необходимости скорректируйте путь к целевому проекту:

```bash
TARGET_PROJECT="/home/<user>/Projects/fiss-lint"
```

Если пути к `skills-research` отличаются от стандартных, обновите абсолютные пути в секции `SKILLS` этого `.env`-файла.

### Шаг 2. Развёртывание навыков через ферму `make farm`
Перейдите в репозиторий `ai-skills` и выполните симуляцию, затем генерацию симлинков:

```bash
cd ~/Projects/ai-skills

# Проверка списка планируемых ссылок без изменения файлов
make dry-run PROJECT=fiss-lint

# Создание симлинков в fiss-lint/.agents/skills/
make farm PROJECT=fiss-lint
```

**Контрольная точка:**
Убедитесь, что симлинки созданы корректно:
```bash
make status PROJECT=fiss-lint
```
Все пути должны отображаться со статусом `[OK]`.

### Шаг 3. Создание симлинков правил воркспейса
Перейдите в корень `fiss-lint` и создайте директорию `.agents/rules/` с симлинками на инструкции:

```bash
cd ~/Projects/fiss-lint
mkdir -p .agents/rules

# Создание симлинков на правила Go и Markdown
ln -s ~/Projects/skills-research/repos/awesome-copilot/instructions/go.instructions.md .agents/rules/go.md
ln -s ~/Projects/skills-research/repos/awesome-copilot/instructions/markdown.instructions.md .agents/rules/markdown.md
```

**Контрольная точка:**
```bash
ls -la .agents/rules/
```
Ссылки не должны быть битыми (должны подсвечиваться существующие файлы-источники).

---

## 4. Диагностика и восстановление (Troubleshooting)

* **Симптом:** Симлинки в `.agents/skills/` или `.agents/rules/` подсвечены красным (целевой файл не найден).  
  **Причина:** Абсолютный путь к каталогу `ai-skills` или `skills-research` отличается от указанного в `.env`.  
  **Исправление:** Исправьте пути в `ai-skills/farms/fiss-lint.env`, удалите устаревшие ссылки и пересоздайте их:
  ```bash
  cd ~/Projects/ai-skills
  make clean PROJECT=fiss-lint
  make farm PROJECT=fiss-lint
  ```
