$ErrorActionPreference = 'Stop'
Start-Transcript -Path C:\vmctl-bootstrap.log -Force
try {
    $msi = Join-Path $PSScriptRoot 'OpenSSH.msi'
    $expected = (Get-Content (Join-Path $PSScriptRoot 'OpenSSH.sha256') -Raw).Trim()
    if ((Get-FileHash $msi -Algorithm SHA256).Hash -ne $expected) { throw 'OpenSSH hash mismatch' }
    $signature = Get-AuthenticodeSignature $msi
    if ($signature.Status -ne 'Valid' -or $signature.SignerCertificate.Subject -notmatch 'O=Microsoft Corporation') { throw 'OpenSSH Microsoft signature verification failed' }
    $process = Start-Process msiexec.exe -ArgumentList @('/i', $msi, 'ADDLOCAL=Server', '/qn', '/norestart', '/l*v', 'C:\vmctl-openssh-msi.log') -Wait -PassThru
    if ($process.ExitCode -notin @(0,3010)) { throw "OpenSSH MSI failed: $($process.ExitCode)" }
    Stop-Service sshd -ErrorAction SilentlyContinue
    $dir = 'C:\ProgramData\ssh'
    New-Item -ItemType Directory -Force $dir | Out-Null
    $key = Join-Path $dir 'administrators_authorized_keys'
    Get-Content (Join-Path $PSScriptRoot 'client-key.pub') | Set-Content -Encoding ascii $key
    & icacls.exe $key /inheritance:r /grant:r '*S-1-5-18:F' '*S-1-5-32-544:F'
    if ($LASTEXITCODE -ne 0) { throw 'Public key ACL failed' }
    @'
Port 22
PubkeyAuthentication yes
PasswordAuthentication no
KbdInteractiveAuthentication no
AllowUsers vmctl
AuthorizedKeysFile __PROGRAMDATA__/ssh/administrators_authorized_keys
Subsystem sftp sftp-server.exe
'@ | Set-Content -Encoding ascii (Join-Path $dir 'sshd_config')
    & 'C:\Program Files\OpenSSH\ssh-keygen.exe' -A
    if ($LASTEXITCODE -ne 0) { throw 'Host key generation failed' }
    & 'C:\Program Files\OpenSSH\sshd.exe' -t
    if ($LASTEXITCODE -ne 0) { throw 'Invalid sshd configuration' }
    New-ItemProperty -Path HKLM:\SOFTWARE\OpenSSH -Name DefaultShell -Value 'C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe' -PropertyType String -Force | Out-Null
    Set-Service sshd -StartupType Automatic
    if (-not (Get-NetFirewallRule -Name vmctl-ssh -ErrorAction SilentlyContinue)) {
        New-NetFirewallRule -Name vmctl-ssh -DisplayName vmctl-ssh -Direction Inbound -Protocol TCP -LocalPort 22 -Action Allow | Out-Null
    }
    Start-Service sshd
    if ((Get-Service sshd).Status -ne 'Running') { throw 'sshd not running' }
    'ready' | Set-Content C:\vmctl-ssh-ready.txt
    $extension = Join-Path $PSScriptRoot 'post-install.ps1'
    if (Test-Path $extension) {
        $extra = Start-Process powershell.exe -ArgumentList @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $extension) -Wait -PassThru
        if ($extra.ExitCode -ne 0) { throw "Post-install failed: $($extra.ExitCode)" }
    }
    'installed' | Set-Content C:\vmctl-install-complete.txt
} catch {
    $_ | Out-String | Set-Content C:\vmctl-bootstrap-error.txt
    exit 1
} finally { Stop-Transcript }
