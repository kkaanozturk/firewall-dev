# Path: start-lab.ps1
# AegisNova Global - Tek Komutla Canlı Web NGFW & SOC Laboratuvarı Başlatıcı (Windows PowerShell)
# Kullanım: .\start-lab.ps1

$ErrorActionPreference = "Stop"
$LabUrl = "http://127.0.0.1:8080"

Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host " AegisNova Global - Canli NGFW, QoS ve 7/24 SOC Laboratuvari Baslatiliyor..." -ForegroundColor Cyan
Write-Host " Tarayici Adresi: $LabUrl" -ForegroundColor Green
Write-Host " Durdurmak icin: Ctrl + C" -ForegroundColor Yellow
Write-Host "==============================================================================" -ForegroundColor Cyan

Start-Job -ScriptBlock {
    Start-Sleep -Seconds 2
    Start-Process "http://127.0.0.1:8080"
} | Out-Null

go run ./cmd/aegisnova-lab -addr 127.0.0.1:8080

