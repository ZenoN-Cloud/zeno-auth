# Changelog

## 2025-11-16 - Исправления и оптимизация

### Исправлено
- ✅ Интеграционные тесты больше не зависают (добавлены таймауты, правильный docker-compose)
- ✅ Обновлены все GitHub Actions до актуальных версий (v5/v6)
- ✅ Улучшен CORS middleware для работы с фронтендом
- ✅ Валидация конфига теперь не требует DATABASE_URL (для health endpoint)
- ✅ Добавлены автоматические миграции в Docker entrypoint

### Добавлено
- Dockerfile.test для запуска интеграционных тестов
- scripts/entrypoint.sh для автоматических миграций
- scripts/test-endpoints.sh для локального тестирования API
- Таймауты для интеграционных тестов (10 минут)

### Удалено
- Мусорные бинарники (main, zeno-auth, cloud-sql-proxy)
- Пустая директория api/proto
- Устаревшие файлы deploy/cloud-run-dev.yaml и deploy/migrate.sh
- Неправильный scripts/migrate.sh

### Изменено
- Go версия: 1.25 (актуальная)
- GitHub Actions: checkout@v5, setup-go@v6, cache@v4, codecov@v5
- CORS headers: добавлены все необходимые заголовки
- .gitignore: добавлены паттерны для бинарников и coverage
- .dockerignore: оптимизирован для быстрой сборки

### Эндпоинты
- GET /health - проверка здоровья сервиса
- GET /jwks - публичные ключи JWT
- GET /debug - отладочная информация
- POST /auth/register - регистрация пользователя
- POST /auth/login - вход пользователя
- POST /auth/refresh - обновление токена
- POST /auth/logout - выход (требует авторизации)
- GET /me - профиль пользователя (требует авторизации)
