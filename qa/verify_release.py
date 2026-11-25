#!/usr/bin/env python3
"""Non-device QC gate for Path Home; no personal user data used."""
from pathlib import Path
from datetime import date
import json, hashlib, re, struct, sys, xml.etree.ElementTree as ET
root=Path(__file__).resolve().parents[1]
checks=[]
def check(name,condition):
    checks.append((name, bool(condition)))
    print(('PASS' if condition else 'FAIL') + ' ' + name)

def h(path):return hashlib.sha256(path.read_bytes()).hexdigest()
data=root/'shared/reentry-public.json'
try: d=json.loads(data.read_text())
except Exception as exc: raise SystemExit('FAIL shared JSON: '+str(exc))
allstates=[x['jurisdiction_code'] for x in d['jurisdictions']]
check('50 states + DC + federal as distinct legal tracks',len(allstates)==len(set(allstates))==52 and {'FED','DC','MI','CA','NY','AK','HI'} <= set(allstates))
check('Public database includes legal citations and local help',len(d['legal_rules'])>100 and len(d['local_providers'])>=6)
check('Exact same public data in Windows and Android', h(data)==h(root/'desktop/windows/assets/reentry-public.json')==h(root/'android/app/src/main/assets/reentry-public.json'))
mission=(root/'desktop/windows/assets/mission.txt').read_text().strip()
mission_md=(root/'MISSION.md').read_text()
mission_paragraphs=[x.strip() for x in mission.split('\n\n') if x.strip()]
check('Identical complete approved mission across sources',h(root/'desktop/windows/assets/mission.txt')==h(root/'android/app/src/main/assets/mission.txt') and len(mission_paragraphs)==6 and all(x in mission_md for x in mission_paragraphs[:-1]) and '**Everyone deserves a Path Home.**' in mission_md)
approved=(root/'branding/approved_checksums.sha256').read_text().splitlines()
check('Approved product art/icons match immutable SHA-256 baseline',len(approved)>=50 and all(len(line.split('  ',1))==2 and (root/line.split('  ',1)[1]).is_file() and h(root/line.split('  ',1)[1])==line.split('  ',1)[0] for line in approved))
check('Provider source/date/eligibility contract',all(p.get('state_code') and p.get('county') and p.get('name') and p.get('source_url','').startswith('https://') and p.get('verified_on') and p.get('eligibility') and p.get('before_you_go') and p.get('if_denied') and (p.get('phone') or p.get('website')) for p in d['local_providers']))
check('Jackson pilot not misrepresented as entire country',all(p['state_code']=='MI' and p['county']=='Jackson' for p in d['local_providers']))
check('Provider IDs unique',len({p['id'] for p in d['local_providers']})==len(d['local_providers']))
check('No unverifiable legal rule URLs',all(r.get('source_url','').startswith('https://') for r in d['legal_rules']))
java=(root/'android/app/src/main/java/org/pathhome/sra/MainActivity.java').read_text()
go=(root/'desktop/windows/main.go').read_text()
manifest=(root/'android/app/src/main/AndroidManifest.xml').read_text()
ET.fromstring(manifest)
check('Android forbids Android Auto Backup', 'android:allowBackup="false"' in manifest and 'android:fullBackupContent="false"' in manifest)
check('Android refuses cleartext requests', 'android:usesCleartextTraffic="false"' in manifest)
check('No silent profile overwrite after decrypt error', 'profileLocked=true' in java and 'No information was overwritten' in java)
check('Android profile save is batched and flushed on pause', 'postDelayed(deferredSave' in java and 'onPause(){flushSave()' in java)
check('Android duplicate local-211 entry removed',java.count('add(no,soft("Find local assistance through 211"')==1)
check('Android legal results separate page, not stacked onto input', 'case "allowed_results":results();break;' in java and 'show("allowed_results")' in java)
check('Android legal search uses explicit federal-supervision choice', 'Federal supervised release'.encode() is not None and '"FED".equals(state)' in java and 'What kind of supervision applies?' in java)
check('Android state change resets visible county', 'if(countyInput!=null)countyInput.setText("")' in java)
check('Android supervision end validation blocks next step', 'if(!WorkflowRules.datesInOrder' in java)
check('Android appointment requires real date and time', 'if(d.isEmpty())' in java and 'if(!t.matches' in java and 'Confirm all details against the original notice' in java)
check('Android money distinguishes incomplete fee estimates', 'Partial shortfall (fees missing)' in java and 'WorkflowRules.validMoney' in java)
check('Android location stays optional and local', 'Are you in the right place, and the right time?' in java and 'permission.ACCESS_BACKGROUND_LOCATION' not in manifest)
check('Windows refuses navigation if encrypted save fails', 'if !save(true) {' in go and 'if !vaultLocked && !save(true) {\n\t\t\treturn 0' in go)
check('Windows invalid money is not silently zeroed on save', 'validateMoneyField(getText(c))' in go)
check('Windows legal question results redraw rather than stacking', 'if save(true) {\n\t\t\tdraw()' in go)
check('Windows reporting/legal warning shown', 'signed release conditions' in go.lower())
pe=(root/'desktop/windows/PathHome_Windows_x64.exe').read_bytes()
peoff=struct.unpack_from('<I',pe,0x3c)[0]
check('Installable Windows PE executable with icon resource',pe[:2]==b'MZ' and pe[peoff:peoff+4]==b'PE\0\0' and b'.rsrc\0' in pe[:8192])
check('No .env or key material in tracked source',not any(x.is_file() and (x.name.endswith('.pem') or x.name=='.env' or x.name.endswith('.key')) for x in root.rglob('*')))
check('GitHub public repository excludes actual profiles',not list(root.rglob('plan.dpapi')) and not list(root.rglob('plan.json')))

