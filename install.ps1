#Requires -Version 5.1

$ErrorActionPreference = "Stop"

$Repo = "mitchell-wallace/plaqq"
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\plaqq"

# Determine architecture
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

# Fetch latest release
$LatestUrl = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $LatestUrl -UseBasicParsing
$Tag = $Release.tag_name

if (-not $Tag) {
    Write-Error "Failed to fetch latest release tag"
    exit 1
}

$Version = $Tag -replace '^v',''
$Asset = "plaqq_${Version}_windows_${Arch}.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$Asset"
$TempFile = Join-Path $env:TEMP $Asset

Write-Host "Downloading $Asset..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing

# Create install directory
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Extract to a temp dir first, then move the binary into place. Windows won't let
# you overwrite or delete a running .exe, but it will let you rename it aside, so
# if plaqq is currently running we move the old exe to plaqq.old.exe before
# dropping the new one in -- the install succeeds without killing the process.
$TempExtract = Join-Path $env:TEMP ("plaqq_extract_" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $TempExtract | Out-Null
Expand-Archive -Path $TempFile -DestinationPath $TempExtract -Force

$TargetExe = Join-Path $InstallDir "plaqq.exe"
if (Test-Path $TargetExe) {
    $OldExe = Join-Path $InstallDir "plaqq.old.exe"
    Remove-Item -Path $OldExe -Force -ErrorAction SilentlyContinue
    try {
        Move-Item -Path $TargetExe -Destination $OldExe -Force
    } catch {
        Write-Error "Could not replace $TargetExe (is plaqq running?). Close all plaqq windows and retry."
        exit 1
    }
}

# Move extracted files into the install dir (plaqq.exe plus any extras)
Get-ChildItem -Path $TempExtract | ForEach-Object {
    Move-Item -Path $_.FullName -Destination (Join-Path $InstallDir $_.Name) -Force
}

Remove-Item -Path $TempExtract -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path $TempFile -Force

Write-Host "Installed plaqq.exe to $InstallDir"

# Update user PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "Added $InstallDir to user PATH"
} else {
    Write-Host "$InstallDir already in user PATH"
}

Write-Host "Installation complete. Restart your terminal to use plaqq."
