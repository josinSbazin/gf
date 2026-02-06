# gf

CLI для GitFlic: merge request'ы, пайплайны, issues, релизы, ветки, теги, коммиты, файлы, вебхуки.

[English](#english) | [Русский](#русский)

---

## English

### Install

**Via Go (recommended):**
```bash
go install github.com/josinSbazin/gf@latest
```
> Note: `go install` uses GitHub as Go modules registry. The code is mirrored on both GitHub and GitFlic.

**Download binary:**

Download from [GitHub Releases](https://github.com/josinSbazin/gf/releases/latest) or [GitFlic Releases](https://gitflic.ru/project/uply-dev/gf/release/74af6e22-d896-4c44-a8fd-1c6ae261412c):

| Platform | File |
|----------|------|
| Linux x64 | `gf-linux-amd64` |
| macOS Intel | `gf-darwin-amd64` |
| macOS Apple Silicon | `gf-darwin-arm64` |
| Windows x64 | `gf-windows-amd64.exe` |

```bash
# Linux/macOS example
chmod +x gf-linux-amd64
sudo mv gf-linux-amd64 /usr/local/bin/gf
```

**Build from source (GitHub):**
```bash
git clone https://github.com/josinSbazin/gf.git && cd gf && go build -o gf .
```

**Build from source (GitFlic):**
```bash
git clone https://gitflic.ru/project/uply-dev/gf.git && cd gf && go build -o gf .
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
gf issue list
gf status                          # Quick overview of current branch
```

### Commands

#### Auth — login, logout, check status
```bash
gf auth login                      # Prompts for token interactively
gf auth login -t <token>           # Login with token directly
gf auth login -H git.company.com   # Login to self-hosted GitFlic
gf auth login --stdin              # Read token from stdin (for CI)
gf auth status                     # Show current auth: user, host, token preview
gf auth logout                     # Remove saved token
gf auth logout -H git.company.com  # Logout from specific host
```

#### Merge Requests — full MR workflow
```bash
# List and view
gf mr list                         # Open MRs in current repo
gf mr list -s merged               # Filter: open | merged | closed | all
gf mr list -L 50                   # Limit results (default: 30)
gf mr list --json                  # Output as JSON (for scripting)
gf mr view 12                      # Show MR #12 details
gf mr view 12 -w                   # Open MR #12 in browser
gf mr view 12 --json               # Output as JSON

# Create
gf mr create                       # Interactive: prompts for title, branches
gf mr create -t "Fix bug" -T main  # Non-interactive with flags
gf mr create --draft               # Create as draft MR
gf mr create --quiet               # Output only MR ID (for scripts)
gf mr create -w                    # Open in browser after creating

# Actions
gf mr merge 12                     # Merge with confirmation prompt
gf mr merge 12 -y                  # Skip confirmation
gf mr merge 12 --squash -d         # Squash commits + delete source branch
gf mr merge                        # Interactive: select from open MRs
gf mr close 12                     # Close MR without merging
gf mr reopen 12                    # Reopen a closed MR
gf mr approve 12                   # Approve MR
gf mr ready 12                     # Mark draft MR as ready

# Edit
gf mr edit 12 -t "New title"       # Edit MR title
gf mr edit 12 -d "Description"     # Edit MR description
gf mr edit 12 --draft              # Convert to draft
gf mr edit 12 --no-draft           # Remove draft status

# Diff and checkout
gf mr diff 12                      # Show MR diff
gf mr checkout 12                  # Checkout MR source branch locally

# Comments
gf mr comment 12                   # Add comment interactively
gf mr comment 12 -b "LGTM!"        # Add comment with body
gf mr comments 12                  # List all comments/discussions
```

#### Issues — full issue workflow
```bash
# List and view
gf issue list                      # Open issues in current repo
gf issue list -s closed            # Filter: open | closed | all
gf issue list -L 50                # Limit results
gf issue list --json               # Output as JSON
gf issue view 42                   # Show issue #42 details
gf issue view #42                  # Also accepts #ID format
gf issue view 42 -w                # Open in browser
gf issue view 42 --json            # Output as JSON

# Create
gf issue create                    # Interactive: prompts for title
gf issue create -t "Bug report"    # With title
gf issue create -t "Bug" -b "Description here"
gf issue create --quiet            # Output only issue ID

# Actions
gf issue close 42                  # Close issue
gf issue reopen 42                 # Reopen closed issue
gf issue delete 42                 # Delete issue (with confirmation)
gf issue delete 42 -f              # Delete without confirmation

# Edit
gf issue edit 42 -t "New title"    # Edit issue title
gf issue edit 42 -d "Description"  # Edit issue description

# Comments
gf issue comment 42                # Add comment interactively
gf issue comment 42 -b "Fixed!"    # Add comment with body
gf issue comments 42               # List all comments
```

#### Releases — manage releases and assets
```bash
# List and view
gf release list                    # List releases
gf release list -L 10              # Limit results
gf release list --json             # Output as JSON
gf release view v1.0.0             # View release details
gf release view v1.0.0 -w          # Open in browser

# Create
gf release create v1.0.0           # Create release for existing tag
gf release create v1.0.0 -t "Version 1.0" -n "Release notes"
gf release create v1.0.0 -F notes.md     # Read notes from file
gf release create v1.0.0 --draft         # Save as draft
gf release create v1.0.0 --prerelease    # Mark as pre-release
gf release create v1.0.0 --quiet         # Output only tag name

# Edit and delete
gf release edit v1.0.0 -t "New Title"    # Edit release title
gf release edit v1.0.0 --no-draft        # Remove draft status
gf release edit v1.0.0 --prerelease      # Mark as prerelease
gf release delete v1.0.0                 # Delete release (with confirmation)
gf release delete v1.0.0 -f              # Delete without confirmation

# Assets
gf release upload v1.0.0 ./dist/app.zip           # Upload asset
gf release upload v1.0.0 ./build/app -n app.exe   # Upload with custom name
gf release download v1.0.0 --list                 # List assets
gf release download v1.0.0 app.zip                # Download specific asset
gf release download v1.0.0 --all                  # Download all assets
gf release download v1.0.0 app.zip -o ./downloads # Download to path
```

#### Pipelines — CI/CD monitoring and control
```bash
# List and view
gf pipeline list                   # Recent pipelines with status
gf pipeline list --json            # Output as JSON
gf pipeline view 45                # Pipeline #45 details + job list
gf pipeline view #45               # Also accepts #ID format
gf pipeline view 45 --json         # Output as JSON

# Watch
gf pipeline watch 45               # Live updates every 3s, exits when done
gf pipeline watch 45 -i 10         # Custom interval: 10 seconds
gf pipeline watch 45 --exit-status # Exit code 0 if success, 1 if failed

# Actions
gf pipeline retry 45               # Retry/restart pipeline
gf pipeline cancel 45              # Cancel running pipeline
gf pipeline delete 45              # Delete pipeline (with confirmation)
gf pipeline delete 45 -f           # Delete without confirmation

# Jobs
gf pipeline job view 45 1          # View job #1 in pipeline #45
gf pipeline job view 45:1          # Alternative format
gf pipeline job log 45 1           # View job log output
gf pipeline job retry 45 1         # Retry failed job
gf pipeline job cancel 45 1        # Cancel running job
```

#### Branches — manage branches
```bash
gf branch list                     # List all branches
gf branch list --json              # Output as JSON
gf branch create feature/new       # Create branch from default branch
gf branch create hotfix --ref main # Create from specific branch
gf branch delete feature/old       # Delete branch (with confirmation)
gf branch delete feature/old -f    # Delete without confirmation
```

#### Tags — manage tags
```bash
gf tag list                        # List all tags
gf tag list --json                 # Output as JSON
gf tag create v1.0.0               # Create tag at default branch
gf tag create v1.0.0 --ref main    # Create at specific branch/commit
gf tag create v1.0.0 -m "Release"  # Annotated tag with message
gf tag delete v0.9.0               # Delete tag (with confirmation)
gf tag delete v0.9.0 -f            # Delete without confirmation
```

#### Commits — view commit history
```bash
gf commit list                     # List recent commits
gf commit list --ref develop       # Commits on specific branch
gf commit list -L 50               # Limit results
gf commit list --json              # Output as JSON
gf commit view abc1234             # View commit details
gf commit view abc1234 --json      # Output as JSON
gf commit diff abc1234             # Show commit diff
gf commit diff abc1234 --stat      # Show diffstat only
```

#### Files — browse repository files
```bash
gf file list                       # List root directory
gf file list src/                  # List specific directory
gf file list --ref develop         # List on specific branch
gf file list --json                # Output as JSON
gf file view README.md             # View file contents
gf file view src/main.go --ref dev # View on specific branch
gf file download README.md         # Download file
gf file download src/config.json -o ./local/  # Download to path
```

#### Webhooks — manage webhooks
```bash
gf webhook list                    # List all webhooks
gf webhook list --json             # Output as JSON
gf webhook create https://example.com/hook --events push
gf webhook create https://example.com/hook -e push,merge_request,pipeline
gf webhook create https://example.com/hook -e push -s mysecret
gf webhook delete <webhook-id>     # Delete webhook (with confirmation)
gf webhook delete <webhook-id> -f  # Delete without confirmation
gf webhook test <webhook-id>       # Send test payload
```

#### Repository — view, clone
```bash
gf repo view                       # Current repo: visibility, clone URLs
gf repo view owner/name            # Specific repo
gf repo view -w                    # Open in browser

gf repo clone owner/project        # Clone repository
gf repo clone owner/project mydir  # Clone to specific directory
gf repo clone owner/project --ssh  # Clone using SSH instead of HTTPS
```

#### Status — current branch overview
```bash
gf status                          # Show current branch, MR, pipeline status
```

#### Browse — open in browser
```bash
gf browse                          # Open repository in browser
gf browse -b                       # Open current branch
gf browse -s                       # Open settings
gf browse --issues                 # Open issues list
gf browse --mrs                    # Open merge requests list
gf browse -p                       # Open pipelines list

gf browse 42                       # Open issue #42
gf browse --mr 10                  # Open MR #10
```

#### API — direct API calls
```bash
gf api /user/me                    # GET current user
gf api /project                    # List projects

gf api /project/owner/repo/issue -X POST -f title="Bug" -f description="Details"
gf api /endpoint --input body.json # POST with JSON body from file
gf api /endpoint -q .data          # Filter response with jq-like syntax
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

### Shell Completion

```bash
# Bash
gf completion bash > /etc/bash_completion.d/gf

# Zsh
gf completion zsh > "${fpath[1]}/_gf"

# Fish
gf completion fish > ~/.config/fish/completions/gf.fish

# PowerShell
gf completion powershell | Out-String | Invoke-Expression
```

### Environment Variables

Override config without editing files. Useful for CI/CD and scripts.

| Variable | What it does | Example |
|----------|--------------|---------|
| `GF_TOKEN` | Use this token instead of config | `GF_TOKEN=abc123 gf mr list` |
| `GF_REPO` | Override repo detection | `GF_REPO=owner/repo gf pipeline list` |
| `NO_COLOR` | Disable colored output | `NO_COLOR=1 gf mr list` |
| `GF_DEBUG` | Show API request/response details | `GF_DEBUG=1 gf mr create` |

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
PIPELINE_ID=$(gf pipeline list -L 1 --json | jq -r '.[0].localId')
gf pipeline watch $PIPELINE_ID --exit-status || echo "Pipeline failed!"
```

### Flags Reference

| Flag | Short | Used in | Description |
|------|-------|---------|-------------|
| `--repo` | `-R` | all | Repository `owner/name`, overrides git remote detection |
| `--hostname` | `-H` | auth | GitFlic host (default: gitflic.ru) |
| `--web` | `-w` | view, create | Open result in browser |
| `--json` | | list, view | Output as JSON for scripting |
| `--yes` | `-y` | merge | Skip confirmation prompt |
| `--force` | `-f` | delete | Skip confirmation prompt |
| `--limit` | `-L` | list | Max number of results |
| `--state` | `-s` | mr/issue list | Filter: `open` `merged` `closed` `all` |
| `--ref` | | branch/tag/commit/file | Branch, tag, or commit reference |
| `--squash` | | merge | Squash all commits into one |
| `--delete-branch` | `-d` | merge, create | Delete source branch after merge |
| `--interval` | `-i` | watch | Refresh interval in seconds (default: 3) |
| `--exit-status` | | watch | Exit 0 on success, 1 on failure |
| `--token` | `-t` | login | Provide token directly |
| `--stdin` | | login | Read token from stdin (for CI) |
| `--draft` | | mr/release create/edit | Create/mark as draft |
| `--no-draft` | | mr/release edit | Remove draft status |
| `--prerelease` | `-p` | release create | Mark as pre-release |
| `--quiet` | | create | Output only ID (for scripting) |
| `--body` | `-b` | comment | Comment body |
| `--message` | `-m` | tag create | Tag message (annotated tag) |
| `--events` | `-e` | webhook create | Webhook events (comma-separated) |
| `--secret` | `-s` | webhook create | Webhook secret |
| `--name` | `-n` | release upload | Custom asset name |
| `--output` | `-o` | download | Output path |
| `--all` | `-a` | release download | Download all assets |
| `--stat` | | commit diff | Show diffstat only |
| `--mr` | `-m` | browse | Open merge request (with number) |
| `--ssh` | | repo clone | Clone using SSH |
| `--method` | `-X` | api | HTTP method (GET, POST, PUT, DELETE) |
| `--field` | `-f` | api | Add JSON field (key=value) |
| `--jq` | `-q` | api | Filter response with jq expression |

### Command Aliases

| Command | Aliases |
|---------|---------|
| `gf mr` | `gf merge-request` |
| `gf issue` | `gf i` |
| `gf pipeline` | `gf pl`, `gf ci` |
| `gf release` | `gf rel` |
| `gf repo` | `gf r` |
| `gf branch` | `gf br` |
| `gf tag` | `gf t` |
| `gf commit` | `gf c` |
| `gf file` | `gf f`, `gf blob` |
| `gf webhook` | `gf hook` |

---

## Русский

### Установка

**Через Go (рекомендуется):**
```bash
go install github.com/josinSbazin/gf@latest
```
> Примечание: `go install` использует GitHub как реестр Go-модулей. Код дублируется на GitHub и GitFlic.

**Скачать бинарник:**

Скачайте с [GitHub Releases](https://github.com/josinSbazin/gf/releases/latest) или [GitFlic Releases](https://gitflic.ru/project/uply-dev/gf/release/74af6e22-d896-4c44-a8fd-1c6ae261412c):

| Платформа | Файл |
|-----------|------|
| Linux x64 | `gf-linux-amd64` |
| macOS Intel | `gf-darwin-amd64` |
| macOS Apple Silicon | `gf-darwin-arm64` |
| Windows x64 | `gf-windows-amd64.exe` |

```bash
# Пример для Linux/macOS
chmod +x gf-linux-amd64
sudo mv gf-linux-amd64 /usr/local/bin/gf
```

**Сборка из исходников (GitHub):**
```bash
git clone https://github.com/josinSbazin/gf.git && cd gf && go build -o gf .
```

**Сборка из исходников (GitFlic):**
```bash
git clone https://gitflic.ru/project/uply-dev/gf.git && cd gf && go build -o gf .
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
gf issue list
gf status                          # Обзор текущей ветки
```

### Команды

#### Auth — вход, выход, статус
```bash
gf auth login                      # Запросит токен интерактивно
gf auth login -t <token>           # Вход с указанием токена
gf auth login -H git.company.com   # Вход на self-hosted GitFlic
gf auth login --stdin              # Читать токен из stdin (для CI)
gf auth status                     # Текущий статус: пользователь, хост
gf auth logout                     # Удалить сохранённый токен
gf auth logout -H git.company.com  # Выйти с конкретного хоста
```

#### Merge Requests — полный workflow
```bash
# Список и просмотр
gf mr list                         # Открытые MR в текущем репозитории
gf mr list -s merged               # Фильтр: open | merged | closed | all
gf mr list -L 50                   # Лимит результатов (по умолчанию: 30)
gf mr list --json                  # Вывод в JSON (для скриптов)
gf mr view 12                      # Детали MR #12
gf mr view 12 -w                   # Открыть MR #12 в браузере
gf mr view 12 --json               # Вывод в JSON

# Создание
gf mr create                       # Интерактивно: запросит title, ветки
gf mr create -t "Fix bug" -T main  # С флагами, без интерактива
gf mr create --draft               # Создать как черновик
gf mr create --quiet               # Вывести только ID (для скриптов)
gf mr create -w                    # Открыть в браузере после создания

# Действия
gf mr merge 12                     # Слить с подтверждением
gf mr merge 12 -y                  # Без подтверждения
gf mr merge 12 --squash -d         # Сквош + удалить ветку
gf mr merge                        # Интерактивно: выбрать из открытых MR
gf mr close 12                     # Закрыть MR без слияния
gf mr reopen 12                    # Переоткрыть закрытый MR
gf mr approve 12                   # Одобрить MR
gf mr ready 12                     # Пометить draft MR как готовый

# Редактирование
gf mr edit 12 -t "Новый заголовок" # Изменить title
gf mr edit 12 -d "Описание"        # Изменить description
gf mr edit 12 --draft              # Сделать черновиком
gf mr edit 12 --no-draft           # Убрать статус черновика

# Diff и checkout
gf mr diff 12                      # Показать diff MR
gf mr checkout 12                  # Checkout ветки MR локально

# Комментарии
gf mr comment 12                   # Добавить комментарий интерактивно
gf mr comment 12 -b "LGTM!"        # Добавить комментарий
gf mr comments 12                  # Список всех комментариев
```

#### Issues — полный workflow
```bash
# Список и просмотр
gf issue list                      # Открытые issues в репозитории
gf issue list -s closed            # Фильтр: open | closed | all
gf issue list -L 50                # Лимит результатов
gf issue list --json               # Вывод в JSON
gf issue view 42                   # Детали issue #42
gf issue view #42                  # Также принимает формат #ID
gf issue view 42 -w                # Открыть в браузере
gf issue view 42 --json            # Вывод в JSON

# Создание
gf issue create                    # Интерактивно: запросит title
gf issue create -t "Bug report"    # С указанием title
gf issue create -t "Bug" -b "Описание"
gf issue create --quiet            # Вывести только ID

# Действия
gf issue close 42                  # Закрыть issue
gf issue reopen 42                 # Переоткрыть issue
gf issue delete 42                 # Удалить (с подтверждением)
gf issue delete 42 -f              # Удалить без подтверждения

# Редактирование
gf issue edit 42 -t "Новый title"  # Изменить title
gf issue edit 42 -d "Описание"     # Изменить description

# Комментарии
gf issue comment 42                # Добавить комментарий интерактивно
gf issue comment 42 -b "Исправлено!" # Добавить комментарий
gf issue comments 42               # Список всех комментариев
```

#### Releases — релизы и assets
```bash
# Список и просмотр
gf release list                    # Список релизов
gf release list -L 10              # Лимит результатов
gf release list --json             # Вывод в JSON
gf release view v1.0.0             # Детали релиза
gf release view v1.0.0 -w          # Открыть в браузере

# Создание
gf release create v1.0.0           # Создать для существующего тега
gf release create v1.0.0 -t "Version 1.0" -n "Release notes"
gf release create v1.0.0 -F notes.md     # Прочитать notes из файла
gf release create v1.0.0 --draft         # Сохранить как черновик
gf release create v1.0.0 --prerelease    # Пометить как pre-release
gf release create v1.0.0 --quiet         # Вывести только имя тега

# Редактирование и удаление
gf release edit v1.0.0 -t "Новое имя"    # Изменить title
gf release edit v1.0.0 --no-draft        # Убрать статус черновика
gf release edit v1.0.0 --prerelease      # Пометить как prerelease
gf release delete v1.0.0                 # Удалить (с подтверждением)
gf release delete v1.0.0 -f              # Удалить без подтверждения

# Assets
gf release upload v1.0.0 ./dist/app.zip           # Загрузить asset
gf release upload v1.0.0 ./build/app -n app.exe   # С кастомным именем
gf release download v1.0.0 --list                 # Список assets
gf release download v1.0.0 app.zip                # Скачать asset
gf release download v1.0.0 --all                  # Скачать все assets
gf release download v1.0.0 app.zip -o ./downloads # Скачать в путь
```

#### Pipelines — мониторинг и управление CI/CD
```bash
# Список и просмотр
gf pipeline list                   # Последние пайплайны со статусом
gf pipeline list --json            # Вывод в JSON
gf pipeline view 45                # Детали пайплайна #45 + джобы
gf pipeline view #45               # Также принимает формат #ID
gf pipeline view 45 --json         # Вывод в JSON

# Отслеживание
gf pipeline watch 45               # Обновления каждые 3 сек
gf pipeline watch 45 -i 10         # Интервал 10 секунд
gf pipeline watch 45 --exit-status # Код выхода: 0=успех, 1=ошибка

# Действия
gf pipeline retry 45               # Перезапустить пайплайн
gf pipeline cancel 45              # Отменить запущенный пайплайн
gf pipeline delete 45              # Удалить (с подтверждением)
gf pipeline delete 45 -f           # Удалить без подтверждения

# Джобы
gf pipeline job view 45 1          # Просмотр джоба #1 в пайплайне #45
gf pipeline job view 45:1          # Альтернативный формат
gf pipeline job log 45 1           # Просмотр лога джоба
gf pipeline job retry 45 1         # Перезапустить джоб
gf pipeline job cancel 45 1        # Отменить джоб
```

#### Branches — управление ветками
```bash
gf branch list                     # Список всех веток
gf branch list --json              # Вывод в JSON
gf branch create feature/new       # Создать от default ветки
gf branch create hotfix --ref main # Создать от конкретной ветки
gf branch delete feature/old       # Удалить (с подтверждением)
gf branch delete feature/old -f    # Удалить без подтверждения
```

#### Tags — управление тегами
```bash
gf tag list                        # Список всех тегов
gf tag list --json                 # Вывод в JSON
gf tag create v1.0.0               # Создать на default ветке
gf tag create v1.0.0 --ref main    # Создать на конкретной ветке
gf tag create v1.0.0 -m "Релиз"    # Аннотированный тег с сообщением
gf tag delete v0.9.0               # Удалить (с подтверждением)
gf tag delete v0.9.0 -f            # Удалить без подтверждения
```

#### Commits — история коммитов
```bash
gf commit list                     # Список последних коммитов
gf commit list --ref develop       # Коммиты на конкретной ветке
gf commit list -L 50               # Лимит результатов
gf commit list --json              # Вывод в JSON
gf commit view abc1234             # Детали коммита
gf commit view abc1234 --json      # Вывод в JSON
gf commit diff abc1234             # Diff коммита
gf commit diff abc1234 --stat      # Только статистика
```

#### Files — просмотр файлов репозитория
```bash
gf file list                       # Список корневой директории
gf file list src/                  # Список конкретной директории
gf file list --ref develop         # Список на конкретной ветке
gf file list --json                # Вывод в JSON
gf file view README.md             # Содержимое файла
gf file view src/main.go --ref dev # На конкретной ветке
gf file download README.md         # Скачать файл
gf file download src/config.json -o ./local/  # Скачать в путь
```

#### Webhooks — управление вебхуками
```bash
gf webhook list                    # Список вебхуков
gf webhook list --json             # Вывод в JSON
gf webhook create https://example.com/hook --events push
gf webhook create https://example.com/hook -e push,merge_request,pipeline
gf webhook create https://example.com/hook -e push -s mysecret
gf webhook delete <webhook-id>     # Удалить (с подтверждением)
gf webhook delete <webhook-id> -f  # Удалить без подтверждения
gf webhook test <webhook-id>       # Отправить тестовый payload
```

#### Repository — просмотр, клонирование
```bash
gf repo view                       # Текущий репо: видимость, URL клонирования
gf repo view owner/name            # Конкретный репозиторий
gf repo view -w                    # Открыть в браузере

gf repo clone owner/project        # Клонировать репозиторий
gf repo clone owner/project mydir  # Клонировать в указанную директорию
gf repo clone owner/project --ssh  # Клонировать через SSH
```

#### Status — обзор текущей ветки
```bash
gf status                          # Текущая ветка, MR, статус пайплайна
```

#### Browse — открыть в браузере
```bash
gf browse                          # Открыть репозиторий в браузере
gf browse -b                       # Открыть текущую ветку
gf browse -s                       # Открыть настройки
gf browse --issues                 # Открыть список issues
gf browse --mrs                    # Открыть список merge requests
gf browse -p                       # Открыть список пайплайнов

gf browse 42                       # Открыть issue #42
gf browse --mr 10                  # Открыть MR #10
```

#### API — прямые вызовы API
```bash
gf api /user/me                    # GET текущего пользователя
gf api /project                    # Список проектов

gf api /project/owner/repo/issue -X POST -f title="Bug" -f description="Details"
gf api /endpoint --input body.json # POST с JSON телом из файла
gf api /endpoint -q .data          # Фильтровать ответ jq выражением
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

### Автодополнение

```bash
# Bash
gf completion bash > /etc/bash_completion.d/gf

# Zsh
gf completion zsh > "${fpath[1]}/_gf"

# Fish
gf completion fish > ~/.config/fish/completions/gf.fish

# PowerShell
gf completion powershell | Out-String | Invoke-Expression
```

### Переменные окружения

| Переменная | Что делает | Пример |
|------------|------------|--------|
| `GF_TOKEN` | Использовать этот токен вместо конфига | `GF_TOKEN=abc123 gf mr list` |
| `GF_REPO` | Переопределить определение репозитория | `GF_REPO=owner/repo gf pipeline list` |
| `NO_COLOR` | Отключить цветной вывод | `NO_COLOR=1 gf mr list` |
| `GF_DEBUG` | Показать детали API запросов/ответов | `GF_DEBUG=1 gf mr create` |

### Справочник флагов

| Флаг | Сокр. | Где | Описание |
|------|-------|-----|----------|
| `--repo` | `-R` | везде | Репозиторий `owner/name` |
| `--hostname` | `-H` | auth | Хост GitFlic |
| `--web` | `-w` | view, create | Открыть в браузере |
| `--json` | | list, view | Вывод в JSON |
| `--force` | `-f` | delete | Пропустить подтверждение |
| `--limit` | `-L` | list | Максимум результатов |
| `--state` | `-s` | mr/issue list | Фильтр: open/closed/all |
| `--ref` | | branch/tag/commit/file | Ветка/тег/коммит |
| `--draft` | | mr/release | Черновик |
| `--quiet` | | create | Только ID |
| `--body` | `-b` | comment | Текст комментария |
| `--message` | `-m` | tag create | Сообщение тега |
| `--events` | `-e` | webhook | События webhook |

### Алиасы команд

| Команда | Алиасы |
|---------|--------|
| `gf mr` | `gf merge-request` |
| `gf issue` | `gf i` |
| `gf pipeline` | `gf pl`, `gf ci` |
| `gf release` | `gf rel` |
| `gf repo` | `gf r` |
| `gf branch` | `gf br` |
| `gf tag` | `gf t` |
| `gf commit` | `gf c` |
| `gf file` | `gf f`, `gf blob` |
| `gf webhook` | `gf hook` |

---

## Changelog

### v0.2.0

**Bug Fixes:**
- **webhook create**: Fixed 422 error - API expects `events` as object with boolean flags, not array; auto-generate secret if not provided
- **webhook list**: Fixed empty list - corrected JSON key from `webhookModelList` to `webhookList`
- **issue create**: Fixed 422 error when `--body ""` - GitFlic requires non-empty description, now auto-fills with "No description provided"
- **release download**: Fixed 404 error - assets are now fetched from release's `attachmentFiles` field instead of non-existent endpoint
- **release view/upload**: Fixed wrong release returned - API returns partial matches, now filters for exact tagName
- **tag create --ref**: Added clear error message for short commit hashes (API requires full 40-character hash)
- **tag delete / branch delete**: Use `git push --delete` fallback since GitFlic API returns 405

**Improvements:**
- Better error messages for webhook events validation
- Auto-generated webhook secrets are displayed after creation
- Improved "no assets" message with helpful instructions

### v0.1.0

Initial release with full GitFlic CLI functionality:
- Auth, MR, Issues, Releases, Pipelines, Branches, Tags, Commits, Files, Webhooks
- JSON output for scripting
- Shell completion
- Multi-host support

---

## License

MIT