# These are structural checks only; executable vault tests run in Go/JVM and native Windows CI.
vault_go=(root/'desktop/windows/vault.go').read_text()
vault_win=(root/'desktop/windows/vault_windows.go').read_text()
vault_java=(root/'android/app/src/main/java/org/pathhome/sra/PrivateVault.java').read_text()
check('Vault uses standard authenticated encryption and independent key derivation on both platforms',
      'pbkdf2.Key' in vault_go and 'vaultRounds = 600000' in vault_go and 'NewGCMWithRandomNonce' in vault_go
      and 'PBKDF2WithHmacSHA256' in vault_java and '600000' in vault_java and 'AES/GCM/NoPadding' in vault_java)
check('Windows vault keeps account encryption and refuses locked writes',
      'if vaultLocked { return fmt.Errorf("private vault is locked") }' in go and 'protectLocal(payload)' in go
      and 'clear(vaultKey)' in vault_win and 'vaultIdle = 5*time.Minute' in vault_win)
check('Android vault retains Keystore encryption and blocks locked writes',
      'PrivateVault.seal' in java and 'encrypt(payload)' in java and 'if(profileLocked||vaultLocked)' in java
      and 'FLAG_SECURE' in java and 'PrivateVault.wipe(vaultKey)' in java)

# Prevent the duplicate Android lifecycle override that passed source-only checks in v0.9.5.
android_top_level_methods=re.findall(r'^    (?:@Override\s+)?(?:public|private|protected)\s+(?:(?:static|final|synchronized)\s+)*(?:[\w<>\[\].]+)\s+([A-Za-z_][A-Za-z_0-9]*)\s*\(([^)]*)\)',java,re.M)
sigs=[name+'('+','.join(p.strip().split()[0] for p in args.split(',') if p.strip())+')' for name,args in android_top_level_methods]
check('Android Activity has no duplicate lifecycle/method signatures',len(sigs)==len(set(sigs)) and sigs.count('onBackPressed()')==1)
# Catch missing icons before Android resource compilation.
android_resources={p.stem for p in (root/'android/app/src/main/res/drawable-nodpi').glob('*')}
referenced_drawables=set(re.findall(r'R\.drawable\.([A-Za-z_][A-Za-z_0-9]*)',java))
check('Every referenced Android drawable exists',referenced_drawables <= android_resources)
# Legal records must not be falsely labelled as personal permission.
check('Legal records do not assert personal legal permission',all(str(rule.get('personal_permission','')).lower() not in ('granted','yes','true','allowed') for rule in d['legal_rules']))
# All state-specific providers must have a genuine, nonempty geographic match.
check('Local referrals anchored to valid jurisdictions',all(p['state_code'] in allstates and p['county'].strip() for p in d['local_providers']))
# Do not bundle end-user identity data, or key material in the public source.
forbidden_names={'private_plan','plan.dpapi','plan.json','.env','users.csv','user_profiles.csv','credentials.json'}
check('Public source has no profile or credential files',not any(f.is_file() and (f.name in forbidden_names or f.suffix.lower() in {'.p12','.pfx','.keystore','.jks'}) for f in root.rglob('*')))

fail=[x for x,ok in checks if not ok]
print('\nTOTAL',len(checks),'passed',len(checks)-len(fail),'failed',len(fail))
if fail: sys.exit(1)
