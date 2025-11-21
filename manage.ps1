<#
.SYNOPSIS
    Project Task Runner (Makefile replacement for Windows)
#>
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("build", "run", "migrate-up", "migrate-down", "generate", "clean", "test")]
    [string]$Task
)

$DB_URL = "sqlite3://dev.db"
$MIGRATE_CMD = "migrate -path sql/schema -database ""$DB_URL"""

switch ($Task) {
    "build" {
        Write-Host "Building binary..." -ForegroundColor Cyan
        go build -o bin/api.exe cmd/api/main.go
    }
    "run" {
        Write-Host "Starting Air..." -ForegroundColor Cyan
        air
    }
    "migrate-up" {
        if (!(Test-Path dev.db)) {
            Write-Host "Creating dev.db..." -ForegroundColor Yellow
            New-Item -Path dev.db -ItemType File | Out-Null
        }
        Write-Host "Running migrations UP..." -ForegroundColor Cyan
        Invoke-Expression "$MIGRATE_CMD up"
    }
    "migrate-down" {
        Write-Host "Running migrations DOWN..." -ForegroundColor Cyan
        Invoke-Expression "$MIGRATE_CMD down"
    }
    "generate" {
        Write-Host "Generating SQLC code..." -ForegroundColor Cyan
        sqlc generate
        Write-Host "Generating Swagger docs..." -ForegroundColor Cyan
        swag init -g cmd/api/main.go
    }
    "test" {
        Write-Host "Running tests..." -ForegroundColor Cyan
        go test -v ./...
    }
    "clean" {
        Write-Host "Cleaning artifacts..." -ForegroundColor Cyan
        if (Test-Path bin/api.exe) { Remove-Item bin/api.exe }
        if (Test-Path dev.db) { Remove-Item dev.db }
        Write-Host "Done."
    }
}
