param(
    [Parameter(Mandatory=$true)][string]$Exe,
    [ValidateRange(1,100)][int]$Launches = 10
)
$ErrorActionPreference = 'Stop'
$Exe = (Resolve-Path -LiteralPath $Exe).Path
Add-Type -TypeDefinition @'
using System;
using System.Text;
using System.Runtime.InteropServices;
public static class PathHomeSmoke {
    [DllImport("user32.dll", SetLastError=true)]
    public static extern IntPtr GetDlgItem(IntPtr parent, int id);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)]
    public static extern IntPtr FindWindowEx(IntPtr parent, IntPtr after, string cls, string title);
    [DllImport("user32.dll", SetLastError=true, CharSet=CharSet.Unicode)]
    public static extern IntPtr SendMessageTimeout(IntPtr window, uint msg, UIntPtr w, IntPtr l, uint flags, uint ms, out UIntPtr result);
    [DllImport("user32.dll", EntryPoint="SendMessageTimeoutW", SetLastError=true, CharSet=CharSet.Unicode)]
    public static extern IntPtr SendText(IntPtr window, uint msg, UIntPtr w, string l, uint flags, uint ms, out UIntPtr result);
    [DllImport("user32.dll", EntryPoint="SendMessageTimeoutW", SetLastError=true, CharSet=CharSet.Unicode)]
    public static extern IntPtr ReadText(IntPtr window, uint msg, UIntPtr w, StringBuilder l, uint flags, uint ms, out UIntPtr result);
    [DllImport("user32.dll")]
    public static extern uint GetGuiResources(IntPtr process, uint flags);
}
'@
function Send-Bounded([IntPtr]$Window, [uint32]$Message, [uint32]$W = 0, [IntPtr]$L = [IntPtr]::Zero) {
    $result = [UIntPtr]::Zero
    $ok = [PathHomeSmoke]::SendMessageTimeout($Window, $Message, [UIntPtr]::new($W), $L, 3, 5000, [ref]$result)
    if ($ok -eq [IntPtr]::Zero) { throw "Window did not respond to message $Message." }
}
function Require-Control([IntPtr]$Window, [int]$Id) {
    $control = [PathHomeSmoke]::GetDlgItem($Window, $Id)
    if ($control -eq [IntPtr]::Zero) { throw "Expected control $Id was not rendered." }
    return $control
}
function Navigate([IntPtr]$Window, [int]$Button, [int]$Expected) {
    $control = Require-Control $Window $Button
    # Send the exact BN_CLICKED notification used by a native button.
    Send-Bounded $Window 0x111 $Button $control
    Send-Bounded $Window 0
    [void](Require-Control $Window $Expected)
}
function Set-State([IntPtr]$Window) {
    $edit = Require-Control $Window 1200
    if ($edit -eq [IntPtr]::Zero) { throw 'State edit control not found.' }
    $result = [UIntPtr]::Zero
    if ([PathHomeSmoke]::SendText($edit, 0xC, [UIntPtr]::Zero, 'MI', 3, 5000, [ref]$result) -eq [IntPtr]::Zero -or $result.ToUInt64() -eq 0) { throw 'Setting synthetic state failed.' }
}
function Assert-State([IntPtr]$Window) {
    $edit = Require-Control $Window 1200
    if ($edit -eq [IntPtr]::Zero) { throw 'State edit control not found.' }
    $buffer = [Text.StringBuilder]::new(64)
    $result = [UIntPtr]::Zero
    if ([PathHomeSmoke]::ReadText($edit, 0xD, [UIntPtr]::new([uint32]64), $buffer, 3, 5000, [ref]$result) -eq [IntPtr]::Zero -or $buffer.ToString() -ne 'MI') { throw 'Synthetic state did not survive navigation/relaunch.' }
}
$fixture = Join-Path ([IO.Path]::GetTempPath()) ('PathHome-QA-' + [Guid]::NewGuid().ToString('N'))
$oldApp = $env:APPDATA
$oldLocal = $env:LOCALAPPDATA
$env:APPDATA = Join-Path $fixture 'roaming'
$env:LOCALAPPDATA = Join-Path $fixture 'local'
[void](New-Item -ItemType Directory -Path $env:APPDATA, $env:LOCALAPPDATA)
$p = $null
try {
    for ($i=1; $i -le $Launches; $i++) {
        $p = Start-Process -FilePath $Exe -PassThru
        $deadline = [DateTime]::UtcNow.AddSeconds(30)
        $window = [IntPtr]::Zero
        do {
            $p.Refresh()
            if ($p.HasExited) { throw "Launch $i exited early: $($p.ExitCode)" }
            $window = $p.MainWindowHandle
            if ($window -ne [IntPtr]::Zero -and [PathHomeSmoke]::GetDlgItem($window, 200) -ne [IntPtr]::Zero) { break }
            Start-Sleep -Milliseconds 200
        } while ([DateTime]::UtcNow -lt $deadline)
        if ($window -eq [IntPtr]::Zero) { throw "Launch $i did not create its window." }
        Send-Bounded $window 0
        [void](Require-Control $window 200)
        Navigate $window 101 350
        if ($i -eq 1) { Set-State $window } else { Assert-State $window }
        Navigate $window 100 200
        Navigate $window 101 350
        Assert-State $window
        $baseline = 0
        for ($cycle=1; $cycle -le 3; $cycle++) {
            Navigate $window 102 360
            Navigate $window 103 450
            Navigate $window 104 375
            Navigate $window 105 380
            Navigate $window 106 501
            Navigate $window 107 400
            Navigate $window 108 410
            Navigate $window 109 815
            Navigate $window 110 431
            Navigate $window 900 902
            Navigate $window 100 200
            $gdi = [PathHomeSmoke]::GetGuiResources($p.Handle, 0)
            if ($gdi -eq 0) { throw 'Unable to measure GDI resources.' }
            if ($cycle -eq 1) { $baseline = $gdi }
            elseif ($gdi -gt ($baseline + 8)) { throw "GDI resource growth: $baseline -> $gdi" }
        }
        Send-Bounded $window 0x10
        if (-not $p.WaitForExit(10000)) { throw "Launch $i did not close cleanly." }
        if ($p.ExitCode -ne 0) { throw "Launch $i exited with code $($p.ExitCode)." }
        if (-not (Test-Path -LiteralPath (Join-Path $env:APPDATA 'PathHome\plan.dpapi'))) { throw 'Encrypted synthetic plan was not written.' }
        Write-Host "PASS launch $i/${Launches}: startup, navigation, persistence, GDI and exit"
        $p.Dispose()
        $p = $null
    }
} finally {
    if ($null -ne $p) {
        if (-not $p.HasExited) { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue }
        $p.Dispose()
    }
    $diagnostic = Join-Path $env:LOCALAPPDATA 'PathHome\startup.log'
    if (Test-Path -LiteralPath $diagnostic) { Get-Content -LiteralPath $diagnostic -Tail 30 }
    $env:APPDATA = $oldApp
    $env:LOCALAPPDATA = $oldLocal
    # Only the unique synthetic fixture created above is removed.
    Remove-Item -LiteralPath $fixture -Recurse -Force
}
