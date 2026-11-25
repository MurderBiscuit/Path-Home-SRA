package main

import (
 "bytes"
 "crypto/aes"
 "crypto/cipher"
 "crypto/pbkdf2"
 "crypto/rand"
 "crypto/sha256"
 "errors"
 "unicode/utf8"
)

// PHV1: magic (4) || random salt (16) || AES-GCM random nonce (12) || ciphertext + tag.
// The whole header is authenticated as associated data. The outer DPAPI wrapper remains.
const vaultMagic = "PHV1"
const vaultSaltSize = 16
const vaultRounds = 600000
const vaultHeaderSize = len(vaultMagic) + vaultSaltSize
const vaultMaxBytes = 16 * 1024 * 1024
var errVaultOpen = errors.New("passphrase incorrect or private vault damaged")

func validVaultPassphrase(p string) bool {
 n := utf8.RuneCountInString(p)
 return utf8.ValidString(p) && n >= 15 && n <= 256
}
func newVaultSalt() ([]byte,error) {
 salt:=make([]byte,vaultSaltSize)
 _,err:=rand.Read(salt)
 return salt,err
}
func isVault(raw []byte) bool { return bytes.HasPrefix(raw, []byte(vaultMagic)) }
func vaultSalt(raw []byte) ([]byte,error) {
 if len(raw)<vaultHeaderSize+28 || len(raw)>vaultMaxBytes || !isVault(raw) {return nil,errVaultOpen}
 return bytes.Clone(raw[len(vaultMagic):vaultHeaderSize]),nil
}
func deriveVaultKey(pass string,salt []byte)([]byte,error){
 if len(salt)!=vaultSaltSize || len(pass)>1024 {return nil,errVaultOpen}
 return pbkdf2.Key(sha256.New,pass,salt,vaultRounds,32)
}
func vaultCipher(key []byte)(cipher.AEAD,error){
 if len(key)!=32{return nil,errVaultOpen}
 block,err:=aes.NewCipher(key)
 if err!=nil{return nil,err}
 return cipher.NewGCMWithRandomNonce(block)
}
func sealVault(key,salt,plain []byte)([]byte,error){
 if len(salt)!=vaultSaltSize || len(plain)>vaultMaxBytes-vaultHeaderSize-28 {return nil,errors.New("invalid vault size")}
 a,err:=vaultCipher(key);if err!=nil{return nil,err}
 header:=append([]byte(vaultMagic),salt...)
 out:=append([]byte{},header...)
 return a.Seal(out,nil,plain,header),nil
}
func openVault(key,raw []byte)([]byte,error){
 if _,err:=vaultSalt(raw);err!=nil{return nil,err}
 a,err:=vaultCipher(key);if err!=nil{return nil,err}
 plain,err:=a.Open(nil,nil,raw[vaultHeaderSize:],raw[:vaultHeaderSize])
 if err!=nil{return nil,errVaultOpen}
 return plain,nil
}
