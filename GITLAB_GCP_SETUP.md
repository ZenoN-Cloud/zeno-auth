# 🔐 Настройка GCP Credentials для GitLab CI/CD

## Быстрая настройка (5 минут)

### 1. Создай Service Account для DEV

```bash
gcloud iam service-accounts create gitlab-ci \
  --display-name="GitLab CI/CD" \
  --project=zeno-cy-dev-001
```

### 2. Выдай права

```bash
PROJECT_ID="zeno-cy-dev-001"
SA_EMAIL="gitlab-ci@${PROJECT_ID}.iam.gserviceaccount.com"

# Cloud Run
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

# Artifact Registry
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.admin"

# Storage
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.admin"

# Cloud SQL
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.client"
```

### 3. Создай ключ

```bash
gcloud iam service-accounts keys create gitlab-ci-dev-key.json \
  --iam-account=gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com
```

### 4. Закодируй в base64

```bash
cat gitlab-ci-dev-key.json | base64 > gitlab-ci-dev-key.base64

# Скопируй в буфер обмена
cat gitlab-ci-dev-key.base64 | pbcopy
```

### 5. Добавь в GitLab

1. Перейди: https://gitlab.com/zeno-cy/zeno-auth/-/settings/ci_cd
2. Expand **Variables**
3. Нажми **Add variable**
4. Заполни:
   - **Key**: `GCP_SERVICE_ACCOUNT_KEY`
   - **Value**: вставь из буфера обмена
   - **Type**: File
   - **Protected**: ✅ Yes
   - **Masked**: ✅ Yes
   - **Environment scope**: All
5. Нажми **Add variable**

### 6. Удали локальные файлы

```bash
rm gitlab-ci-dev-key.json gitlab-ci-dev-key.base64
```

---

## Production (опционально)

Повтори те же шаги для production:

```bash
PROJECT_ID="zeno-cy-prod-001"

# 1. Создай SA
gcloud iam service-accounts create gitlab-ci \
  --display-name="GitLab CI/CD" \
  --project=${PROJECT_ID}

# 2. Выдай права
SA_EMAIL="gitlab-ci@${PROJECT_ID}.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.admin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.client"

# 3. Создай ключ
gcloud iam service-accounts keys create gitlab-ci-prod-key.json \
  --iam-account=gitlab-ci@${PROJECT_ID}.iam.gserviceaccount.com

# 4. Закодируй
cat gitlab-ci-prod-key.json | base64 > gitlab-ci-prod-key.base64
cat gitlab-ci-prod-key.base64 | pbcopy

# 5. Добавь в GitLab как GCP_SERVICE_ACCOUNT_KEY_PROD

# 6. Удали файлы
rm gitlab-ci-prod-key.json gitlab-ci-prod-key.base64
```

---

## Проверка

После добавления переменных:

1. Перейди: https://gitlab.com/zeno-cy/zeno-auth/-/pipelines
2. Нажми **Run pipeline**
3. Выбери `main`
4. Pipeline должен пройти все стадии включая Build и Deploy

---

## Troubleshooting

### Service account уже существует

```bash
# Просто создай новый ключ
gcloud iam service-accounts keys create gitlab-ci-dev-key.json \
  --iam-account=gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com
```

### Ошибка прав доступа

```bash
# Проверь права
gcloud projects get-iam-policy zeno-cy-dev-001 \
  --flatten="bindings[].members" \
  --filter="bindings.members:gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com"
```

### Pipeline падает на build

Проверь:
1. Переменная `GCP_SERVICE_ACCOUNT_KEY` добавлена
2. Type = File (не Variable!)
3. Base64 закодирован правильно

---

## ⚠️ Безопасность

- ❌ Никогда не коммить .json файлы
- ❌ Никогда не коммить .base64 файлы
- ✅ Удаляй локальные файлы после добавления в GitLab
- ✅ Используй Protected variables для production
- ✅ Регулярно ротируй ключи (каждые 90 дней)

---

**Готово! Теперь GitLab CI/CD может деплоить в GCP! 🚀**
