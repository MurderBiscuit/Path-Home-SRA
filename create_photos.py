from PIL import Image,ImageEnhance,ImageOps,ImageDraw
from pathlib import Path
base=Path('/mnt/data/Reentry_Native_Apps_v1')
source=Image.open('/mnt/data/a_realistic_polished_desktop_application_screensh.png').convert('RGB')
regions={
 'release_plan':(320,328,645,472),
 'appointments':(669,328,974,472),
 'local_help':(997,327,1295,472),
 'money':(320,580,645,703),
 'family':(669,581,973,704),
 'progress':(990,581,1297,703),
 'hero':(805,38,1278,303),
}
android=base/'android/app/src/main/res/drawable-nodpi'; desktop=base/'desktop/windows/assets/photos'
android.mkdir(parents=True,exist_ok=True); desktop.mkdir(parents=True,exist_ok=True)
for key,rect in regions.items():
 im=source.crop(rect)
 im=ImageEnhance.Color(im).enhance(0.93)
 if key=='hero':
  im=ImageOps.fit(im,(520,260),method=Image.Resampling.LANCZOS)
 else:
  im=ImageOps.fit(im,(480,194),method=Image.Resampling.LANCZOS)
 im.save(android/f'{key}.jpg',quality=88,optimize=True)
 im.save(desktop/f'{key}.bmp')
 print(key, im.size)
