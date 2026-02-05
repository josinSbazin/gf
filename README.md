# gf

CLI для работы с GitFlic: merge request'ы, пайплайны, репозитории.

[English](#english) | [Русский](#русский)

---

## English

### Install

```bash
# Go install
go install github.com/josinSbazin/gf@latest

# Or build from source
git clone https://github.com/josinSbazin/gf.git
cd gf && go build -o gf .
```

### Setup

```bash
# Get token: https://gitflic.ru/settings/oauth/token
gf auth login
```

### Commands

#### Auth

```bash
gf auth login                      # Interactive login
gf auth login -H git.company.com   # Self-hosted instance
gf auth login -t TOKEN             # With token
gf auth status                     # Check auth
```

#### Merge Requests

```bash
gf mr list                         # List open MRs
gf mr list -s all                  # All MRs (open/merged/closed)
gf mr list -s merged               # Only merged
gf mr view 12                      # View MR #12
gf mr view 12 -w                   # Open in browser
gf mr create                       # Create MR (interactive)
gf mr create -t "Title" -T main    # Create with flags
gf mr merge 12                     # Merge MR #12
gf mr merge 12 --squash            # Squash merge
gf mr merge 12 -d                  # Merge + delete branch
gf mr merge 12 -y                  # Skip confirmation
```

#### Pipelines

```bash
gf pipeline list                   # List pipelines
gf pipeline view 45                # View pipeline #45 + jobs
gf pipeline watch 45               # Watch in real-time
gf pipeline watch 45 -i 5          # Custom interval (5s)
gf pipeline watch 45 --exit-status # Exit with pipeline status code
```

#### Repository

```bash
gf repo view                       # Current repo info
gf repo view owner/name            # Specific repo
gf repo view -w                    # Open in browser
```

### Flags Reference

| Flag | Short | Description |
|------|-------|-------------|
| `--repo` | `-R` | Repository `owner/name` (overrides auto-detect) |
| `--hostname` | `-H` | GitFlic host (default: gitflic.ru) |
| `--web` | `-w` | Open in browser |
| `--yes` | `-y` | Skip confirmation |
| `--limit` | `-L` | Max results |
| `--state` | `-s` | Filter: open/merged/closed/all |
| `--squash` | | Squash commits on merge |
| `--delete-branch` | `-d` | Delete branch after merge |
| `--interval` | `-i` | Watch refresh interval (seconds) |
| `--exit-status` | | Exit with pipeline status (0/1) |

### Config

Location: `~/.gf/config.json`

```json
{
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "xxx",
      "user": "username"
    }
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GF_TOKEN` | Override token |
| `GF_REPO` | Override repo (`owner/name`) |
| `NO_COLOR` | Disable colors |

### CI Example

```yaml
deploy:
  script:
    - echo $GF_TOKEN | gf auth login --stdin
    - gf pipeline watch $CI_PIPELINE_ID --exit-status
```

---

## Русский

### Установка

```bash
# Go install
go install github.com/josinSbazin/gf@latest

# Или сборка из исходников
git clone https://github.com/josinSbazin/gf.git
cd gf && go build -o gf .
```

### Настройка

```bash
# Получить токен: https://gitflic.ru/settings/oauth/token
gf auth login
```

### Команды

#### Аутентификация

```bash
gf auth login                      # Интерактивный вход
gf auth login -H git.company.com   # Self-hosted
gf auth login -t TOKEN             # С токеном
gf auth status                     # Проверить статус
```

#### Merge Request'ы

```bash
gf mr list                         # Открытые MR
gf mr list -s all                  # Все MR
gf mr list -s merged               # Только влитые
gf mr view 12                      # Просмотр MR #12
gf mr view 12 -w                   # Открыть в браузере
gf mr create                       # Создать MR (интерактивно)
gf mr create -t "Заголовок" -T main # Создать с флагами
gf mr merge 12                     # Влить MR #12
gf mr merge 12 --squash            # Сквош
gf mr merge 12 -d                  # Влить + удалить ветку
gf mr merge 12 -y                  # Без подтверждения
```

#### Пайплайны

```bash
gf pipeline list                   # Список пайплайнов
gf pipeline view 45                # Просмотр #45 + джобы
gf pipeline watch 45               # Следить в реальном времени
gf pipeline watch 45 -i 5          # Интервал 5 сек
gf pipeline watch 45 --exit-status # Выйти с кодом статуса
```

#### Репозиторий

```bash
gf repo view                       # Текущий репозиторий
gf repo view owner/name            # Конкретный репозиторий
gf repo view -w                    # Открыть в браузере
```

### Флаги

| Флаг | Сокр. | Описание |
|------|-------|----------|
| `--repo` | `-R` | Репозиторий `owner/name` |
| `--hostname` | `-H` | Хост GitFlic (по умолчанию: gitflic.ru) |
| `--web` | `-w` | Открыть в браузере |
| `--yes` | `-y` | Без подтверждения |
| `--limit` | `-L` | Лимит результатов |
| `--state` | `-s` | Фильтр: open/merged/closed/all |
| `--squash` | | Сквош коммитов |
| `--delete-branch` | `-d` | Удалить ветку после влития |
| `--interval` | `-i` | Интервал обновления (сек) |
| `--exit-status` | | Выйти с кодом статуса пайплайна |

### Конфигурация

Путь: `~/.gf/config.json`

```json
{
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "xxx",
      "user": "username"
    }
  }
}
```

### Переменные окружения

| Переменная | Описание |
|------------|----------|
| `GF_TOKEN` | Переопределить токен |
| `GF_REPO` | Переопределить репозиторий |
| `NO_COLOR` | Отключить цвета |

### Пример для CI

```yaml
deploy:
  script:
    - echo $GF_TOKEN | gf auth login --stdin
    - gf pipeline watch $CI_PIPELINE_ID --exit-status
```

---

## License

MIT
