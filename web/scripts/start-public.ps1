$ErrorActionPreference = 'Stop'

function Get-NgrokExe {
  try {
    $cmd = Get-Command ngrok -ErrorAction Stop
    return $cmd.Source
  } catch {
    $fallback = Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Packages\Ngrok.Ngrok_Microsoft.Winget.Source_8wekyb3d8bbwe\ngrok.exe'
    if (Test-Path $fallback) {
      return $fallback
    }
    throw 'ngrok executable not found. Install ngrok first.'
  }
}

function Get-NgrokPublicUrl {
  param(
    [int]$TimeoutSeconds = 20
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    try {
      $resp = Invoke-RestMethod -Uri 'http://127.0.0.1:4040/api/tunnels' -Method Get -TimeoutSec 2
      if ($resp.tunnels) {
        $httpsTunnel = $resp.tunnels | Where-Object { $_.public_url -like 'https://*' } | Select-Object -First 1
        if ($httpsTunnel) {
          return $httpsTunnel.public_url
        }
      }
    } catch {
      Start-Sleep -Milliseconds 800
      continue
    }
    Start-Sleep -Milliseconds 800
  }

  return $null
}

function Test-PortListening {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port
  )

  $line = netstat -ano | Select-String -Pattern ":$Port\s+.*LISTENING"
  return [bool]$line
}

function Wait-PortListening {
  param(
    [Parameter(Mandatory = $true)]
    [int]$Port,
    [int]$TimeoutSeconds = 20
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    if (Test-PortListening -Port $Port) {
      return $true
    }
    Start-Sleep -Milliseconds 500
  }

  return $false
}

# 1) Start Vite on a fixed LAN/public-friendly port.
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$repoRoot = (Resolve-Path (Join-Path $projectRoot '..')).Path

# 0) Ensure backend is running, otherwise login and API calls will return proxy 500.
if (-not (Test-PortListening -Port 8080)) {
  $backendCmd = "Set-Location '$repoRoot\\server'; go run main.go"
  Start-Process powershell -ArgumentList @('-NoExit', '-Command', $backendCmd)
  if (Wait-PortListening -Port 8080 -TimeoutSeconds 30) {
    Write-Host "Backend started on 8080."
  } else {
    Write-Host "Backend did not become ready on 8080 within 30s. Check server terminal logs."
  }
} else {
  Write-Host "Backend already listening on 8080, reusing existing process."
}

if (-not (Test-PortListening -Port 5173)) {
  $viteCmd = "Set-Location '$projectRoot'; npm run dev:public"
  Start-Process powershell -ArgumentList @('-NoExit', '-Command', $viteCmd)
} else {
  Write-Host "Vite already listening on 5173, reusing existing process."
}

Start-Sleep -Seconds 3

# 2) Start ngrok tunnel using local config if available.
$configPath = Join-Path $projectRoot 'ngrok.yml'
$ngrokExe = Get-NgrokExe
if (-not (Get-NgrokPublicUrl -TimeoutSeconds 2)) {
  if (Test-Path $configPath) {
    $ngrokCmd = "Set-Location '$projectRoot'; & '$ngrokExe' start --config '$configPath' interview-web"
    Start-Process powershell -ArgumentList @('-NoExit', '-Command', $ngrokCmd)
  } else {
    $ngrokCmd = "Set-Location '$projectRoot'; & '$ngrokExe' http https://localhost:5173"
    Start-Process powershell -ArgumentList @('-NoExit', '-Command', $ngrokCmd)
  }
} else {
  Write-Host "ngrok tunnel already online, reusing existing endpoint."
}

Write-Host "Public startup command sent."
Write-Host "If this is first time, copy ngrok.yml.example to ngrok.yml and fill authtoken first."

$publicUrl = Get-NgrokPublicUrl -TimeoutSeconds 25
if ($publicUrl) {
  Write-Host "Share this URL with interviewer: $publicUrl"
} else {
  Write-Host "Could not fetch ngrok URL automatically yet."
  Write-Host "Open http://127.0.0.1:4040/api/tunnels to view public_url."
}
