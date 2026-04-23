# Agenzia Scanner — one-line installer for Windows
# Usage: iwr -useb https://api.getagenzia.fr/scanner/install.ps1 | iex

$ErrorActionPreference = 'Stop'
$Repo = 'ramzilbscontact-ctrl/scanner'
$Version = if ($env:AGENZIA_VERSION) { $env:AGENZIA_VERSION } else { 'latest' }

function Info($msg) { Write-Host "i " -ForegroundColor Cyan -NoNewline; Write-Host $msg }
function Ok($msg)   { Write-Host "+ " -ForegroundColor Green -NoNewline; Write-Host $msg }
function Err($msg)  { Write-Host "x " -ForegroundColor Red -NoNewline; Write-Host $msg; exit 1 }

# ── Resolve version ──────────────────────────────
if ($Version -eq 'latest') {
    try {
        $release = Invoke-RestMethod -UseBasicParsing `
            "https://api.github.com/repos/$Repo/releases/latest"
        $Version = $release.tag_name
    } catch {
        Err "Could not resolve latest version"
    }
}
Info "Installing $Repo@$Version"

# ── Detect architecture ──────────────────────────
$archtag = if ([Environment]::Is64BitOperatingSystem) { 'x86_64' } else { '386' }
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { $archtag = 'arm64' }
Info "Detected windows/$archtag"

# ── Download ─────────────────────────────────────
$versionNum = $Version -replace '^v',''
$archive = "agenzia-scan_${versionNum}_windows_${archtag}.zip"
$url = "https://github.com/$Repo/releases/download/$Version/$archive"
$tmp = New-Item -ItemType Directory -Path "$env:TEMP\agenzia-scan-$(Get-Random)"

Info "Downloading $archive"
try {
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile "$tmp\archive.zip"
} catch {
    Err "Download failed — check version tag"
}
Expand-Archive -Path "$tmp\archive.zip" -DestinationPath $tmp -Force

# ── Install to %LOCALAPPDATA%\Agenzia ─────────────
$dest = "$env:LOCALAPPDATA\Agenzia"
New-Item -ItemType Directory -Path $dest -Force | Out-Null
Move-Item -Force "$tmp\agenzia-scan.exe" "$dest\agenzia-scan.exe"
Remove-Item -Recurse -Force $tmp

# ── Add to user PATH ─────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$dest*") {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$dest", 'User')
    Info "Added $dest to your user PATH (restart terminal to apply)"
}

Ok "Installed $dest\agenzia-scan.exe"
Write-Host
Info "Run a first scan now:"
Write-Host "  > agenzia-scan"
Write-Host
Info "Star us on GitHub if this helps 🙏 https://github.com/$Repo"
