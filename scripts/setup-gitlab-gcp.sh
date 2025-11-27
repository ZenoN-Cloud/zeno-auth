#!/bin/bash

set -e

echo "🔐 Setting up GCP credentials for GitLab CI/CD"
echo ""

PROJECT_DEV="zeno-cy-dev-001"
PROJECT_PROD="zeno-cy-prod-001"
SA_NAME="gitlab-ci"
REGION="europe-west3"

# Проверяем, что gcloud установлен
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI not found. Install it first:"
    echo "   https://cloud.google.com/sdk/docs/install"
    exit 1
fi

echo "📋 Current GCP project:"
gcloud config get-value project
echo ""

# Development
echo "🔧 Step 1: Creating service account for DEV"
echo ""

# Проверяем, существует ли service account
if gcloud iam service-accounts describe ${SA_NAME}@${PROJECT_DEV}.iam.gserviceaccount.com --project=${PROJECT_DEV} &>/dev/null; then
    echo "✅ Service account ${SA_NAME}@${PROJECT_DEV}.iam.gserviceaccount.com already exists"
else
    echo "Creating service account..."
    gcloud iam service-accounts create ${SA_NAME} \
        --display-name="GitLab CI/CD" \
        --project=${PROJECT_DEV}
    echo "✅ Service account created"
fi

echo ""
echo "🔑 Step 2: Granting permissions for DEV"
echo ""

# Выдаем права
for role in "roles/run.admin" "roles/storage.admin" "roles/artifactregistry.admin" "roles/cloudsql.client"; do
    echo "Granting ${role}..."
    gcloud projects add-iam-policy-binding ${PROJECT_DEV} \
        --member="serviceAccount:${SA_NAME}@${PROJECT_DEV}.iam.gserviceaccount.com" \
        --role="${role}" \
        --quiet
done

echo "✅ Permissions granted"
echo ""

echo "🔐 Step 3: Creating key for DEV"
echo ""

# Создаем ключ
KEY_FILE="gitlab-ci-dev-key.json"
if [ -f "${KEY_FILE}" ]; then
    echo "⚠️  ${KEY_FILE} already exists. Delete it? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        rm "${KEY_FILE}"
    else
        echo "Skipping key creation"
        KEY_FILE=""
    fi
fi

if [ -n "${KEY_FILE}" ]; then
    gcloud iam service-accounts keys create ${KEY_FILE} \
        --iam-account=${SA_NAME}@${PROJECT_DEV}.iam.gserviceaccount.com \
        --project=${PROJECT_DEV}
    
    echo "✅ Key created: ${KEY_FILE}"
    echo ""
    
    # Кодируем в base64
    echo "📦 Encoding to base64..."
    cat ${KEY_FILE} | base64 > ${KEY_FILE}.base64
    
    echo "✅ Base64 encoded: ${KEY_FILE}.base64"
    echo ""
    
    # Копируем в буфер обмена (если доступно)
    if command -v pbcopy &> /dev/null; then
        cat ${KEY_FILE}.base64 | pbcopy
        echo "✅ Copied to clipboard!"
    elif command -v xclip &> /dev/null; then
        cat ${KEY_FILE}.base64 | xclip -selection clipboard
        echo "✅ Copied to clipboard!"
    else
        echo "📋 Copy this value manually:"
        cat ${KEY_FILE}.base64
    fi
    
    echo ""
    echo "🎯 Add to GitLab:"
    echo "   Settings → CI/CD → Variables"
    echo "   Name: GCP_SERVICE_ACCOUNT_KEY"
    echo "   Type: File"
    echo "   Protected: Yes"
    echo "   Masked: Yes"
    echo "   Value: <paste from clipboard>"
    echo ""
fi

# Production (опционально)
echo ""
echo "🔧 Setup PRODUCTION credentials? (y/n)"
read -r response

if [[ "$response" =~ ^[Yy]$ ]]; then
    echo ""
    echo "🔧 Step 4: Creating service account for PROD"
    echo ""
    
    if gcloud iam service-accounts describe ${SA_NAME}@${PROJECT_PROD}.iam.gserviceaccount.com --project=${PROJECT_PROD} &>/dev/null; then
        echo "✅ Service account ${SA_NAME}@${PROJECT_PROD}.iam.gserviceaccount.com already exists"
    else
        echo "Creating service account..."
        gcloud iam service-accounts create ${SA_NAME} \
            --display-name="GitLab CI/CD" \
            --project=${PROJECT_PROD}
        echo "✅ Service account created"
    fi
    
    echo ""
    echo "🔑 Step 5: Granting permissions for PROD"
    echo ""
    
    for role in "roles/run.admin" "roles/storage.admin" "roles/artifactregistry.admin" "roles/cloudsql.client"; do
        echo "Granting ${role}..."
        gcloud projects add-iam-policy-binding ${PROJECT_PROD} \
            --member="serviceAccount:${SA_NAME}@${PROJECT_PROD}.iam.gserviceaccount.com" \
            --role="${role}" \
            --quiet
    done
    
    echo "✅ Permissions granted"
    echo ""
    
    echo "🔐 Step 6: Creating key for PROD"
    echo ""
    
    KEY_FILE_PROD="gitlab-ci-prod-key.json"
    gcloud iam service-accounts keys create ${KEY_FILE_PROD} \
        --iam-account=${SA_NAME}@${PROJECT_PROD}.iam.gserviceaccount.com \
        --project=${PROJECT_PROD}
    
    echo "✅ Key created: ${KEY_FILE_PROD}"
    echo ""
    
    cat ${KEY_FILE_PROD} | base64 > ${KEY_FILE_PROD}.base64
    echo "✅ Base64 encoded: ${KEY_FILE_PROD}.base64"
    echo ""
    
    echo "📋 Copy this value:"
    cat ${KEY_FILE_PROD}.base64
    echo ""
    echo "🎯 Add to GitLab:"
    echo "   Settings → CI/CD → Variables"
    echo "   Name: GCP_SERVICE_ACCOUNT_KEY_PROD"
    echo "   Type: File"
    echo "   Protected: Yes"
    echo "   Masked: Yes"
    echo ""
fi

echo ""
echo "✅ Done!"
echo ""
echo "📚 Next steps:"
echo "1. Go to https://gitlab.com/zeno-cy/zeno-auth/-/settings/ci_cd"
echo "2. Expand 'Variables'"
echo "3. Add the credentials as described above"
echo "4. Run a new pipeline"
echo ""
echo "⚠️  Security reminder:"
echo "   - Delete the .json files after adding to GitLab"
echo "   - Never commit these files to git"
echo ""
