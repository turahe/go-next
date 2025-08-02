param(
    [string]$envType = "development"
)

$envFiles = @{
    "development" = @"
DB_HOST=localhost
DB_PORT=5432
DB_USER=devuser
DB_PASSWORD=devpass
DB_NAME=devdb
OSS_ACCESS_KEY=minioadmin
OSS_SECRET_KEY=minioadmin
OSS_BUCKET=dev-bucket
OSS_ENDPOINT=http://minio:9000
OSS_REGION=us-east-1
EMAIL_HOST=mailpit
EMAIL_PORT=1025
EMAIL_FROM=noreply@example.com
EMAIL_BASE_URL=http://localhost:8080
"@
    "staging" = @"
DB_HOST=staging-db
DB_PORT=5432
DB_USER=staginguser
DB_PASSWORD=stagingpass
DB_NAME=stagingdb
OSS_ACCESS_KEY=stagingaccess
OSS_SECRET_KEY=stagingsecret
OSS_BUCKET=staging-bucket
OSS_ENDPOINT=http://minio:9000
OSS_REGION=us-east-1
EMAIL_HOST=mailpit
EMAIL_PORT=1025
EMAIL_FROM=noreply@example.com
EMAIL_BASE_URL=https://staging.example.com
"@
    "production" = @"
DB_HOST=prod-db
DB_PORT=5432
DB_USER=produser
DB_PASSWORD=prodpass123!@#
DB_NAME=proddb
OSS_ACCESS_KEY=prodaccesskey
OSS_SECRET_KEY=prodsecretkey
OSS_BUCKET=prod-bucket
OSS_ENDPOINT=http://minio:9000
OSS_REGION=us-east-1
EMAIL_HOST=mailpit
EMAIL_PORT=1025
EMAIL_FROM=noreply@example.com
EMAIL_BASE_URL=https://yourdomain.com
"@
}

foreach ($key in $envFiles.Keys) {
    $fileName = ".env.$key"
    if (-not (Test-Path $fileName)) {
        $envFiles[$key] | Set-Content $fileName
        Write-Host "Created $fileName"
    } else {
        Write-Host "$fileName already exists, skipping."
    }
}

Write-Host "\nTo switch environments, set ENV and BUILD_TARGET before running Docker Compose."
Write-Host "Example for production:"
Write-Host "$env:ENV='production'; $env:BUILD_TARGET='prod'; docker compose up --build" 