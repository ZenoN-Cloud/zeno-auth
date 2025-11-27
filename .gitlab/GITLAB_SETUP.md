# 🦊 GitLab CI/CD Setup Guide

## 📋 Необходимые CI/CD переменные

Перейди в **Settings → CI/CD → Variables** и добавь следующие переменные:

### 🔐 GCP Credentials

| Variable | Type | Protected | Masked | Description |
|----------|------|-----------|--------|-------------|
| `GCP_SERVICE_ACCOUNT_KEY` | File | ✅ | ✅ | Base64-encoded GCP service account key для dev |
| `GCP_SERVICE_ACCOUNT_KEY_PROD` | File | ✅ | ✅ | Base64-encoded GCP service account key для prod |

### 📧 SendGrid (опционально для тестов)

| Variable | Type | Protected | Masked | Description |
|----------|------|-----------|--------|-------------|
| `SENDGRID_API_KEY` | Variable | ❌ | ✅ | SendGrid API key для тестов |

## 🔧 Как получить GCP Service Account Key

```bash
# 1. Создай service account (если еще нет)
gcloud iam service-accounts create gitlab-ci \
  --display-name="GitLab CI/CD" \
  --project=zeno-cy-dev-001

# 2. Выдай необходимые права
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.admin"

# 3. Создай и скачай ключ
gcloud iam service-accounts keys create gitlab-ci-key.json \
  --iam-account=gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com

# 4. Закодируй в base64
cat gitlab-ci-key.json | base64 > gitlab-ci-key-base64.txt

# 5. Скопируй содержимое gitlab-ci-key-base64.txt в GitLab CI/CD Variables
```

## 🏷️ Настройка Labels

Создай следующие labels в **Settings → Labels**:

- `~bug` (красный) - Баги и ошибки
- `~feature` (зеленый) - Новая функциональность
- `~enhancement` (синий) - Улучшения
- `~documentation` (желтый) - Документация
- `~security` (оранжевый) - Безопасность
- `~performance` (фиолетовый) - Производительность
- `~refactoring` (серый) - Рефакторинг

## 🔒 Protected Branches

Настрой в **Settings → Repository → Protected branches**:

- `main` - Allowed to merge: Maintainers, Allowed to push: No one
- `develop` - Allowed to merge: Developers, Allowed to push: Developers

## 🎯 Merge Request Settings

В **Settings → Merge requests**:

- ✅ Enable "Delete source branch" option by default
- ✅ Enable "Squash commits when merging"
- ✅ Pipelines must succeed
- ✅ All threads must be resolved

## 🚀 Deployment Environments

Environments будут созданы автоматически при первом деплое:

- `development` - https://zeno-auth-dev-umu7aajgeq-ey.a.run.app
- `production` - https://zeno-auth-prod.zeno-cy.com

## 📊 Code Coverage

Coverage badge будет доступен после первого успешного pipeline:

```markdown
[![coverage](https://gitlab.com/zeno-cy/zeno-auth/badges/main/coverage.svg)](https://gitlab.com/zeno-cy/zeno-auth/-/commits/main)
```

## 🔔 Notifications

Настрой в **Settings → Integrations**:

- Slack/Discord для уведомлений о деплоях
- Email для failed pipelines

## ✅ Готово!

После настройки всех переменных, сделай push в `main` или `develop` ветку, и pipeline запустится автоматически.
