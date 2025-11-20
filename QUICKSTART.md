# ⚡ Быстрый старт Zeno Auth

## 🚀 Запуск за 30 секунд

```bash
# 1. Запустить все сервисы
make local-up

# 2. Дождаться готовности (10-15 сек)
make local-status

# 3. Протестировать API
make local-test
```

## 📍 Доступ

- **Console (Frontend):** http://localhost:5173
- **API:** http://localhost:8080
- **Health:** http://localhost:8080/health
- **pgAdmin:** http://localhost:5050 (`admin@zeno.local` / `admin`)

## 🛠️ Основные команды

```bash
make local-up        # Запустить
make local-down      # Остановить
make local-logs      # Логи
make local-test      # Тестировать API
make local-clean     # Очистить всё
make local-rebuild   # Пересобрать
```

## 📚 Подробнее

См. [LOCAL_DEV.md](./LOCAL_DEV.md) для полной документации.

---

**Готово! Теперь можно обкатывать локально! 🎉**
