from pathlib import Path
import argparse
import struct
from math import ceil
cli=argparse.ArgumentParser(description='Embed the approved Path Home ICO in a freshly compiled unsigned Windows PE')
cli.add_argument('--exe',required=True)
cli.add_argument('--ico',required=True)
args=cli.parse_args()
exe=Path(args.exe)
ico=Path(args.ico).read_bytes()
data=bytearray(exe.read_bytes())
pe=struct.unpack_from('<I',data,0x3c)[0]
assert data[pe:pe+4]==b'PE\0\0'
coff=pe+4
num=struct.unpack_from('<H',data,coff+2)[0];opt_size=struct.unpack_from('<H',data,coff+16)[0]
opt=coff+20
assert struct.unpack_from('<H',data,opt)[0]==0x20b
section_align=struct.unpack_from('<I',data,opt+32)[0]
file_align=struct.unpack_from('<I',data,opt+36)[0]
headers=struct.unpack_from('<I',data,opt+60)[0]
section_table=opt+opt_size
assert section_table+(num+1)*40<=headers,(section_table+(num+1)*40,headers)
sects=[]
for n in range(num):
 o=section_table+40*n
 name=data[o:o+8].strip(b'\0')
 virtualsize,rva,rawsize,rawptr=struct.unpack_from('<IIII',data,o+8)
 sects.append((name,virtualsize,rva,rawsize,rawptr))
assert not any(x[0]==b'.rsrc' for x in sects)
assert struct.unpack_from('<II',data,opt+112+8*2)==(0,0)
def align(v,a):return (v+a-1)//a*a
new_rva=align(max(rva+max(vsz,rsz) for _,vsz,rva,rsz,ptr in sects),section_align)
rawptr=align(len(data),file_align)
# Parse all icon frames and group icon directory; resource types 3 and 14.
reserved,itype,count=struct.unpack_from('<HHH',ico,0)
assert (reserved,itype)==(0,1)
frames={};group=bytearray(struct.pack('<HHH',0,1,count))
for idx in range(count):
 entry=ico[6+16*idx:6+16*(idx+1)]
 width,height,colors,rsrv,planes,bpp,size,offset=struct.unpack('<BBBBHHII',entry)
 frames[idx+1]=ico[offset:offset+size]
 assert len(frames[idx+1])==size
 group.extend(entry[:12]);group.extend(struct.pack('<H',idx+1))
# Tree is RT_ICON -> n -> en-US -> icon blob and RT_GROUP_ICON -> 1 -> en-US -> group blob
root={3:{i:{1033:frame} for i,frame in frames.items()},14:{1:{1033:bytes(group)}}}
res=bytearray(); pending=[]
def reserve(n):
 off=align(len(res),4)
 if off>len(res):res.extend(b'\0'*(off-len(res)))
 res.extend(b'\0'*n)
 return off
def make_tree(children):
 ids=sorted(children.items())
 off=reserve(16+8*len(ids))
 struct.pack_into('<IIHHHH',res,off,0,0,0,0,0,len(ids))
 for i,(ident,child) in enumerate(ids):
  if isinstance(child,dict): target=make_tree(child)|0x80000000
  else:
   target=reserve(16);pending.append((target,child))
  struct.pack_into('<II',res,off+16+i*8,ident,target)
 return off
assert make_tree(root)==0
for data_entry,blob in pending:
 body=reserve(len(blob));res[body:body+len(blob)]=blob
 struct.pack_into('<IIII',res,data_entry,new_rva+body,len(blob),0,0)
rawsize=align(len(res),file_align)
res.extend(b'\0'*(rawsize-len(res)))
# add section and adjust headers
s=section_table+40*num
struct.pack_into('<8sIIIIIIHHI',data,s,b'.rsrc\0\0\0',len(res),new_rva,rawsize,rawptr,0,0,0,0,0x40000040)
struct.pack_into('<H',data,coff+2,num+1)
struct.pack_into('<I',data,opt+56,align(new_rva+len(res),section_align))
init=struct.unpack_from('<I',data,opt+8)[0]
struct.pack_into('<I',data,opt+8,init+rawsize)
struct.pack_into('<II',data,opt+112+8*2,new_rva,len(res))
assert rawptr>=len(data)
data.extend(b'\0'*(rawptr-len(data)))
data.extend(res)
exe.write_bytes(data)
print('Embedded',count,'Windows icon sizes into PE .rsrc; Rsrc',hex(new_rva),'file',len(data))
