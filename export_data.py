import sqlite3, json, os
src='/mnt/data/reentry_navigator_v05/reentry_navigator.sqlite'
cn=sqlite3.connect(src); cn.row_factory=sqlite3.Row
out={}
for table in ('jurisdictions','legal_rules','assistance_programs','birth_record_fee_observations','aid_referral_networks','eligibility_fallbacks'):
 out[table]=[dict(r) for r in cn.execute('SELECT * FROM '+table)]
for path in ('/mnt/data/Reentry_Native_Apps_v1/shared/reentry-public.json','/mnt/data/Reentry_Native_Apps_v1/desktop/windows/reentry-public.json','/mnt/data/Reentry_Native_Apps_v1/android/app/src/main/assets/reentry-public.json'):
 with open(path,'w') as f: json.dump(out,f,ensure_ascii=False,separators=(',',':'))
 print(path, os.stat(path).st_size)
