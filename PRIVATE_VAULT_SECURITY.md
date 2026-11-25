# Path Home Private Vault — security qualification notes

Status: opt-in development feature; do not treat as approved for production until the release gates pass.

## Scope and threat model
Existing Windows DPAPI and Android Keystore encryption remain the outer protection. Private Vault adds an independently chosen passphrase on each device. It protects the confidentiality and integrity of the current saved plan when the outer device encryption alone has been defeated, provided the passphrase remains secret and is sufficiently strong. It does not make a compromised running device trustworthy, prevent deletion or rollback, protect prior unvaulted backups, or guarantee erasure from managed-runtime memory. No user profiles, passphrases or keys are uploaded.

## Format
PHV1 authenticated envelope: four ASCII magic bytes, 16-byte random salt, 12-byte random AES-GCM nonce, ciphertext and 16-byte authentication tag. The header is authenticated as associated data. The 256-bit key is derived by PBKDF2-HMAC-SHA-256 with 600,000 iterations. Standard Go 1.24 cryptography and Java JCA are used, not custom ciphers. The format is bounded to 16 MiB. Windows wraps the binary envelope with existing user-scoped DPAPI. Android wraps a PHV1-prefixed Base64 envelope with the existing Keystore AES-GCM cipher. Keys and passphrases are not persisted. Key buffers are cleared on lock on a best-effort basis; other runtime copies may persist until reclaimed.

## User operation
- Windows: Privacy & lock in the navigation, enable with two matching passphrases, then lock manually or after five minutes of no app interaction. Windows session-lock notification also requests locking. A disabled modal may defer locking until it closes; unavailable timer is disclosed. Windows capture exclusion is best effort.
- Android: Privacy & Private Vault from the home screen or toolbar. Manual locking returns to a passphrase screen; backgrounding or five minutes of inactivity ends the protected activity and requires reopening. Android FLAG_SECURE is requested for vault-enabled activities. Permission and lifecycle behavior needs physical testing.
- The current baseline account encryption continues if Vault is not enabled. Minimum 15 characters; use at least five independently random words. A short predictable password is vulnerable to offline guessing. Passphrases are never sent to Genyxis. There is no password reset or publisher recovery key. Do not downgrade or run old builds against a vault. Close previous copies before updating. Existing Windows profile writes are serialized in the new app.
- Existing profiles remain readable before opt-in. On successful enrollment the current record is replaced through existing encrypted persistence; previous backups may not have this layer. Failed unlock must not change stored data. On a failed automatic lock save, the last persisted encrypted record is retained; unsaved edits may be lost.
- Code signing and verified update distribution are separate integrity requirements; encrypting an executable does not protect secrets during execution. These builds are unsigned QA builds.

## Mandatory verification
1. Go and JVM crypto tests: round trip, wrong key, fresh nonces/salts, changed header/ciphertext, truncated envelope and missing key.
2. Windows CI: synthetic enrollment, password mismatch, DPAPI unwrap still revealing authenticated ciphertext, wrong-password file hash unchanged, lock, reopen, correct password, data persistence, simulated session-lock message.
3. Physical Windows and Android: actual lock/unlock, idle timeout, backgrounding, rotation, OS lock, password entry/accessibility, UI capture exclusions and failure behavior. Test interruption, damaged ciphertext, failed writes, upgrade, device-specific key loss, and old backups with synthetic profiles.
4. Security review: inspect KDF latency on low-end phones, memory exposure, runtime malware limits, rollback risks, profile recovery, signing and upgrade behavior. No public production approval without review.

## References
- https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- https://pkg.go.dev/crypto/pbkdf2
- https://pkg.go.dev/crypto/cipher
- https://learn.microsoft.com/en-us/windows/win32/api/dpapi/nf-dpapi-cryptunprotectdata
- https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowdisplayaffinity
- https://developer.android.com/privacy-and-security/cryptography
- https://developer.android.com/reference/javax/crypto/SecretKeyFactory
