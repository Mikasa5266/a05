param(
  [Parameter(Mandatory = $true)]
  [string]$AuthToken
)

$ErrorActionPreference = 'Stop'

$ngrokExe = Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Packages\Ngrok.Ngrok_Microsoft.Winget.Source_8wekyb3d8bbwe\ngrok.exe'
if (-not (Test-Path $ngrokExe)) {
  try {
    $cmd = Get-Command ngrok -ErrorAction Stop
    $ngrokExe = $cmd.Source
  } catch {
    throw 'ngrok executable not found. Please install with: winget install --id Ngrok.Ngrok --exact'
  }
}

& $ngrokExe config add-authtoken $AuthToken
Write-Host 'ngrok authtoken configured successfully.'
Write-Host 'Next: copy ngrok.yml.example to ngrok.yml, then run ./scripts/start-public.ps1'
