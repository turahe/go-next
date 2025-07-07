# switch-env.ps1
param(
    [ValidateSet("development", "staging", "production")]
    [string]$envType = "development"
)

switch ($envType) {
    "development" {
        $env:ENV = "development"
        $env:BUILD_TARGET = "dev"
    }
    "staging" {
        $env:ENV = "staging"
        $env:BUILD_TARGET = "staging"
    }
    "production" {
        $env:ENV = "production"
        $env:BUILD_TARGET = "prod"
    }
}
Write-Host "Switched to $envType environment."
Write-Host "ENV=$env:ENV, BUILD_TARGET=$env:BUILD_TARGET"