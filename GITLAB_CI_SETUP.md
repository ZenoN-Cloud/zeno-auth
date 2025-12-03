# 🚀 Настройка GitLab CI/CD для автодеплоя

## ⚠️ Org Policy блокирует создание ключей

У тебя включена политика `constraints/iam.managed.disableServiceAccountKeyCreation`.
Это хорошо для безопасности! Но нужно использовать альтернативный подход.

## Вариант 1: Попросить админа создать ключ (временно)

Если у тебя есть доступ к org admin:
```bash
# Отключить политику временно
gcloud resource-manager org-policies disable-enforce \
  iam.disableServiceAccountKeyCreation \
  --project=zeno-cy-dev-001
```

## Вариант 2: Использовать существующий Service Account

Проверь есть ли уже ключи:
```bash
gcloud iam service-accounts keys list \
  --iam-account=gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com
```

## Вариант 3: Деплой вручную (пока)

Пока GitLab CI не настроен, деплой вручную:
```bash
./deploy/gcp-deploy.sh
```

## Шаг 1: Service Account уже создан ✅

```bash
# Создать Service Account
gcloud iam service-accounts create gitlab-ci \
  --display-name="GitLab CI/CD" \
  --description="Service account for GitLab CI/CD pipelines"

# Назначить роли
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.admin"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Создать ключ
gcloud iam service-accounts keys create gitlab-ci-key.json \
  --iam-account=gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com

# Закодировать в base64
cat gitlab-ci-key.json | base64 | tr -d '\n' > gitlab-ci-key-base64.txt

echo "✅ Ключ создан в gitlab-ci-key-base64.txt"
```

## Шаг 2: Добавить переменные в GitLab

Перейди в GitLab:
```
https://gitlab.com/zeno-cy/zeno-auth/-/settings/ci_cd
```

Раздел **Variables** → **Add variable**

### Переменная 1: GCP_SERVICE_ACCOUNT_KEY

- **Key:** `GCP_SERVICE_ACCOUNT_KEY`
- **Value:** Содержимое файла `gitlab-ci-key-base64.txt`
- **Type:** Variable
- **Environment scope:** All
- **Protect variable:** ✅ Yes
- **Mask variable:** ✅ Yes

### Переменная 2: GCP_SERVICE_ACCOUNT_KEY_PROD (для production)

- **Key:** `GCP_SERVICE_ACCOUNT_KEY_PROD`
- **Value:** (создашь позже для prod проекта)
- **Type:** Variable
- **Environment scope:** production
- **Protect variable:** ✅ Yes
- **Mask variable:** ✅ Yes

## Шаг 3: Проверить Artifact Registry

```bash
# Создать репозиторий для Docker образов
gcloud artifacts repositories create zeno-auth \
  --repository-format=docker \
  --location=europe-west3 \
  --description="Docker images for zeno-auth service"

# Проверить
gcloud artifacts repositories list --location=europe-west3
```

## Шаг 4: Проверить Service Account для Cloud Run

```bash
# Проверить что zeno-auth-sa существует
gcloud iam service-accounts describe zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com

# Если нет - создать
gcloud iam service-accounts create zeno-auth-sa \
  --display-name="Zeno Auth Service Account"

# Назначить роли
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:zeno-auth-sa@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## Шаг 5: Тестовый пуш

```bash
# Сделай любое изменение и запуш
git commit --allow-empty -m "test: trigger CI/CD pipeline"
git push gitlab main
```

Проверь pipeline:
```
https://gitlab.com/zeno-cy/zeno-auth/-/pipelines
```

## Что будет происходить:

### При пуше в `main`:
1. ✅ Lint + Tests
2. ✅ Security scans
3. ✅ Build Docker image → Artifact Registry
4. ✅ Deploy на Cloud Run (dev)
5. ✅ Health check

### При пуше в другие ветки:
1. ✅ Lint + Tests
2. ✅ Security scans
3. ❌ Build/Deploy пропускаются

## Troubleshooting

### Ошибка: "Permission denied"
```bash
# Проверь роли
gcloud projects get-iam-policy zeno-cy-dev-001 \
  --flatten="bindings[].members" \
  --filter="bindings.members:gitlab-ci@"
```

### Ошибка: "Artifact Registry not found"
```bash
# Создай репозиторий
gcloud artifacts repositories create zeno-auth \
  --repository-format=docker \
  --location=europe-west3
```

### Ошибка: "Cloud Run service not found"
Это нормально при первом деплое - сервис создастся автоматически.

## Безопасность

⚠️ **ВАЖНО:**
- Файл `gitlab-ci-key.json` добавлен в `.gitignore`
- Никогда не коммить ключи в репозиторий
- После добавления в GitLab удали локальные файлы:
  ```bash
  rm gitlab-ci-key.json gitlab-ci-key-base64.txt
  ```

## Готово! 🎉

После настройки каждый пуш в `main` будет автоматически деплоиться на Cloud Run!
