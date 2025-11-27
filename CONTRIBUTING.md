# 🤝 Contributing to Zeno Auth

Спасибо за интерес к проекту! Мы рады любому вкладу.

## 📋 Процесс разработки

### 1. Fork и Clone

```bash
# Fork проекта через GitLab UI
# Затем клонируй свой fork
git clone git@gitlab.com:YOUR_USERNAME/zeno-auth.git
cd zeno-auth

# Добавь upstream remote
git remote add upstream git@gitlab.com:zeno-cy/zeno-auth.git
```

### 2. Создай ветку

```bash
# Обнови main
git checkout main
git pull upstream main

# Создай feature ветку
git checkout -b feature/amazing-feature
```

### 3. Разработка

```bash
# Запусти локальное окружение
make local-up

# Запусти тесты
make test

# Проверь код
make check
```

### 4. Commit

Используй [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add amazing feature"
git commit -m "fix: resolve bug in auth"
git commit -m "docs: update README"
```

**Типы коммитов:**
- `feat:` - новая функциональность
- `fix:` - исправление бага
- `docs:` - изменения в документации
- `style:` - форматирование кода
- `refactor:` - рефакторинг
- `test:` - добавление тестов
- `chore:` - обновление зависимостей, конфигов

### 5. Push и Merge Request

```bash
# Push в свой fork
git push origin feature/amazing-feature

# Создай Merge Request через GitLab UI
```

## ✅ Чеклист перед MR

- [ ] Код отформатирован (`make fmt`)
- [ ] Все тесты проходят (`make test`)
- [ ] Добавлены новые тесты (если нужно)
- [ ] Обновлена документация
- [ ] Нет конфликтов с `main`
- [ ] Commit messages следуют Conventional Commits
- [ ] Нет секретов/credentials в коде

## 🧪 Тестирование

```bash
# Unit тесты
make test

# Integration тесты
make integration

# E2E тесты
make e2e

# Coverage
make cover
```

## 📝 Code Style

- Используй `gofmt` для форматирования
- Следуй [Effective Go](https://golang.org/doc/effective_go)
- Пиши понятные комментарии
- Избегай сложных конструкций

## 🔒 Security

- Никогда не коммить credentials
- Используй `.env` файлы для секретов
- Проверяй код на уязвимости (`make security-scan`)

## 📚 Документация

- Обновляй README при изменении API
- Документируй публичные функции
- Добавляй примеры использования

## 🐛 Reporting Bugs

Используй [Bug template](.gitlab/issue_templates/Bug.md):

1. Опиши проблему
2. Шаги для воспроизведения
3. Ожидаемое поведение
4. Фактическое поведение
5. Логи/скриншоты

## ✨ Feature Requests

Используй [Feature template](.gitlab/issue_templates/Feature.md):

1. Опиши фичу
2. Проблему, которую она решает
3. Предлагаемое решение
4. Альтернативы

## 📞 Контакты

- **Issues**: [gitlab.com/zeno-cy/zeno-auth/issues](https://gitlab.com/zeno-cy/zeno-auth/issues)
- **Merge Requests**: [gitlab.com/zeno-cy/zeno-auth/merge_requests](https://gitlab.com/zeno-cy/zeno-auth/merge_requests)

## 📄 License

Внося вклад, вы соглашаетесь с [MIT License](LICENSE).
