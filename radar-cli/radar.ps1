#Requires -Version 5.1
<#
.SYNOPSIS
  radar CLI: lifecycle for the local Docker stack (radar-api, radar-agent, radar-webui).

.EXAMPLE
  radar run
  radar port --port=5190
  radar reinstall -y
#>
param(
    [Parameter(Position = 0)][string]$Command = 'help',
    [Parameter(Position = 1, ValueFromRemainingArguments = $true)][string[]]$Rest = @()
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$CliVersion = '1.0.0'
$Project = 'radar'
$Workspace = Split-Path -Parent $PSScriptRoot
$ComposeFile = Join-Path $Workspace 'radar-api\docker-compose.local.yml'
$Repos = @('radar-api', 'radar-agent', 'radar-webui') | ForEach-Object { Join-Path $Workspace $_ }
$LegacyPathDir = Join-Path $Workspace 'radar-api'
$Network = 'pc-armin-local'
$DataRoot = Join-Path $env:LOCALAPPDATA 'radar'
$ConfigFile = Join-Path $DataRoot 'config.json'
$LogFile = Join-Path $DataRoot 'logs\radar.log'
$DefaultPort = 5184
$ApiPort = 8088
# The webui container runs npm install on every start, so cold starts take minutes.
$ReadyTimeoutSec = 300
$script:ExitCode = 0

function Write-Step([string]$Message) {
    Write-Host ">> $Message" -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path (Split-Path $LogFile) | Out-Null
    if ((Test-Path $LogFile) -and (Get-Item $LogFile).Length -gt 10MB) { Move-Item $LogFile "$LogFile.1" -Force }
    Add-Content -Path $LogFile -Value "$(Get-Date -Format s) $Message"
}

function Show-Help {
    Write-Host @"
radar $CliVersion - manage the radar Docker stack (api, agent, webui)

USAGE:
  radar status                       Running state, port, URL checks (0 running / 3 stopped / 4 not installed)
  radar version                      CLI version and git version of each repo
  radar run [--port=N]               Start the stack (optionally save WebUI port N) and wait for the URL
  radar stop                         Stop the stack (keeps containers and data)
  radar port [--port=N]              Show saved WebUI port, or save N (applies if running)
  radar update                       git pull --ff-only in each repo, then rebuild keeping data
  radar rebuild                      Recreate containers with latest code; keeps data and caches (no re-downloads)
  radar reinstall [-y]               Delete containers, images, all data and config, then install from scratch
  radar uninstall [-y]               Remove containers, volumes (all data), config and PATH entry
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

function Confirm-Destructive([string]$Message) {
    if (Test-Yes) { return }
    $answer = Read-Host "$Message Type 'yes' to continue"
    if ($answer -ne 'yes') { throw 'Aborted.' }
}

function Assert-Docker {
    cmd /c "docker info >nul 2>&1"
    if ($LASTEXITCODE -ne 0) { throw 'Docker is not available. Start Docker Desktop / daemon.' }
}

function Assert-Network {
    cmd /c "docker network inspect $Network >nul 2>&1"
    if ($LASTEXITCODE -ne 0) { throw "Docker network '$Network' is missing. Create it: docker network create $Network" }
}

function Get-Config {
    # Configs written by the old api-only CLI lack composeFile; treat them as not installed.
    if (-not (Test-Path $ConfigFile)) { return $null }
    $cfg = Get-Content $ConfigFile -Raw | ConvertFrom-Json
    if ($cfg.PSObject.Properties['composeFile']) { return $cfg }
    return $null
}

function Save-Config([int]$Port) {
    New-Item -ItemType Directory -Force -Path $DataRoot | Out-Null
    [ordered]@{ port = $Port; mode = 'docker'; project = $Project; composeFile = $ComposeFile } |
        ConvertTo-Json | Set-Content -Path $ConfigFile
}

function Ensure-Config {
    if (-not (Get-Config)) { Save-Config $DefaultPort }
}

function Get-SavedPort {
    $cfg = Get-Config
    if ($cfg) { return [int]$cfg.port }
    return $DefaultPort
}

function Get-ValidPort([string]$Requested) {
    $number = 0
    if (-not [int]::TryParse($Requested, [ref]$number) -or $number -lt 1 -or $number -gt 65535) {
        throw "Invalid port '$Requested' (expected 1-65535)"
    }
    return $number
}

function Test-PortFree([int]$Port) {
    return -not (Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue)
}

function Set-PortEnv { $env:RADAR_WEBUI_PORT = [string](Get-SavedPort) }

function Invoke-Compose {
    Set-PortEnv
    & docker compose -p $Project -f $ComposeFile @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose $($args -join ' ') failed" }
}

function Get-RunningCount {
    Set-PortEnv
    $ids = cmd /c "docker compose -p $Project -f `"$ComposeFile`" ps -q --status running 2>nul"
    return @($ids | Where-Object { $_ }).Count
}

function Test-Running { return (Get-RunningCount) -gt 0 }

function Test-Url([string]$Url) {
    try {
        Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 3 | Out-Null
        return $true
    }
    catch {
        # A refused/timed-out connection has no Response property; strict mode would throw on direct access.
        $response = $_.Exception.PSObject.Properties['Response']
        return ($null -ne $response -and $null -ne $response.Value)
    }
}

function Wait-Ready([int]$Port) {
    $url = "http://localhost:$Port/radar/"
    $deadline = (Get-Date).AddSeconds($ReadyTimeoutSec)
    while ((Get-Date) -lt $deadline) {
        if (Test-Url $url) { Write-Step "Ready: $url"; return }
        Start-Sleep -Seconds 1
    }
    throw "Not reachable at $url within ${ReadyTimeoutSec}s. See: docker compose -p $Project logs"
}

function Start-Stack {
    Assert-Network
    Write-Step 'Starting stack'
    Invoke-Compose up -d
}

function Set-UserPathEntry([string]$Dir, [bool]$Present) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $entries = @($current -split ';' | Where-Object { $_ })
    $found = @($entries | Where-Object { $_.TrimEnd('\') -ieq $Dir.TrimEnd('\') }).Count -gt 0
    if ($Present -eq $found) { return }
    if ($Present) { $entries += $Dir } else { $entries = @($entries | Where-Object { $_.TrimEnd('\') -ine $Dir.TrimEnd('\') }) }
    [Environment]::SetEnvironmentVariable('Path', ($entries -join ';'), 'User')
    Write-Step ($(if ($Present) { "Added $Dir to user PATH" } else { "Removed $Dir from user PATH" }))
}

function Show-Status {
    $cfg = Get-Config
    if (-not $cfg) { Write-Host 'installed: no (run: radar run)' -ForegroundColor Yellow; return 4 }
    $port = Get-SavedPort
    $running = Test-Running
    Write-Host "installed: yes (mode: docker, cli: $CliVersion)"
    Write-Host "running:   $(if ($running) { 'yes' } else { 'no' })"
    Write-Host "webui:     http://localhost:$port/radar/"
    Write-Host "api:       http://localhost:$ApiPort/api/probes"
    Write-Host "compose:   $ComposeFile"
    cmd /c "docker network inspect $Network >nul 2>&1"
    if ($LASTEXITCODE -ne 0) { Write-Host "network:   $Network MISSING" -ForegroundColor Yellow }
    if ($running) {
        Write-Host "webui check: $(if (Test-Url "http://localhost:$port/radar/") { 'ok' } else { 'unreachable' })"
        Write-Host "api check:   $(if (Test-Url "http://localhost:$ApiPort/api/probes") { 'ok' } else { 'unreachable' })"
        return 0
    }
    return 3
}

function Show-Version {
    Write-Host "radar CLI: $CliVersion"
    foreach ($repo in $Repos) {
        $head = cmd /c "git -C `"$repo`" describe --always --dirty 2>nul"
        Write-Host ("{0,-12} {1}" -f (Split-Path -Leaf $repo), $(if ($head) { $head } else { 'unknown' }))
    }
}

function Set-Port([int]$Number) {
    if (-not (Test-Running) -and -not (Test-PortFree $Number)) { throw "Port $Number is already in use" }
    Save-Config $Number
    Write-Step "Saved WebUI port $Number"
}

function Start-Radar {
    Assert-Docker
    Assert-Network
    Ensure-Config
    $requested = Get-Flag 'port'
    if ($requested) { Set-Port (Get-ValidPort $requested) }
    elseif (Test-Running) { $script:ExitCode = Show-Status; return }
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
    Assert-Network
    Ensure-Config
    $wasRunning = Test-Running
    Set-Port (Get-ValidPort $requested)
    if ($wasRunning) {
        Start-Stack
        Wait-Ready (Get-SavedPort)
    }
}

function Invoke-Rebuild {
    # Recreate containers with the latest code; keep volumes (data, Go module and npm caches) so nothing cached is downloaded again.
    Assert-Docker
    Assert-Network
    Ensure-Config
    Write-Step 'Recreating containers (data and caches kept)'
    Invoke-Compose down --remove-orphans
    Invoke-Compose up -d --pull missing --force-recreate
    Wait-Ready (Get-SavedPort)
    Write-Step 'Rebuild complete'
}

function Invoke-Reinstall {
    # Remove containers, images, volumes and config completely, then install from scratch.
    Assert-Docker
    Assert-Network
    Confirm-Destructive "This deletes the stack containers, its images, ALL volumes (data and caches) and $DataRoot, then installs again."
    Write-Step 'Removing containers, images, volumes and config'
    Invoke-Compose down -v --rmi all --remove-orphans
    Remove-Item -Recurse -Force $DataRoot -ErrorAction SilentlyContinue
    Save-Config $DefaultPort
    Write-Step 'Installing fresh'
    Invoke-Compose up -d --pull always --force-recreate
    Wait-Ready (Get-SavedPort)
    Write-Step 'Reinstall complete'
}
function Update-Radar {
    Assert-Docker
    foreach ($repo in $Repos) {
        Write-Step "Pulling $(Split-Path -Leaf $repo) (git pull --ff-only)"
        $pull = cmd /c "git -C `"$repo`" pull --ff-only 2>&1"
        if ($LASTEXITCODE -ne 0) { throw "git pull --ff-only failed in $repo; update aborted:`n$pull" }
        Write-Host $pull
    }
    Invoke-Rebuild
}

function Uninstall-Radar {
    Assert-Docker
    Confirm-Destructive "This removes the stack containers, ALL volumes (data), $DataRoot and the PATH entries. Network $Network is kept (shared)."
    Invoke-Compose down -v --remove-orphans
    Remove-Item -Recurse -Force $DataRoot -ErrorAction SilentlyContinue
    Set-UserPathEntry $PSScriptRoot $false
    Set-UserPathEntry $LegacyPathDir $false
    Write-Host "Uninstalled. The CLI files in $PSScriptRoot remain; delete them manually if you want them gone." -ForegroundColor Green
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
            if ((Get-Flag 'remove-data') -eq 'yes') { Invoke-Reinstall } else { Invoke-Rebuild }
        }
        'reinstall' { Invoke-Reinstall }
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
