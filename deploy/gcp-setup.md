# GCP Setup Guide

## 1. Create Projects

```bash
gcloud projects create zenon-cloud-dev-001 --name="ZenoN Cloud Dev"
gcloud projects create zenon-cloud-prod-001 --name="ZenoN Cloud Prod"
```

## 2. Enable APIs

```bash
# Dev
gcloud config set project zenon-cloud-dev-001
gcloud services enable secretmanager.googleapis.com sqladmin.googleapis.com artifactregistry.googleapis.com run.googleapis.com iam.googleapis.com compute.googleapis.com

# Prod  
gcloud config set project zenon-cloud-prod-001
gcloud services enable secretmanager.googleapis.com sqladmin.googleapis.com artifactregistry.googleapis.com run.googleapis.com iam.googleapis.com compute.googleapis.com
```

## 3. Create Service Accounts

```bash
# Dev
gcloud config set project zenon-cloud-dev-001
gcloud iam service-accounts create zeno-auth-dev --display-name="Zeno Auth Dev"
gcloud iam service-accounts create zeno-auth-cicd-dev --display-name="Zeno Auth CI/CD Dev"

# Prod
gcloud config set project zenon-cloud-prod-001  
gcloud iam service-accounts create zeno-auth-prod --display-name="Zeno Auth Prod"
gcloud iam service-accounts create zeno-auth-cicd-prod --display-name="Zeno Auth CI/CD Prod"
```

## 4. Create Cloud SQL

```bash
# Dev
gcloud config set project zenon-cloud-dev-001
gcloud sql instances create zenon-dev-sql --database-version=POSTGRES_17 --tier=db-f1-micro --region=europe-west3
gcloud sql databases create zeno_auth --instance=zenon-dev-sql
gcloud sql users create zeno_auth --instance=zenon-dev-sql --password=CHANGE_ME

# Prod
gcloud config set project zenon-cloud-prod-001
gcloud sql instances create zenon-prod-sql --database-version=POSTGRES_17 --tier=db-g1-small --region=europe-west3
gcloud sql databases create zeno_auth --instance=zenon-prod-sql
gcloud sql users create zeno_auth --instance=zenon-prod-sql --password=CHANGE_ME
```

## 5. Create Secrets

```bash
# Dev
gcloud config set project zenon-cloud-dev-001
echo "postgres://zeno_auth:PASSWORD@/zeno_auth?host=/cloudsql/zenon-cloud-dev-001:europe-west3:zenon-dev-sql" | gcloud secrets create zeno-auth-db-dsn --data-file=-
openssl genrsa -out jwt-private-dev.pem 2048
gcloud secrets create zeno-auth-jwt-private-key --data-file=jwt-private-dev.pem

# Prod
gcloud config set project zenon-cloud-prod-001
echo "postgres://zeno_auth:PASSWORD@/zeno_auth?host=/cloudsql/zenon-cloud-prod-001:europe-west3:zenon-prod-sql" | gcloud secrets create zeno-auth-db-dsn-prod --data-file=-
openssl genrsa -out jwt-private-prod.pem 2048
gcloud secrets create zeno-auth-jwt-private-key-prod --data-file=jwt-private-prod.pem
```

## 6. Create Artifact Registry

```bash
# Dev
gcloud config set project zenon-cloud-dev-001
gcloud artifacts repositories create zeno-auth-repo --repository-format=docker --location=europe-west3

# Prod
gcloud config set project zenon-cloud-prod-001
gcloud artifacts repositories create zeno-auth-repo --repository-format=docker --location=europe-west3
```

## 7. Setup IAM

```bash
# Dev service account
gcloud config set project zenon-cloud-dev-001
gcloud projects add-iam-policy-binding zenon-cloud-dev-001 --member="serviceAccount:zeno-auth-dev@zenon-cloud-dev-001.iam.gserviceaccount.com" --role="roles/cloudsql.client"
gcloud projects add-iam-policy-binding zenon-cloud-dev-001 --member="serviceAccount:zeno-auth-dev@zenon-cloud-dev-001.iam.gserviceaccount.com" --role="roles/secretmanager.secretAccessor"

# Dev CI/CD
gcloud projects add-iam-policy-binding zenon-cloud-dev-001 --member="serviceAccount:zeno-auth-cicd-dev@zenon-cloud-dev-001.iam.gserviceaccount.com" --role="roles/run.admin"
gcloud projects add-iam-policy-binding zenon-cloud-dev-001 --member="serviceAccount:zeno-auth-cicd-dev@zenon-cloud-dev-001.iam.gserviceaccount.com" --role="roles/iam.serviceAccountUser"
gcloud projects add-iam-policy-binding zenon-cloud-dev-001 --member="serviceAccount:zeno-auth-cicd-dev@zenon-cloud-dev-001.iam.gserviceaccount.com" --role="roles/artifactregistry.writer"
```