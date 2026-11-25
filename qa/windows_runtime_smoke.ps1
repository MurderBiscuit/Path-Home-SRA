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
    public delegate bool EnumProc(IntPtr h, IntPtr p);
    [StructLayout(LayoutKind.Sequential)] public struct RECT { public int L,T,R,B; }
    [DllImport("user32.dll")] public static extern bool EnumChildWindows(IntPtr h, EnumProc cb, IntPtr p);
    [DllImport("user32.dll")] public static extern bool EnumWindows(EnumProc cb, IntPtr p);
    [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr h, out uint pid);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetClassName(IntPtr h,StringBuilder b,int n);
    [DllImport("user32.dll")] public static extern bool IsWindow(IntPtr h);
    [DllImport("user32.dll")] public static extern bool IsWindowEnabled(IntPtr h);
    [DllImport("user32.dll",EntryPoint="PostMessageW")] public static extern bool Post(IntPtr h,uint m,UIntPtr w,IntPtr l);
    public static IntPtr[] Dialogs(uint pid) {
        var list=new System.Collections.Generic.List<IntPtr>();
        EnumWindows(delegate(IntPtr h,IntPtr p) {
            uint owner; GetWindowThreadProcessId(h,out owner);
            var cls=new StringBuilder(256); GetClassName(h,cls,256);
            if(owner==pid && IsWindowVisible(h) && cls.ToString()=="#32770") list.Add(h);
            return true;
        },IntPtr.Zero);
        return list.ToArray();
    }
    [DllImport("user32.dll")] public static extern IntPtr GetParent(IntPtr h);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr h);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
    [DllImport("user32.dll")] public static extern bool GetClientRect(IntPtr h, out RECT r);
    [DllImport("user32.dll")] public static extern bool ClientToScreen(IntPtr h, ref POINT p);
    [StructLayout(LayoutKind.Sequential)] public struct POINT { public int X,Y; }
    public static IntPtr[] Children(IntPtr h) {
        var list = new System.Collections.Generic.List<IntPtr>();
        EnumChildWindows(h, delegate(IntPtr c,IntPtr p) { if(GetParent(c)==h && IsWindowVisible(c))list.Add(c); return true; }, IntPtr.Zero);
        return list.ToArray();
    }
    public static string GeometryFailure(IntPtr h) {
        RECT client; POINT origin=new POINT();
        if(!GetClientRect(h,out client)||!ClientToScreen(h,ref origin)) return "Unable to inspect client geometry";
        var children=Children(h); var rects=new RECT[children.Length];
        for(int i=0;i<children.Length;i++){
            if(!GetWindowRect(children[i],out rects[i])) return "Unable to inspect child geometry";
            var r=rects[i];
            if(r.L<origin.X || r.T<origin.Y || r.R>origin.X+client.R || r.B>origin.Y+client.B) return "Child extends outside client: "+i+" at "+r.L+","+r.T+","+r.R+","+r.B;
            for(int j=0;j<i;j++){
                var q=rects[j];
                if(Math.Min(r.R,q.R)>Math.Max(r.L,q.L) && Math.Min(r.B,q.B)>Math.Max(r.T,q.T)) return "Overlapping child rectangles "+j+" and "+i+" at "+r.L+","+r.T;
            }
        }
        return "";
    }
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
function Assert-Geometry([IntPtr]$Window) {
    $problem = [PathHomeSmoke]::GeometryFailure($Window)
    if ($problem) { throw $problem }
}
function Read-Control([IntPtr]$Control) {
    $buffer = [Text.StringBuilder]::new(8192)
    $result = [UIntPtr]::Zero
    if ([PathHomeSmoke]::ReadText($Control, 0xD, [UIntPtr]::new([uint32]8192), $buffer, 3, 5000, [ref]$result) -eq [IntPtr]::Zero) { throw 'Cannot read control text.' }
    return $buffer.ToString()
}
function Set-Field([IntPtr]$Window,[int]$Id,[string]$Value) {
    $edit = Require-Control $Window $Id
    $result = [UIntPtr]::Zero
    if ([PathHomeSmoke]::SendText($edit, 0xC, [UIntPtr]::Zero, $Value, 3, 5000, [ref]$result) -eq [IntPtr]::Zero -or $result.ToUInt64() -eq 0) { throw "Cannot write field $Id." }
}
function Assert-Screen([IntPtr]$Window,[string]$Expected) {
    $screen = ([PathHomeSmoke]::Children($Window) | ForEach-Object { Read-Control $_ }) -join ' '
    if (-not $screen.Contains($Expected)) { throw "Expected screen text missing: $Expected" }
}
function Dialog-Action([IntPtr]$Window,[uint32]$ProcessId,[int]$Button,[string]$Expected) {
    if (@([PathHomeSmoke]::Dialogs($ProcessId)).Count -ne 0) { throw 'Stale modal before action.' }
    $h = Require-Control $Window $Button
    if (-not [PathHomeSmoke]::Post($Window,0x111,[UIntPtr]::new([uint32]$Button),$h)) { throw 'Cannot post dialog action.' }
    $deadline = [DateTime]::UtcNow.AddSeconds(8)
    $dialog = [IntPtr]::Zero
    do {
        $ds = @([PathHomeSmoke]::Dialogs($ProcessId))
        if ($ds.Count -eq 1 -and [PathHomeSmoke]::GetDlgItem($ds[0],1) -ne [IntPtr]::Zero) { $dialog = $ds[0]; break }
        Start-Sleep -Milliseconds 50
    } while ([DateTime]::UtcNow -lt $deadline)
    if ($dialog -eq [IntPtr]::Zero) { throw "Expected dialog not shown: $Expected" }
    $actual = ([PathHomeSmoke]::Children($dialog) | ForEach-Object { Read-Control $_ }) -join ' '
    $ok = Require-Control $dialog 1
    if (-not [PathHomeSmoke]::Post($dialog,0x111,[UIntPtr]::new([uint32]1),$ok)) { throw 'Cannot dismiss modal.' }
    $deadline = [DateTime]::UtcNow.AddSeconds(8)
    do {
        if (-not [PathHomeSmoke]::IsWindow($dialog) -and @([PathHomeSmoke]::Dialogs($ProcessId)).Count -eq 0 -and [PathHomeSmoke]::IsWindowEnabled($Window)) {
            Send-Bounded $Window 0
            if ($actual.IndexOf($Expected,[StringComparison]::OrdinalIgnoreCase) -lt 0) { throw "Unexpected dialog. Expected: $Expected; actual: $actual" }
            return
        }
        Start-Sleep -Milliseconds 50
    } while ([DateTime]::UtcNow -lt $deadline)
    throw 'Dialog did not close.'
}
function Navigate([IntPtr]$Window, [int]$Button, [int]$Expected) {
    $control = Require-Control $Window $Button
    # Send the exact BN_CLICKED notification used by a native button.
    Send-Bounded $Window 0x111 $Button $control
    Send-Bounded $Window 0
    [void](Require-Control $Window $Expected)
    Assert-Geometry $Window
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
        Assert-Geometry $window
        Navigate $window 101 350
        if ($i -eq 1) { Set-State $window } else { Assert-State $window }
        Navigate $window 100 200
        Navigate $window 101 350
        Assert-State $window
        if ($i -eq 1) {
            $plan = Join-Path $env:APPDATA 'PathHome\plan.dpapi'
            $oldHash = (Get-FileHash -LiteralPath $plan -Algorithm SHA256).Hash
            Set-Field $window 1200 'ZZ'
            Dialog-Action $window $p.Id 100 'valid two-letter state abbreviation'
            Set-Field $window 1200 'MI'
            Set-Field $window 1205 '2026-02-30'
            Dialog-Action $window $p.Id 100 'Enter dates as YYYY-MM-DD'
            if ((Get-FileHash -LiteralPath $plan -Algorithm SHA256).Hash -ne $oldHash) { throw 'Invalid release data replaced encrypted plan.' }
            Set-Field $window 1205 ''
            Navigate $window 104 375
            Set-Field $window 1208 '2030-02-31'
            Dialog-Action $window $p.Id 376 'Enter dates as YYYY-MM-DD'
            Dialog-Action $window $p.Id 375 'Enter dates as YYYY-MM-DD'
            Set-Field $window 1208 '2030-01-15'
            Navigate $window 376 375
            Navigate $window 110 431
            Dialog-Action $window $p.Id 432 'Turn on private reminders first'
            Navigate $window 431 431
            Assert-Screen $window 'Optional reminder: ON'
            Dialog-Action $window $p.Id 432 'Are you in the right place'
            Navigate $window 431 431
            Assert-Screen $window 'Optional reminder: OFF'
        }
        # Confirm optional amounts remain unknown instead of silently becoming zero.
        Navigate $window 105 380
        if ($i -eq 1) {
            Set-Field $window 1210 '20'
            Set-Field $window 1211 ''
            Set-Field $window 1212 '0'
            Set-Field $window 1213 '5'
            Navigate $window 380 380
        }
        Assert-Screen $window 'Partial shortfall'
        if ((Read-Control (Require-Control $window 1211)) -ne '') { throw 'Unknown fee became zero.' }
        if ((Read-Control (Require-Control $window 1212)) -ne '0.00') { throw 'Explicit zero was not preserved.' }
        Navigate $window 903 910
        if ($i -eq 1) {
            foreach ($theme in 911..917) { Navigate $window $theme 910 }
            Navigate $window 915 910
        }
        Assert-Screen $window 'Current palette: Nebula Orchid'
        Navigate $window 900 902
        $mission = Require-Control $window 902
        $original = Read-Control $mission
        # EM_LINESCROLL, then return to start; this is a text/scroll responsiveness gate,
        # not a substitute for pixel-level physical-device visual review.
        Send-Bounded $mission 0xB6 0 ([IntPtr]10)
        Send-Bounded $mission 0xB6 0 ([IntPtr](-10))
        if ((Read-Control $mission) -ne $original) { throw 'Mission content changed while scrolling.' }
        Navigate $window 100 200
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
