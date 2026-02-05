# gf

CLI для GitFlic: merge request'ы, пайплайны, репозитории.

[English](#english) | [Русский](#русский)

---

## English

### Install

**Via Go (recommended):**
```bash
go install github.com/josinSbazin/gf@latest
```
> Note: `go install` uses GitHub as Go modules registry. The code is mirrored on both GitHub and GitFlic.

**Build from source (GitHub):**
```bash
git clone https://github.com/josinSbazin/gf.git && cd gf && go build -o gf .
```

**Build from source (GitFlic):**
```bash
git clone https://gitflic.ru/project/josinsbazin/gf.git && cd gf && go build -o gf .
```

### Quick Start

```bash
# 1. Get API token at https://gitflic.ru/settings/oauth/token
# 2. Login
gf auth login

# 3. Go to your repo and use
cd my-project
gf mr list
gf pipeline list
```

### Commands

**Auth** — login and check status:
```bash
gf auth login                      # Prompts for token interactively
gf auth login -t <token>           # Login with token directly
gf auth login -H git.company.com   # Login to self-hosted GitFlic
gf auth status                     # Show current auth: user, host, token preview
```

**Merge Requests** — list, view, create, merge:
```bash
gf mr list                         # Open MRs in current repo
gf mr list -s merged               # Filter: open | merged | closed | all
gf mr list -L 50                   # Limit results (default: 30)

gf mr view 12                      # Show MR #12 details
gf mr view 12 -w                   # Open MR #12 in browser

gf mr create                       # Interactive: prompts for title, branches
gf mr create -t "Fix bug" -T main  # Non-interactive with flags

gf mr merge 12                     # Merge with confirmation prompt
gf mr merge 12 -y                  # Skip confirmation
gf mr merge 12 --squash -d         # Squash commits + delete source branch
```

**Pipelines** — monitor CI/CD:
```bash
gf pipeline list                   # Recent pipelines with status
gf pipeline view 45                # Pipeline #45 details + job list

gf pipeline watch 45               # Live updates every 3s, exits when done
gf pipeline watch 45 -i 10         # Custom interval: 10 seconds
gf pipeline watch 45 --exit-status # Exit code 0 if success, 1 if failed
```

**Repository** — view info:
```bash
gf repo view                       # Current repo: visibility, clone URLs
gf repo view owner/name            # Specific repo
gf repo view -w                    # Open in browser
```

### Configuration

**File:** `~/.gf/config.json` (created on first `gf auth login`)

```json
{
  "version": 1,
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "d3a84de8-7738-4018-...",
      "user": "myusername",
      "protocol": "https"
    },
    "git.company.com": {
      "token": "another-token",
      "user": "workuser",
      "protocol": "https"
    }
  }
}
```

**Fields:**
- `active_host` — default host when `-H` not specified
- `hosts` — map of host → credentials
- `token` — API access token (get at Settings → API Tokens)
- `user` — your username (saved automatically on login)

**Multiple hosts:** login to each, then use `-H` to switch or `-R` for specific repo:
```bash
gf auth login -H gitflic.ru
gf auth login -H git.company.com
gf mr list -R company-org/project  # Uses git.company.com
```

### Environment Variables

Override config without editing files. Useful for CI/CD and scripts.

| Variable | What it does | Example |
|----------|--------------|---------|
| `GF_TOKEN` | Use this token instead of config | `GF_TOKEN=abc123 gf mr list` |
| `GF_REPO` | Override repo detection | `GF_REPO=owner/repo gf pipeline list` |
| `NO_COLOR` | Disable colored output | `NO_COLOR=1 gf mr list` |

**CI/CD example** — GitFlic CI:
```yaml
variables:
  GF_TOKEN: $CI_JOB_TOKEN  # or use Settings → CI Variables

build:
  script:
    - echo "$GF_TOKEN" | gf auth login --stdin
    - gf pipeline watch $CI_PIPELINE_ID --exit-status
```

**Script example** — wait for pipeline after push:
```bash
#!/bin/bash
git push origin feature/my-branch
sleep 5  # wait for pipeline to start
gf pipeline list -L 1  # see latest pipeline ID
gf pipeline watch 123 --exit-status || echo "Pipeline failed!"
```

### Flags Reference

| Flag | Short | Used in | Description |
|------|-------|---------|-------------|
| `--repo` | `-R` | all | Repository `owner/name`, overrides git remote detection |
| `--hostname` | `-H` | auth | GitFlic host (default: gitflic.ru) |
| `--web` | `-w` | view | Open result in browser |
| `--yes` | `-y` | merge | Skip confirmation prompt |
| `--limit` | `-L` | list | Max number of results |
| `--state` | `-s` | mr list | Filter: `open` `merged` `closed` `all` |
| `--squash` | | merge | Squash all commits into one |
| `--delete-branch` | `-d` | merge | Delete source branch after merge |
| `--interval` | `-i` | watch | Refresh interval in seconds (default: 3) |
| `--exit-status` | | watch | Exit 0 on success, 1 on failure |
| `--token` | `-t` | login | Provide token directly |
| `--stdin` | | login | Read token from stdin (for CI) |

---

## Русский

### Установка

**Через Go (рекомендуется):**
```bash
go install github.com/josinSbazin/gf@latest
```
> Примечание: `go install` использует GitHub как реестр Go-модулей. Код дублируется на GitHub и GitFlic.

**Сборка из исходников (GitHub):**
```bash
git clone https://github.com/josinSbazin/gf.git && cd gf && go build -o gf .
```

**Сборка из исходников (GitFlic):**
```bash
git clone https://gitflic.ru/project/josinsbazin/gf.git && cd gf && go build -o gf .
```

### Быстрый старт

```bash
# 1. Получить токен: https://gitflic.ru/settings/oauth/token
# 2. Залогиниться
gf auth login

# 3. Перейти в репозиторий и использовать
cd my-project
gf mr list
gf pipeline list
```

### Команды

**Auth** — вход и проверка статуса:
```bash
gf auth login                      # Запросит токен интерактивно
gf auth login -t <token>           # Вход с указанием токена
gf auth login -H git.company.com   # Вход на self-hosted GitFlic
gf auth status                     # Текущий статус: пользователь, хост
```

**Merge Requests** — список, просмотр, создание, слияние:
```bash
gf mr list                         # Открытые MR в текущем репозитории
gf mr list -s merged               # Фильтр: open | merged | closed | all
gf mr list -L 50                   # Лимит результатов (по умолчанию: 30)

gf mr view 12                      # Детали MR #12
gf mr view 12 -w                   # Открыть MR #12 в браузере

gf mr create                       # Интерактивно: запросит title, ветки
gf mr create -t "Fix bug" -T main  # С флагами, без интерактива

gf mr merge 12                     # Слить с подтверждением
gf mr merge 12 -y                  # Без подтверждения
gf mr merge 12 --squash -d         # Сквош + удалить ветку
```

**Pipelines** — мониторинг CI/CD:
```bash
gf pipeline list                   # Последние пайплайны со статусом
gf pipeline view 45                # Детали пайплайна #45 + список джобов

gf pipeline watch 45               # Обновления каждые 3 сек, выход по завершении
gf pipeline watch 45 -i 10         # Интервал 10 секунд
gf pipeline watch 45 --exit-status # Код выхода: 0=успех, 1=ошибка
```

**Repository** — информация:
```bash
gf repo view                       # Текущий репо: видимость, URL клонирования
gf repo view owner/name            # Конкретный репозиторий
gf repo view -w                    # Открыть в браузере
```

### Конфигурация

**Файл:** `~/.gf/config.json` (создаётся при первом `gf auth login`)

```json
{
  "version": 1,
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "d3a84de8-7738-4018-...",
      "user": "myusername",
      "protocol": "https"
    },
    "git.company.com": {
      "token": "another-token",
      "user": "workuser",
      "protocol": "https"
    }
  }
}
```

**Поля:**
- `active_host` — хост по умолчанию, когда `-H` не указан
- `hosts` — словарь хост → учётные данные
- `token` — API токен (получить в Настройки → API Токены)
- `user` — ваш username (сохраняется автоматически при входе)

**Несколько хостов:** залогиньтесь в каждый, потом `-H` для переключения:
```bash
gf auth login -H gitflic.ru
gf auth login -H git.company.com
gf mr list -R company-org/project  # Использует git.company.com
```

### Переменные окружения

Переопределяют конфиг без редактирования файлов. Полезно для CI/CD и скриптов.

| Переменная | Что делает | Пример |
|------------|------------|--------|
| `GF_TOKEN` | Использовать этот токен вместо конфига | `GF_TOKEN=abc123 gf mr list` |
| `GF_REPO` | Переопределить определение репозитория | `GF_REPO=owner/repo gf pipeline list` |
| `NO_COLOR` | Отключить цветной вывод | `NO_COLOR=1 gf mr list` |

**Пример CI/CD** — GitFlic CI:
```yaml
variables:
  GF_TOKEN: $CI_JOB_TOKEN  # или через Settings → CI Variables

build:
  script:
    - echo "$GF_TOKEN" | gf auth login --stdin
    - gf pipeline watch $CI_PIPELINE_ID --exit-status
```

**Пример скрипта** — ждать пайплайн после push:
```bash
#!/bin/bash
git push origin feature/my-branch
sleep 5  # подождать запуск пайплайна
gf pipeline list -L 1  # увидеть ID последнего
gf pipeline watch 123 --exit-status || echo "Pipeline failed!"
```

### Справочник флагов

| Флаг | Сокр. | Где | Описание |
|------|-------|-----|----------|
| `--repo` | `-R` | везде | Репозиторий `owner/name`, переопределяет автоопределение |
| `--hostname` | `-H` | auth | Хост GitFlic (по умолчанию: gitflic.ru) |
| `--web` | `-w` | view | Открыть в браузере |
| `--yes` | `-y` | merge | Пропустить подтверждение |
| `--limit` | `-L` | list | Максимум результатов |
| `--state` | `-s` | mr list | Фильтр: `open` `merged` `closed` `all` |
| `--squash` | | merge | Объединить коммиты в один |
| `--delete-branch` | `-d` | merge | Удалить ветку после слияния |
| `--interval` | `-i` | watch | Интервал обновления в секундах (по умолчанию: 3) |
| `--exit-status` | | watch | Код выхода: 0=успех, 1=ошибка |
| `--token` | `-t` | login | Указать токен напрямую |
| `--stdin` | | login | Читать токен из stdin (для CI) |

---

## License

MIT
