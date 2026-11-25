//go:build windows

package main

import (
 "encoding/json"
 "fmt"
 "path/filepath"
 "syscall"
 "time"
 "unsafe"
)

var (
 vaultKey []byte
 vaultEnvelope []byte
 vaultLocked bool
 vaultNotice string
 vaultLastActivity time.Time
 vaultPendingLock bool
 vaultNextAttempt time.Time
 vaultFailures int
 wts = syscall.NewLazyDLL("wtsapi32.dll")
 wtsRegister = wts.NewProc("WTSRegisterSessionNotification")
 wtsUnregister = wts.NewProc("WTSUnRegisterSessionNotification")
)
const vaultIdle = 5*time.Minute

// Serialize writes by holding an exclusive, profile-specific file handle until exit.
// Old releases lack this lease; close all old versions before upgrading.
func acquireProfileLease()(syscall.Handle,error){
 p:=storePath()
 if p==""{return syscall.InvalidHandle,fmt.Errorf("private profile directory unavailable")}
 name,err:=syscall.UTF16PtrFromString(filepath.Join(filepath.Dir(p),"profile.lock"))
 if err!=nil{return syscall.InvalidHandle,err}
 return syscall.CreateFile(name,syscall.GENERIC_READ|syscall.GENERIC_WRITE,0,nil,syscall.OPEN_ALWAYS,syscall.FILE_ATTRIBUTE_NORMAL,0)
}
func vaultPassword(x,y,id int)uintptr{
 c:=create("EDIT","",WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|0x0020|0x0080,x,y,560,42,hwnd,id)
 proc(pSendMessage,c,WM_SETFONT,font,1)
 proc(pSendMessage,c,0x00C5,256,0) // EM_SETLIMITTEXT
 return c
}
func takeVaultPassword(c uintptr)string{
 p:=getText(c)
 proc(pSetText,c,uintptr(unsafe.Pointer(ptr(""))))
 return p
}
func enableVault(pass,confirm string){
 if len(vaultEnvelope)>0 || len(vaultKey)>0 {return}
 if !validVaultPassphrase(pass)||pass!=confirm{
  alert("Use a matching passphrase of at least 15 characters. Spaces count; do not reuse your Windows password.")
  return
 }
 salt,err:=newVaultSalt()
 if err!=nil {alert("Could not create secure vault. No changes made.");return}
 key,err:=deriveVaultKey(pass,salt)
 if err!=nil {alert("Could not create secure vault. No changes made.");return}
 // An empty authenticated envelope provides the salt until the first real save.
 marker,err:=sealVault(key,salt,nil)
 if err!=nil {clear(key);alert("Could not create secure vault. No changes made.");return}
 vaultKey,vaultEnvelope=key,marker
 if !save(true) {
  clear(vaultKey);vaultKey=nil;vaultEnvelope=nil
  return
 }
 vaultLastActivity=time.Now()
 draw()
 alert("Private Vault enabled. Keep your passphrase somewhere safe. Genyxis cannot recover it.")
}
func unlockPrivateVault(pass string){
 if !vaultLocked{return}
 if time.Now().Before(vaultNextAttempt){alert("Please wait before trying again.");return}
 salt,err:=vaultSalt(vaultEnvelope)
 if err!=nil {alert("Private Vault is damaged. The original file has not been changed.");return}
 key,err:=deriveVaultKey(pass,salt)
 if err!=nil {alert("Unable to unlock the private vault.");return}
 raw,err:=openVault(key,vaultEnvelope)
 var next state
 if err==nil {err=json.Unmarshal(raw,&next)}
 clear(raw)
 if err!=nil{
  clear(key)
  vaultFailures++
  delay:=vaultFailures
  if delay>15 {delay=15}
  vaultNextAttempt=time.Now().Add(time.Duration(delay)*time.Second)
  alert("Passphrase incorrect or private vault damaged. Your saved plan was not changed.")
  return
 }
 if next.Checklist==nil {next.Checklist=map[string]bool{}}
 input=next
 vaultKey=key
 vaultLocked=false
 vaultFailures=0
 vaultNextAttempt=time.Time{}
 vaultNotice=""
 vaultPendingLock=false
 vaultLastActivity=time.Now()
 page="home"
 applyTheme()
 draw()
}
func lockPrivateVault(automatic bool){
 if vaultLocked || len(vaultKey)==0{return}
 if automatic && proc(user32.NewProc("IsWindowEnabled"),hwnd)==0 {vaultPendingLock=true;return}
 // Attempt to persist valid edits; if auto-lock cannot save, retain the last encrypted copy.
 if automatic {
  if !saveWithFeedback(true,false){vaultNotice="Some unsaved changes could not be saved. The last encrypted plan is preserved."}
 }else if !save(true){return}
 clear(vaultKey);vaultKey=nil
 input=state{Checklist:map[string]bool{}}
 vaultLocked=true
 vaultPendingLock=false
 page="home"
 draw()
}
func drawVaultUnlock(){
 clearPanel()
 navGlyph("brand",35,13,72)
 title(115,19,750,"PATH HOME · PRIVATE VAULT")
 title(284,155,760,"Your plan is locked")
 paragraph(284,215,740,"Enter your separate passphrase to unlock your private plan. There is no publisher reset or bypass.",3)
 text(284,325,700,28,"Vault passphrase",true)
 p:=vaultPassword(284,365,1300)
 btn(284,430,260,53,"Unlock private plan",1301,func(){unlockPrivateVault(takeVaultPassword(p))})
 btn(570,430,220,53,"Close application",1302,func(){proc(pDestroyWindow,hwnd)})
 if vaultNotice!="" {paragraph(284,515,740,vaultNotice,3)}
 proc(user32.NewProc("SetFocus"),p)
}
func drawPrivacy(x,w int){
 title(x,118,w,"Privacy & Private Vault")
 if len(vaultEnvelope)>0 {
  paragraph(x,175,745,"Your plan is protected by your passphrase and Windows account encryption. Lock it now or it locks after five minutes without interaction with Path Home.",4)
  btn(x,310,300,55,"Lock my private plan",1303,func(){lockPrivateVault(false)})
  paragraph(x,390,745,"A compromised unlocked device can still expose what you see and type. Keep your passphrase offline and use a trusted device.",4)
 }else{
  paragraph(x,175,745,"Add an independent passphrase to your Windows-encrypted plan. If you forget it, Genyxis cannot recover your data. Use at least 15 characters; five or more random words are better.",4)
  text(x,302,700,28,"New vault passphrase",true)
  p:=vaultPassword(x,337,1304)
  text(x,399,700,28,"Confirm passphrase",true)
  q:=vaultPassword(x,434,1305)
  btn(x,502,325,55,"Enable Private Vault",1306,func(){enableVault(takeVaultPassword(p),takeVaultPassword(q))})
  paragraph(x,584,745,"Enabling is optional and cannot be undone from this version. Existing encrypted copies or backups may have only account-level protection.",4)
 }
 btn(x,675,250,44,"Return Home",1307,func(){nav("home")})
}
