# 🚀 Deployment Setup Required

## ⚠️ Current Status

Deploy workflows требуют настройки Workload Identity Federation в GCP перед использованием.

## 📋 Что нужно сделать

### 1. Создать Workload Identity Pool

```bash
gcloud iam workload-identity-pools create "github" \
  --project="zeno-cy-dev-001" \
  --location="global" \
  --display-name="GitHub Actions Pool"
```

### 2. Создать Workload Identity Provider

```bash
gcloud iam workload-identity-pools providers create-oidc "github-provider" \
  --project="zeno-cy-dev-001" \
  --location="global" \
  --workload-identity-pool="github" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"
```

### 3. Создать Service Account

```bash
gcloud iam service-accounts create github-actions \
  --project="zeno-cy-dev-001" \
  --display-name="GitHub Actions"
```

### 4. Настроить IAM Bindings

```bash
# Allow GitHub Actions to impersonate service account
gcloud iam service-accounts add-iam-policy-binding "github-actions@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --project="zeno-cy-dev-001" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github/attribute.repository/ZenoN-Cloud/zeno-auth"

# Grant necessary permissions
gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:github-actions@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:github-actions@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding zeno-cy-dev-001 \
  --member="serviceAccount:github-actions@zeno-cy-dev-001.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"
```

### 5. Добавить GitHub Secrets

В настройках репозитория добавить:

```
WIF_PROVIDER=projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github/providers/github-provider
WIF_SERVICE_ACCOUNT=github-actions@zeno-cy-dev-001.iam.gserviceaccount.com
```

Для production:
```
WIF_PROVIDER_PROD=projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github/providers/github-provider
WIF_SERVICE_ACCOUNT_PROD=github-actions@zeno-cy-prod-001.iam.gserviceaccount.com
```

## 🔧 Временное решение

Пока WIF не настроен, deploy workflows будут падать на этапе аутентификации. Это нормально.

Тесты (test.yml) работают независимо и должны проходить успешно.

## ✅ После настройки

1. Проверить workflows вручную: `gh workflow run deploy-dev.yml`
2. Убедиться что деплой проходит успешно
3. Проверить health endpoints
4. Настроить production окружение аналогично

---

**Статус:** ⏳ Требуется настройка GCP  
**Приоритет:** Средний (тесты работают, локальная разработка не затронута)
