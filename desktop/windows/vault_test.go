package main
import (
 "bytes"
 "testing"
)
func TestVaultRoundTripAndTamper(t *testing.T) {
 pass:="a synthetic test phrase only"
 salt,err:=newVaultSalt();if err!=nil{t.Fatal(err)}
 key,err:=deriveVaultKey(pass,salt);if err!=nil{t.Fatal(err)}
 defer clear(key)
 plain:=[]byte(`{"test":"synthetic private plan"}`)
 raw,err:=sealVault(key,salt,plain);if err!=nil{t.Fatal(err)}
 second,err:=sealVault(key,salt,plain);if err!=nil{t.Fatal(err)}
 if bytes.Equal(raw,second)||bytes.Contains(raw,plain){t.Fatal("nonce or plaintext exposure")}
 got,err:=openVault(key,raw);if err!=nil||!bytes.Equal(got,plain){t.Fatal("round trip failed")}
 clear(got)
 wrong,err:=deriveVaultKey("another synthetic passphrase",salt);if err!=nil{t.Fatal(err)}
 defer clear(wrong)
 if _,err=openVault(wrong,raw);err==nil{t.Fatal("wrong passphrase accepted")}
 for _,i:=range []int{0,4,20,len(raw)-1}{
  damaged:=bytes.Clone(raw);damaged[i]^=1
  if _,err=openVault(key,damaged);err==nil{t.Fatalf("tampering accepted at %d",i)}
 }
 for _,n:=range []int{0,4,20,47}{
  if _,err=openVault(key,raw[:n]);err==nil{t.Fatalf("truncated vault accepted at %d",n)}
 }
 if _,err=openVault(nil,raw);err==nil{t.Fatal("missing key accepted")}
}
func TestVaultPassphrasePolicy(t *testing.T){
 if validVaultPassphrase("123456")||!validVaultPassphrase("five synthetic test words"){t.Fatal("bad passphrase policy")}
 a,err:=newVaultSalt();if err!=nil{t.Fatal(err)}
 b,err:=newVaultSalt();if err!=nil||bytes.Equal(a,b){t.Fatal("salt generation failed")}
}
