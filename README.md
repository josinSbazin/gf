# gf - GitFlic CLI

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

`gf` is a command-line tool for working with [GitFlic](https://gitflic.ru) — a Russian code hosting platform. Similar to GitHub's `gh` CLI, it allows you to manage merge requests, pipelines, and repositories directly from your terminal.

[Russian documentation below / Документация на русском ниже](#gf---gitflic-cli-русский)

---

## Features

- **Authentication** — Secure token-based authentication with support for multiple hosts
- **Merge Requests** — List, view, create, and merge MRs
- **Pipelines** — Monitor CI/CD pipelines in real-time
- **Repository Detection** — Automatically detects repository from git remotes
- **Multiple Hosts** — Support for gitflic.ru and self-hosted instances

## Installation

### From Releases (Recommended)

Download the latest binary for your platform from [GitHub Releases](https://github.com/josinSbazin/gf/releases).

**Linux/macOS:**
```bash
# Download (replace VERSION and OS/ARCH as needed)
curl -LO https://github.com/josinSbazin/gf/releases/latest/download/gf_linux_amd64.tar.gz
tar -xzf gf_linux_amd64.tar.gz
sudo mv gf /usr/local/bin/
```

**Windows:**
```powershell
# Download from releases page and add to PATH
# Or use scoop/chocolatey when available
```

### From Source

Requires Go 1.21 or later:

```bash
go install github.com/josinSbazin/gf@latest
```

### Build Manually

```bash
git clone https://github.com/josinSbazin/gf.git
cd gf
go build -o gf .
```

## Quick Start

```bash
# 1. Authenticate with GitFlic
gf auth login

# 2. Navigate to your git repository
cd your-project

# 3. List merge requests
gf mr list

# 4. Watch a pipeline
gf pipeline watch 45
```

## Commands

### Authentication

Manage authentication with GitFlic hosts.

```bash
# Interactive login (opens browser or prompts for token)
gf auth login

# Login to self-hosted GitFlic instance
gf auth login --hostname git.company.com

# Login with token directly
gf auth login --token YOUR_TOKEN

# Login from CI (read token from stdin)
echo $GF_TOKEN | gf auth login --stdin

# Check authentication status
gf auth status

# Check status for specific host
gf auth status --hostname git.company.com
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--hostname` | `-H` | GitFlic hostname (default: gitflic.ru) |
| `--token` | `-t` | Access token |
| `--stdin` | | Read token from standard input |

### Merge Requests

Manage merge requests in your repository.

```bash
# List open merge requests
gf mr list

# List all merge requests
gf mr list --state all

# List merged merge requests
gf mr list --state merged

# List with custom limit
gf mr list --limit 50

# View merge request details
gf mr view 12

# Open merge request in browser
gf mr view 12 --web

# Create merge request (interactive)
gf mr create

# Create with flags
gf mr create --title "Add new feature" --target main

# Create with description from file
gf mr create --title "Fix bug" --body-file description.md

# Merge a merge request
gf mr merge 12

# Merge with squash
gf mr merge 12 --squash

# Merge and delete source branch
gf mr merge 12 --delete-branch

# Merge without confirmation
gf mr merge 12 --yes
```

**List Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--state` | `-s` | Filter by state: open, merged, closed, all (default: open) |
| `--limit` | `-L` | Maximum number of results (default: 30) |
| `--repo` | `-R` | Repository in owner/name format |
| `--json` | | Output as JSON |

**Create Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--title` | `-t` | Merge request title |
| `--body` | `-b` | Merge request description |
| `--body-file` | `-F` | Read description from file |
| `--source` | `-s` | Source branch (default: current branch) |
| `--target` | `-T` | Target branch (default: main/master) |
| `--draft` | `-d` | Create as draft |
| `--repo` | `-R` | Repository in owner/name format |

**Merge Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--squash` | | Squash commits when merging |
| `--delete-branch` | `-d` | Delete source branch after merge |
| `--yes` | `-y` | Skip confirmation prompt |
| `--repo` | `-R` | Repository in owner/name format |

### Pipelines

Monitor CI/CD pipelines.

```bash
# List recent pipelines
gf pipeline list

# List with custom limit
gf pipeline list --limit 10

# View pipeline details and jobs
gf pipeline view 45

# Open pipeline in browser
gf pipeline view 45 --web

# Watch pipeline in real-time (auto-refresh)
gf pipeline watch 45

# Watch with custom refresh interval
gf pipeline watch 45 --interval 5

# Exit with pipeline status code (useful in CI)
gf pipeline watch 45 --exit-status
```

**List Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--limit` | `-L` | Maximum number of results (default: 20) |
| `--repo` | `-R` | Repository in owner/name format |

**Watch Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--interval` | `-i` | Refresh interval in seconds (default: 3) |
| `--exit-status` | | Exit with pipeline status (0=success, 1=failed) |
| `--repo` | `-R` | Repository in owner/name format |

### Repository

View repository information.

```bash
# View current repository
gf repo view

# View specific repository
gf repo view owner/name

# Open in browser
gf repo view --web
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--web` | `-w` | Open in browser |

## Configuration

Configuration is stored in `~/.gf/config.json` (with 0600 permissions for security).

```json
{
  "version": 1,
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "your-api-token",
      "user": "username",
      "protocol": "https"
    },
    "git.company.com": {
      "token": "another-token",
      "user": "username",
      "protocol": "https"
    }
  }
}
```

### Getting an API Token

1. Go to [GitFlic Settings](https://gitflic.ru/settings/oauth/token)
2. Click "Create token"
3. Select required permissions:
   - `read_repository` — for viewing repositories and pipelines
   - `write_repository` — for creating merge requests
   - `api` — for full API access
4. Copy the generated token
5. Run `gf auth login` and paste the token

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GF_TOKEN` | Override authentication token |
| `GF_REPO` | Override repository detection (format: `owner/repo`) |
| `GF_HOST` | Override default host |
| `NO_COLOR` | Disable colored output |

**Example CI usage:**
```yaml
# GitFlic CI
stages:
  - deploy

deploy:
  script:
    - gf auth login --token $GF_TOKEN --stdin
    - gf pipeline watch $CI_PIPELINE_ID --exit-status
```

## Examples

### Create MR from feature branch

```bash
# On feature branch
git checkout -b feature/new-button
# ... make changes ...
git commit -am "Add new button"
git push -u origin feature/new-button

# Create merge request
gf mr create --title "Add new button" --target develop
```

### Wait for CI in scripts

```bash
#!/bin/bash
# Push and wait for pipeline
git push origin main

# Get latest pipeline ID and watch it
PIPELINE_ID=$(gf pipeline list --limit 1 --json | jq '.[0].localId')
gf pipeline watch $PIPELINE_ID --exit-status

if [ $? -eq 0 ]; then
  echo "Pipeline passed!"
else
  echo "Pipeline failed!"
  exit 1
fi
```

### Work with multiple hosts

```bash
# Login to different hosts
gf auth login --hostname gitflic.ru
gf auth login --hostname git.company.com

# Check all auth statuses
gf auth status

# Specify repo explicitly for different host
gf mr list --repo company/project
```

## Troubleshooting

### "could not determine repository"

The tool couldn't detect the repository from git remotes. Solutions:
- Make sure you're in a git repository with a GitFlic remote
- Use `--repo owner/name` flag explicitly
- Set `GF_REPO` environment variable

### "not authenticated"

Run `gf auth login` to authenticate, or check `gf auth status` for existing auth.

### "token expired or invalid"

Your API token may have expired. Generate a new one at GitFlic settings and run `gf auth login` again.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License. See [LICENSE](LICENSE) for details.

---

# gf - GitFlic CLI (Русский)

`gf` — это инструмент командной строки для работы с [GitFlic](https://gitflic.ru) — российской платформой для хостинга кода. Аналогично `gh` CLI для GitHub, он позволяет управлять merge request'ами, пайплайнами и репозиториями прямо из терминала.

---

## Возможности

- **Аутентификация** — Безопасная аутентификация по токену с поддержкой нескольких хостов
- **Merge Request'ы** — Просмотр, создание и слияние MR
- **Пайплайны** — Мониторинг CI/CD пайплайнов в реальном времени
- **Автоопределение репозитория** — Автоматическое определение репозитория по git remote
- **Несколько хостов** — Поддержка gitflic.ru и self-hosted инстансов

## Установка

### Из релизов (рекомендуется)

Скачайте бинарник для вашей платформы из [GitHub Releases](https://github.com/josinSbazin/gf/releases).

**Linux/macOS:**
```bash
# Скачать (замените VERSION и OS/ARCH)
curl -LO https://github.com/josinSbazin/gf/releases/latest/download/gf_linux_amd64.tar.gz
tar -xzf gf_linux_amd64.tar.gz
sudo mv gf /usr/local/bin/
```

**Windows:**
```powershell
# Скачайте со страницы релизов и добавьте в PATH
```

### Из исходников

Требуется Go 1.21 или выше:

```bash
go install github.com/josinSbazin/gf@latest
```

### Сборка вручную

```bash
git clone https://github.com/josinSbazin/gf.git
cd gf
go build -o gf .
```

## Быстрый старт

```bash
# 1. Авторизоваться в GitFlic
gf auth login

# 2. Перейти в git репозиторий
cd your-project

# 3. Посмотреть merge request'ы
gf mr list

# 4. Следить за пайплайном
gf pipeline watch 45
```

## Команды

### Аутентификация

Управление аутентификацией на хостах GitFlic.

```bash
# Интерактивный вход
gf auth login

# Вход на self-hosted GitFlic
gf auth login --hostname git.company.com

# Вход с токеном напрямую
gf auth login --token YOUR_TOKEN

# Вход из CI (токен из stdin)
echo $GF_TOKEN | gf auth login --stdin

# Проверить статус аутентификации
gf auth status

# Проверить статус для конкретного хоста
gf auth status --hostname git.company.com
```

**Флаги:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--hostname` | `-H` | Хост GitFlic (по умолчанию: gitflic.ru) |
| `--token` | `-t` | Токен доступа |
| `--stdin` | | Читать токен из стандартного ввода |

### Merge Request'ы

Управление merge request'ами в репозитории.

```bash
# Список открытых MR
gf mr list

# Список всех MR
gf mr list --state all

# Список влитых MR
gf mr list --state merged

# Список с лимитом
gf mr list --limit 50

# Просмотр деталей MR
gf mr view 12

# Открыть MR в браузере
gf mr view 12 --web

# Создать MR (интерактивно)
gf mr create

# Создать с флагами
gf mr create --title "Добавить фичу" --target main

# Создать с описанием из файла
gf mr create --title "Исправить баг" --body-file description.md

# Влить MR
gf mr merge 12

# Влить со сквошем
gf mr merge 12 --squash

# Влить и удалить ветку
gf mr merge 12 --delete-branch

# Влить без подтверждения
gf mr merge 12 --yes
```

**Флаги list:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--state` | `-s` | Фильтр по статусу: open, merged, closed, all (по умолчанию: open) |
| `--limit` | `-L` | Максимум результатов (по умолчанию: 30) |
| `--repo` | `-R` | Репозиторий в формате owner/name |
| `--json` | | Вывод в JSON |

**Флаги create:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--title` | `-t` | Заголовок MR |
| `--body` | `-b` | Описание MR |
| `--body-file` | `-F` | Читать описание из файла |
| `--source` | `-s` | Исходная ветка (по умолчанию: текущая) |
| `--target` | `-T` | Целевая ветка (по умолчанию: main/master) |
| `--draft` | `-d` | Создать как черновик |
| `--repo` | `-R` | Репозиторий в формате owner/name |

**Флаги merge:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--squash` | | Сквошить коммиты при слиянии |
| `--delete-branch` | `-d` | Удалить ветку после слияния |
| `--yes` | `-y` | Пропустить подтверждение |
| `--repo` | `-R` | Репозиторий в формате owner/name |

### Пайплайны

Мониторинг CI/CD пайплайнов.

```bash
# Список пайплайнов
gf pipeline list

# Список с лимитом
gf pipeline list --limit 10

# Просмотр деталей пайплайна и джобов
gf pipeline view 45

# Открыть пайплайн в браузере
gf pipeline view 45 --web

# Следить за пайплайном в реальном времени
gf pipeline watch 45

# Следить с другим интервалом обновления
gf pipeline watch 45 --interval 5

# Выйти с кодом статуса пайплайна (полезно в CI)
gf pipeline watch 45 --exit-status
```

**Флаги list:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--limit` | `-L` | Максимум результатов (по умолчанию: 20) |
| `--repo` | `-R` | Репозиторий в формате owner/name |

**Флаги watch:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--interval` | `-i` | Интервал обновления в секундах (по умолчанию: 3) |
| `--exit-status` | | Выйти со статусом пайплайна (0=успех, 1=ошибка) |
| `--repo` | `-R` | Репозиторий в формате owner/name |

### Репозиторий

Просмотр информации о репозитории.

```bash
# Просмотр текущего репозитория
gf repo view

# Просмотр конкретного репозитория
gf repo view owner/name

# Открыть в браузере
gf repo view --web
```

**Флаги:**
| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--web` | `-w` | Открыть в браузере |

## Конфигурация

Конфигурация хранится в `~/.gf/config.json` (с правами 0600 для безопасности).

```json
{
  "version": 1,
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "ваш-api-токен",
      "user": "username",
      "protocol": "https"
    },
    "git.company.com": {
      "token": "другой-токен",
      "user": "username",
      "protocol": "https"
    }
  }
}
```

### Получение API токена

1. Перейдите в [Настройки GitFlic](https://gitflic.ru/settings/oauth/token)
2. Нажмите "Создать токен"
3. Выберите необходимые права:
   - `read_repository` — для просмотра репозиториев и пайплайнов
   - `write_repository` — для создания merge request'ов
   - `api` — для полного доступа к API
4. Скопируйте сгенерированный токен
5. Выполните `gf auth login` и вставьте токен

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `GF_TOKEN` | Переопределить токен аутентификации |
| `GF_REPO` | Переопределить определение репозитория (формат: `owner/repo`) |
| `GF_HOST` | Переопределить хост по умолчанию |
| `NO_COLOR` | Отключить цветной вывод |

**Пример использования в CI:**
```yaml
# GitFlic CI
stages:
  - deploy

deploy:
  script:
    - gf auth login --token $GF_TOKEN --stdin
    - gf pipeline watch $CI_PIPELINE_ID --exit-status
```

## Примеры

### Создание MR из feature ветки

```bash
# На feature ветке
git checkout -b feature/new-button
# ... внесите изменения ...
git commit -am "Добавить новую кнопку"
git push -u origin feature/new-button

# Создать merge request
gf mr create --title "Добавить новую кнопку" --target develop
```

### Ожидание CI в скриптах

```bash
#!/bin/bash
# Запушить и подождать пайплайн
git push origin main

# Получить ID последнего пайплайна и следить за ним
PIPELINE_ID=$(gf pipeline list --limit 1 --json | jq '.[0].localId')
gf pipeline watch $PIPELINE_ID --exit-status

if [ $? -eq 0 ]; then
  echo "Пайплайн прошёл!"
else
  echo "Пайплайн упал!"
  exit 1
fi
```

### Работа с несколькими хостами

```bash
# Войти на разные хосты
gf auth login --hostname gitflic.ru
gf auth login --hostname git.company.com

# Проверить статус всех
gf auth status

# Явно указать репозиторий для другого хоста
gf mr list --repo company/project
```

## Решение проблем

### "could not determine repository"

Инструмент не смог определить репозиторий по git remotes. Решения:
- Убедитесь, что вы в git репозитории с GitFlic remote
- Используйте флаг `--repo owner/name` явно
- Установите переменную окружения `GF_REPO`

### "not authenticated"

Выполните `gf auth login` для аутентификации, или проверьте `gf auth status`.

### "token expired or invalid"

Ваш API токен мог истечь. Сгенерируйте новый в настройках GitFlic и выполните `gf auth login` снова.

## Участие в разработке

Мы рады вашему участию! Присылайте issues и pull request'ы.

1. Форкните репозиторий
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Добавить фичу'`)
4. Запушьте ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## Лицензия

MIT License. См. [LICENSE](LICENSE).

---

**Links / Ссылки:**
- GitHub: https://github.com/josinSbazin/gf
- GitFlic: https://gitflic.ru/project/uply-dev/gf
- GitFlic Platform: https://gitflic.ru
