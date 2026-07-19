param(
    [string]$Version = $env:TAB_VERSION,
    [string]$InstallDir = $env:TAB_INSTALL_DIR,
    [string]$ReleaseBaseUrl = $env:TAB_RELEASE_BASE_URL
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repository = "hmmhmmhm/tmux-attach-browser"
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Join-Path $HOME ".local\bin"
}

$runtimeArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
switch ($runtimeArch) {
    "x64" { $arch = "amd64" }
    "arm64" { $arch = "arm64" }
    default { throw "Unsupported Windows architecture: $runtimeArch" }
}

$archive = "tab_windows_${arch}.zip"
if ([string]::IsNullOrWhiteSpace($ReleaseBaseUrl)) {
    if ([string]::IsNullOrWhiteSpace($Version)) {
        $ReleaseBaseUrl = "https://github.com/$repository/releases/latest/download"
    } else {
        $ReleaseBaseUrl = "https://github.com/$repository/releases/download/$Version"
    }
}

$temporaryDir = Join-Path ([System.IO.Path]::GetTempPath()) ("tab-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $temporaryDir | Out-Null

try {
    $archivePath = Join-Path $temporaryDir $archive
    $checksumsPath = Join-Path $temporaryDir "checksums.txt"
    Invoke-WebRequest -UseBasicParsing -Uri "$ReleaseBaseUrl/$archive" -OutFile $archivePath
    Invoke-WebRequest -UseBasicParsing -Uri "$ReleaseBaseUrl/checksums.txt" -OutFile $checksumsPath

    $checksumLine = Get-Content $checksumsPath | Where-Object {
        $_ -match "^([0-9a-fA-F]{64})\s+\*?$([regex]::Escape($archive))$"
    } | Select-Object -First 1
    if (-not $checksumLine) {
        throw "Checksum entry for $archive was not found"
    }
    $expected = ($checksumLine -split "\s+")[0].ToLowerInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        throw "Checksum verification failed for $archive"
    }

    $expanded = Join-Path $temporaryDir "expanded"
    Expand-Archive -Path $archivePath -DestinationPath $expanded
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Force (Join-Path $expanded "tab.exe") (Join-Path $InstallDir "tab.exe")
    Write-Output "Installed tab to $(Join-Path $InstallDir 'tab.exe')"

    if (-not (Get-Command tmux -ErrorAction SilentlyContinue)) {
        Write-Warning "tmux was not found. WSL2 is the supported Windows environment for tab."
    }
} finally {
    Remove-Item -Recurse -Force $temporaryDir -ErrorAction SilentlyContinue
}
