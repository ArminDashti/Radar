#Requires -Version 5.1
<#
.SYNOPSIS
  radar CLI: lifecycle for the radar-api Docker stack (SQLite + API).

.EXAMPLE
  radar run --port=8090
  radar rebuild --remove-data=yes -y
#>
param(
    [Parameter(Position = 0)][string]$Command = 'help',
    [Parameter(Position = 1, ValueFromRemainingArguments = $true)][string[]]$Rest = @()
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RadarVersion = '0.2.0'
$Project = 'radar'
$Root = $PSScriptRoot
$ComposeFile = Join-Path $Root 'docker-compose.yml'
$Dockerfile = Join-Path $Root 'dockerfile'
$ImageTag = if ($env:IMAGE_TAG) { $env:IMAGE_TAG } else { 'radar-api:latest' }
$Network = if ($env:DOCKER_NETWORK) { $env:DOCKER_NETWORK } else { 't3-net' }
$env:IMAGE_TAG = $ImageTag
$env:DOCKER_NETWORK = $Network

$DataRoot = Join-Path $env:LOCALAPPDATA 'radar'
$ConfigFile = Join-Path $DataRoot 'config.json'
$LogFile = Join-Path $DataRoot 'logs\radar.log'
$DefaultPort = 8088
$ReadyTimeoutSec = 30
$script:ExitCode = 0

function Write-Step([string]$Message) {
    Write-Host ">> $Message" -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path (Split-Path $LogFile) | Out-Null
    if ((Test-Path $LogFile) -and (Get-Item $LogFile).Length -gt 10MB) { Move-Item $LogFile "$LogFile.1" -Force }
    Add-Content -Path $LogFile -Value "$(Get-Date -Format s) $Message"
}

function Show-Help {
    Write-Host @"
radar $RadarVersion - manage the radar-api Docker stack

USAGE:
  radar status                       Installed, running, port, URL check (0 running / 3 stopped / 4 not installed)
  radar version                      CLI version, installed vs source version
  radar run [--port=N]               Start the stack (optionally on port N) and wait for the URL
  radar stop                         Stop the stack (keeps containers and data)
  radar port [--port=N]              Show saved port, or save N (restarts stack if running)
  radar update                       git pull --ff-only, then rebuild keeping data
  radar rebuild [--remove-data=no|yes] [-y]
                                     Rebuild image and recreate stack; 'yes' wipes data and config
  radar uninstall [-y]               Remove containers, volumes (all data), image, config and PATH entry
  radar help                         Show this help
"@ -ForegroundColor Cyan
}

function Get-Flag([string]$Name) {
    for ($i = 0; $i -lt $Rest.Count; $i++) {
        $token = $Rest[$i]
        if ($token -like "--$Name=*") { return $token.Substring($Name.Length + 3) }
        if ($token -eq "--$Name" -and ($i + 1) -lt $Rest.Count) { return $Rest[$i + 1] }
    }
    return $null
}

function Test-Yes { return ($Rest -contains '-y') }

function Assert-Docker {
    cmd /c "docker info >nul 2>&1"
    if ($LASTEXITCODE -ne 0) { throw 'Docker is not available. Start Docker Desktop / daemon.' }
}

function Get-Config {
    if (Test-Path $ConfigFile) { return Get-Content $ConfigFile -Raw | ConvertFrom-Json }
    return $null
}

function Assert-Installed {
    if (-not (Get-Config)) { throw 'Not installed. Run: radar rebuild' }
}

function Get-SavedPort {
    $cfg = Get-Config
    if ($cfg) { return [int]$cfg.port }
    return $DefaultPort
}

function Get-SourceVersion {
    $head = cmd /c "git -C `"$Root`" rev-parse --short HEAD 2>nul"
    if ($head) { return [string]$head } else { return 'unknown' }
}

function Save-Config([int]$Port, [string]$Version) {
    New-Item -ItemType Directory -Force -Path $DataRoot | Out-Null
    [ordered]@{ port = $Port; version = $Version; sourceDir = $Root; mode = 'docker'; project = $Project } |
        ConvertTo-Json | Set-Content -Path $ConfigFile
}

function Set-PortEnv { $env:RADAR_PORT = [string](Get-SavedPort) }

function Invoke-Compose {
    Set-PortEnv
    & docker compose -p $Project -f $ComposeFile --project-directory $Root @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose $($args -join ' ') failed" }
}

function Get-RunningCount {
    Set-PortEnv
    $ids = cmd /c "docker compose -p $Project -f `"$ComposeFile`" --project-directory `"$Root`" ps -q --status running 2>nul"
    return @($ids | Where-Object { $_ }).Count
}

function Test-Running { return (Get-RunningCount) -gt 0 }

function Test-Image {
    cmd /c "docker image inspect $ImageTag >nul 2>&1"
    return ($LASTEXITCODE -eq 0)
}

function Ensure-Network {
    cmd /c "docker network inspect $Network >nul 2>&1"
    if ($LASTEXITCODE -ne 0) {
        Write-Step "Creating network $Network"
        cmd /c "docker network create $Network >nul 2>&1"
        if ($LASTEXITCODE -ne 0) { throw "Failed to create network $Network" }
    }
}

function Build-Image([switch]$Fresh) {
    Write-Step "Building image $ImageTag"
    $buildArgs = @('build', '-f', $Dockerfile, '-t', $ImageTag)
    if ($Fresh) { $buildArgs += @('--no-cache', '--pull') }
    & docker @buildArgs $Root
    if ($LASTEXITCODE -ne 0) { throw 'docker build failed; existing install untouched' }
}

function Start-Stack {
    Ensure-Network
    if (-not (Test-Image)) { Build-Image }
    Write-Step 'Starting stack'
    Invoke-Compose up -d
}

function Test-Url([int]$Port) {
    try {
        Invoke-WebRequest -Uri "http://localhost:$Port/api/probes" -UseBasicParsing -TimeoutSec 3 | Out-Null
        return $true
    }
    catch {
        return ($null -ne $_.Exception.Response)
    }
}

function Wait-Ready([int]$Port) {
    $deadline = (Get-Date).AddSeconds($ReadyTimeoutSec)
    while ((Get-Date) -lt $deadline) {
        if (Test-Url $Port) { Write-Step "Ready: http://localhost:$Port"; return }
        Start-Sleep -Seconds 1
    }
    throw "Not reachable at http://localhost:$Port within ${ReadyTimeoutSec}s. See: docker logs radar-api"
}

function Test-PortFree([int]$Port) {
    return -not (Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue)
}

function Confirm-Destructive([string]$Message) {
    if (Test-Yes) { return }
    $answer = Read-Host "$Message Type 'yes' to continue"
    if ($answer -ne 'yes') { throw 'Aborted.' }
}

function Set-UserPath([bool]$Add) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $entries = @($current -split ';' | Where-Object { $_ })
    $present = @($entries | Where-Object { $_.TrimEnd('\') -ieq $Root.TrimEnd('\') }).Count -gt 0
    if ($Add -eq $present) { return }
    if ($Add) { $entries += $Root } else { $entries = @($entries | Where-Object { $_.TrimEnd('\') -ine $Root.TrimEnd('\') }) }
    [Environment]::SetEnvironmentVariable('Path', ($entries -join ';'), 'User')
    Write-Step ($(if ($Add) { "Added $Root to user PATH" } else { "Removed $Root from user PATH" }))
}

function Show-Status {
    $cfg = Get-Config
    if (-not $cfg) { Write-Host 'installed: no (run: radar rebuild)' -ForegroundColor Yellow; return 4 }
    $running = Test-Running
    $port = Get-SavedPort
    Write-Host "installed: yes (mode: docker, version: $($cfg.version))"
    Write-Host "running:   $(if ($running) { 'yes' } else { 'no' })"
    Write-Host "port:      $port"
    Write-Host "url:       http://localhost:$port"
    Write-Host "compose:   $ComposeFile"
    Write-Host "config:    $ConfigFile"
    if ($running) {
        Write-Host "url check: $(if (Test-Url $port) { 'ok' } else { 'unreachable' })"
        return 0
    }
    return 3
}

function Show-Version {
    $cfg = Get-Config
    $installed = if ($cfg) { [string]$cfg.version } else { 'not installed' }
    $source = Get-SourceVersion
    Write-Host "radar CLI:  $RadarVersion"
    Write-Host "installed:  $installed"
    Write-Host "source:     $source"
    if ($cfg -and $installed -ne $source) { Write-Host 'mismatch: run radar update or radar rebuild' -ForegroundColor Yellow }
}

function Set-Port([string]$Requested) {
    $number = 0
    if (-not [int]::TryParse($Requested, [ref]$number) -or $number -lt 1 -or $number -gt 65535) {
        throw "Invalid port '$Requested' (expected 1-65535)"
    }
    if (-not (Test-Running) -and -not (Test-PortFree $number)) { throw "Port $number is already in use" }
    Save-Config -Port $number -Version (Get-Config).version
    Write-Step "Saved port $number"
}

function Start-Radar {
    Assert-Docker
    Assert-Installed
    $requested = Get-Flag 'port'
    $wasRunning = Test-Running
    if ($requested) { Set-Port $requested }
    elseif ($wasRunning) { $script:ExitCode = Show-Status; return }
    Start-Stack
    Wait-Ready (Get-SavedPort)
}

function Stop-Radar {
    Assert-Docker
    Invoke-Compose stop
    Write-Step 'Stopped'
}

function Set-PortCommand {
    $requested = Get-Flag 'port'
    if (-not $requested) { Write-Output (Get-SavedPort); return }
    Assert-Docker
    Assert-Installed
    $wasRunning = Test-Running
    Set-Port $requested
    if ($wasRunning) {
        Start-Stack
        Wait-Ready (Get-SavedPort)
    }
}

function Invoke-Rebuild([string]$RemoveData) {
    Assert-Docker
    $wasRunning = Test-Running
    if ($RemoveData -eq 'yes') {
        Confirm-Destructive "This deletes the database and logo volumes, the API image, and $DataRoot."
        Write-Step 'Removing stack, volumes, image and config'
        Invoke-Compose down -v --rmi local --remove-orphans
        Remove-Item -Recurse -Force $DataRoot -ErrorAction SilentlyContinue
        Build-Image -Fresh
    }
    else {
        Build-Image -Fresh
        Write-Step 'Recreating containers (data kept)'
        Invoke-Compose down --remove-orphans
    }
    Ensure-Network
    Save-Config -Port (Get-SavedPort) -Version (Get-SourceVersion)
    Set-UserPath -Add $true
    if ($wasRunning) {
        Start-Stack
        Wait-Ready (Get-SavedPort)
    }
    Write-Step 'Rebuild complete'
}

function Update-Radar {
    Assert-Docker
    Assert-Installed
    Write-Step 'Pulling source (git pull --ff-only)'
    $pull = cmd /c "git -C `"$Root`" pull --ff-only 2>&1"
    if ($LASTEXITCODE -ne 0) { throw "git pull --ff-only failed; update aborted:`n$pull" }
    Write-Host $pull
    Invoke-Rebuild 'no'
}

function Uninstall-Radar {
    Assert-Docker
    Confirm-Destructive "This removes containers, volumes (ALL data), the API image, $DataRoot and the PATH entry. Network $Network is kept (shared)."
    Invoke-Compose down -v --rmi local --remove-orphans
    Remove-Item -Recurse -Force $DataRoot -ErrorAction SilentlyContinue
    Set-UserPath -Add $false
    Write-Host "Uninstalled. The CLI files in $Root remain; delete them manually if you want them gone." -ForegroundColor Green
}

function Invoke-Radar {
    switch ($Command) {
        'help' { Show-Help }
        'version' { Show-Version }
        'status' { $script:ExitCode = Show-Status }
        'run' { Start-Radar }
        'stop' { Stop-Radar }
        'port' { Set-PortCommand }
        'update' { Update-Radar }
        'rebuild' {
            $mode = Get-Flag 'remove-data'
            if (-not $mode) { $mode = 'no' }
            if ($mode -notin 'no', 'yes') { throw "--remove-data must be 'no' or 'yes'" }
            Invoke-Rebuild $mode
        }
        'uninstall' { Uninstall-Radar }
        default {
            Show-Help
            $script:ExitCode = 2
        }
    }
}

try {
    Invoke-Radar
}
catch {
    Write-Host "ERR $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
exit $script:ExitCode
