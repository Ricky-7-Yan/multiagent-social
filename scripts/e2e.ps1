Param()
Write-Host "Starting docker compose..."
Push-Location (Split-Path -Path $PSScriptRoot -Parent)
docker compose -f .\deployments\docker\docker-compose.yml up -d --build
Start-Sleep -Seconds 5
$env:PG_DSN = "postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
Write-Host "Running migrations..."
psql $env:PG_DSN -f .\migrations\001_init.sql
Write-Host "Creating conversation..."
$conv = curl -s -X POST http://localhost:8080/api/v1/conversations
Write-Host "conversation id: $conv"
Start-Sleep -Seconds 1
curl -s -X POST --data "hello from e2e" "http://localhost:8080/api/v1/conversations/$conv/messages"
Write-Host "posted message"
Write-Host "Done"
Pop-Location

