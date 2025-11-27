# 🔐 GitLab CI/CD Variables

## Быстрая настройка

Перейди в: **Settings → CI/CD → Variables** и добавь:

## Development Environment

```bash
# Variable: GCP_SERVICE_ACCOUNT_KEY
# Type: File
# Protected: Yes
# Masked: Yes
# Environment: development

# Получить значение:
gcloud iam service-accounts keys create gitlab-ci-dev-key.json \
  --iam-account=gitlab-ci@zeno-cy-dev-001.iam.gserviceaccount.com

cat gitlab-ci-dev-key.json | base64 | pbcopy
# Вставь в GitLab
```

## Production Environment

```bash
# Variable: GCP_SERVICE_ACCOUNT_KEY_PROD
# Type: File
# Protected: Yes
# Masked: Yes
# Environment: production

# Получить значение:
gcloud iam service-accounts keys create gitlab-ci-prod-key.json \
  --iam-account=gitlab-ci@zeno-cy-prod-001.iam.gserviceaccount.com

cat gitlab-ci-prod-key.json | base64 | pbcopy
# Вставь в GitLab
```

## Проверка

После добавления переменных:

1. Перейди в **CI/CD → Pipelines**
2. Нажми **Run pipeline**
3. Выбери ветку `main` или `develop`
4. Нажми **Run pipeline**

Pipeline должен пройти все стадии:
- ✅ Lint
- ✅ Test
- ✅ Security
- ✅ Build (только для main/develop)
- ✅ Deploy (только для main, manual для prod)

## Troubleshooting

### Pipeline fails на стадии build

**Проблема:** `gcloud auth activate-service-account` fails

**Решение:**
1. Проверь, что `GCP_SERVICE_ACCOUNT_KEY` в base64
2. Проверь права service account:
   - `roles/run.admin`
   - `roles/storage.admin`
   - `roles/artifactregistry.admin`

### Pipeline fails на стадии deploy

**Проблема:** `gcloud run deploy` fails

**Решение:**
1. Проверь, что Cloud Run API включен
2. Проверь, что Cloud SQL instance существует
3. Проверь, что secrets существуют в Secret Manager

### Coverage badge не работает

**Решение:**
1. Дождись успешного завершения pipeline
2. Обнови страницу README
3. Badge появится автоматически

## Полезные ссылки

- [GitLab CI/CD Variables](https://docs.gitlab.com/ee/ci/variables/)
- [GCP Service Accounts](https://cloud.google.com/iam/docs/service-accounts)
- [GitLab Environments](https://docs.gitlab.com/ee/ci/environments/)
