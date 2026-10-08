# Installs acta: downloads the release zip, checks its SHA-256 against
# checksums.txt, and copies acta.exe into the install dir. It never edits the
# user PATH.
# The body runs in its own scope, so its settings and variables do not
# stay in the session when the script runs under "irm | iex".
& {
    $ErrorActionPreference = 'Stop'
    # The progress bar makes downloads very slow in Windows PowerShell 5.1.
    $ProgressPreference = 'SilentlyContinue'

    $base = if ($env:ACTA_DOWNLOAD_URL) { $env:ACTA_DOWNLOAD_URL } else { 'https://github.com/iyay/acta/releases' }
    $dir = if ($env:ACTA_INSTALL_DIR) { $env:ACTA_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'acta\bin' }

    # Say what is wrong and stop. Nothing is installed yet when this runs.
    function Fail($msg) {
        throw "install.ps1: $msg"
    }

    # A 32-bit PowerShell on a 64-bit Windows reports the wrong CPU, so look at
    # the real one first.
    $cpu = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
    switch ($cpu) {
        'AMD64' { $arch = 'amd64' }
        'ARM64' { $arch = 'arm64' }
        default { Fail "unsupported CPU: $cpu" }
    }

    if ($env:ACTA_VERSION) {
        # Release tags start with v, so 0.1.44 and v0.1.44 mean the same.
        $tag = if ($env:ACTA_VERSION.StartsWith('v')) { $env:ACTA_VERSION } else { "v$($env:ACTA_VERSION)" }
        $url = "$base/download/$tag"
    } else {
        $url = "$base/latest/download"
    }

    $asset = "acta_windows_$arch.zip"
    $tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("acta-install-" + [guid]::NewGuid())
    New-Item -ItemType Directory -Path $tmp | Out-Null

    try {
        Write-Host "Downloading $asset"
        Invoke-WebRequest -UseBasicParsing -Uri "$url/$asset" -OutFile (Join-Path $tmp $asset)
        Invoke-WebRequest -UseBasicParsing -Uri "$url/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

        # A "*" before the name marks binary mode in some checksum files.
        $want = $null
        foreach ($line in Get-Content (Join-Path $tmp 'checksums.txt')) {
            $parts = $line.Trim() -split '\s+'
            if ($parts.Count -ge 2 -and ($parts[1] -eq $asset -or $parts[1] -eq "*$asset")) {
                $want = $parts[0]
                break
            }
        }
        if (-not $want) { Fail "no checksum line for $asset" }
        $got = (Get-FileHash -Algorithm SHA256 -Path (Join-Path $tmp $asset)).Hash
        # -ne ignores case, and Get-FileHash prints upper case.
        if ($got -ne $want) { Fail "checksum mismatch for $asset" }

        Expand-Archive -Path (Join-Path $tmp $asset) -DestinationPath $tmp -Force
        if (-not (Test-Path (Join-Path $tmp 'acta.exe'))) { Fail "$asset has no acta.exe" }

        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        $target = Join-Path $dir 'acta.exe'
        # Copy under a temp name, then rename, so a half-copied acta.exe never sits in the dir.
        $part = Join-Path $dir (".acta.$PID.tmp")
        Copy-Item -Path (Join-Path $tmp 'acta.exe') -Destination $part -Force
        Move-Item -Path $part -Destination $target -Force
        Write-Host "Installed $target"
    } finally {
        Remove-Item -Path $tmp -Recurse -Force -ErrorAction SilentlyContinue
    }

    $onPath = $false
    foreach ($p in ($env:PATH -split ';')) {
        if ($p.TrimEnd('\') -eq $dir.TrimEnd('\')) { $onPath = $true }
    }
    if (-not $onPath) {
        Write-Host "$dir is not on your PATH. Add it to your user PATH in Environment Variables."
        Write-Host "For this window only, run:"
        Write-Host "  `$env:PATH = `"$dir;`$env:PATH`""
    }

    # acta uses git, and the hook scripts need bash, which comes with Git for Windows.
    $gitBash = @(
        (Join-Path $env:ProgramFiles 'Git\bin\bash.exe'),
        (Join-Path $env:LOCALAPPDATA 'Programs\Git\bin\bash.exe')
    ) | Where-Object { Test-Path $_ }
    if (-not (Get-Command git.exe -ErrorAction SilentlyContinue) -and -not $gitBash) {
        Write-Warning "Neither git.exe nor Git Bash was found. Install Git for Windows, or acta and its hooks will not work."
    }

    # Under "irm | iex" the console is still the keyboard, but a pipe or a CI run
    # has no one to answer, so setup only runs when input is a real console.
    if ([Console]::IsInputRedirected) {
        Write-Host "run: acta setup"
    } else {
        & $target setup
        if ($LASTEXITCODE -ne 0) { Write-Host "run: acta setup" }
    }
}
