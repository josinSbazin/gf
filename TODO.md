# GF CLI - План реализации

## Реализовано ✅

- [x] `gf auth login` — авторизация
- [x] `gf auth logout` — выход
- [x] `gf auth status` — статус авторизации
- [x] `gf mr list` — список MR
- [x] `gf mr view` — просмотр MR
- [x] `gf mr create` — создание MR
- [x] `gf mr merge` — слияние MR
- [x] `gf mr close` — закрыть MR без слияния
- [x] `gf mr checkout` — checkout ветки MR локально
- [x] `gf pipeline list` — список пайплайнов
- [x] `gf pipeline view` — просмотр пайплайна
- [x] `gf pipeline watch` — отслеживание пайплайна
- [x] `gf repo view` — информация о репозитории
- [x] `gf repo clone` — клонировать репозиторий
- [x] `gf issue list` — список issues
- [x] `gf issue view` — просмотр issue
- [x] `gf issue create` — создание issue
- [x] `gf browse` — открыть в браузере
- [x] DDoS Guard cookie handling

## Запланировано 📋

### Высокий приоритет
- [ ] `gf auth login --oauth` — авторизация через OAuth (с refresh token)
- [x] `gf api` — прямой вызов API (как `gh api`)

### Средний приоритет
- [x] `gf release list` — список релизов
- [x] `gf release view` — просмотр релиза
- [x] `gf release create` — создать релиз

### Низкий приоритет
(API endpoints not available - see "Не поддерживается API" section)

## Не поддерживается API ❌
- `gf issue close` — API GitFlic не поддерживает закрытие issues
- `gf mr comment` — API комментариев к MR не найден
- `gf issue comment` — API комментариев к issues не найден
- `gf repo fork` — API форка не найден
- `gf label list` — API меток не найден
- `gf ssh-key list/add` — API SSH ключей не найден
- `gf search` — API поиска не найден
- `gf snippet create` — API сниппетов не найден

## Безопасность ✅ (исправлено)

### Критические (Phase 1)
- [x] Command Injection fix в `browse.go` — использование безопасного `browser.Open()`
- [x] Path Traversal защита в `clone.go` — валидация директории
- [x] Branch name validation в `checkout.go` — защита от инъекций
- [x] URL encoding в `browse.go` — корректное экранирование branch names
- [x] Race condition fix в `client.go` — double-checked locking с atomic
- [x] Git operations timeout — 2 минуты таймаут для git команд
- [x] HTTP method validation в `api.go` — whitelist разрешенных методов
- [x] CIDR parsing optimization — pre-parsed networks в init()
- [x] Magic strings → constants — `DefaultHostname`, `DefaultAPIBaseURL`

### Средние (Phase 2)
- [x] Interval validation в `watch.go` — min=1s, max=300s
- [x] TTY detection в `watch.go` — ANSI escapes только для терминала
- [x] API timeout в `watch.go` — 30s timeout для API вызовов
- [x] Signal cleanup в `watch.go` — `signal.Stop()` для очистки
- [x] Branch validation в `mr/create.go` — валидация source/target
- [x] Token masking в `status.go` — показ только первых 4 символов
- [x] Smart filtering в `issue.go` — убрано дублирование фильтрации

## Технический долг
- [ ] Тесты для issue команд
- [ ] Тесты для DDoS Guard
- [ ] Тесты для security fixes
- [ ] Документация README
