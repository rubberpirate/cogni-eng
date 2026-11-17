# /// script
# requires-python = ">=3.9"
# dependencies = ["reportlab==5.1.0"]
# ///
import json
from pathlib import Path
base={'marks':'2','penalty':'0.5','difficulty':'standard','language':'en','provenance':'Original Learning Workspace practice sample','license':'CC0-1.0','rubric':[]}
items=[('single','Seven seeds are joined by five more seeds. How many seeds are there?',['10','12','14','15'],['B'],'Seven plus five is twelve: 7 + 5 = 12.'),('single','What is half of 18?',['6','7','9','12'],['C'],'Two groups of nine make eighteen, so half of eighteen is nine.'),('multiple','Select every number divisible by 3.',['12','14','21','25'],['A','C'],'12 = 3 x 4 and 21 = 3 x 7. The other numbers leave a remainder.'),('single','A cart travels 24 metres in 6 seconds at constant speed. What is its speed?',['3 m/s','4 m/s','6 m/s','12 m/s'],['B'],'Speed = distance / time = 24 / 6 = 4 metres per second.'),('single','How many oxygen atoms are represented by one CO2 molecule?',['1','2','3','4'],['B'],'The subscript 2 applies to oxygen: a CO2 molecule has two oxygen atoms.'),('single','Which structure controls substances entering and leaving an animal cell?',['Cell wall','Cell membrane','Nucleus','Chloroplast'],['B'],'The selectively permeable cell membrane regulates exchange with the surroundings.')]
qs=[dict(base,type=t,stem=stem,options=[{'id':chr(65+i),'text':v} for i,v in enumerate(opts)],key=key,solution=sol) for t,stem,opts,key,sol in items]
Path('internal/academic/fixtures/questions.json').write_text(json.dumps(qs,indent=2)+'\n')
Path('apps/web/public/samples/questions.json').write_text(json.dumps(qs,indent=2)+'\n')
import csv
with open('apps/web/public/samples/questions.csv','w',newline='') as f:
 w=csv.writer(f);w.writerow(['type','question','A','B','C','D','answer','marks','penalty','solution','provenance','license'])
 for q in qs:w.writerow([q['type'],q['stem'],*[o['text'] for o in q['options']],'|'.join(q['key']),'2','0.5',q['solution'],q['provenance'],q['license']])
Path("internal/academic/fixtures/questions.csv").write_bytes(Path("apps/web/public/samples/questions.csv").read_bytes())
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4
from reportlab.lib.colors import HexColor
from reportlab.lib.utils import simpleSplit
for key in (False,True):
 name='marking-guide.pdf' if key else 'question-paper.pdf'
 c=canvas.Canvas('apps/web/public/samples/'+name,pagesize=A4,invariant=1);W,H=A4
 c.setTitle('Practice sample - '+('marking guide' if key else 'question paper'));c.setAuthor('Learning Workspace - original CC0 content')
 for page in range(2):
  c.setFillColor(HexColor('#173e31'));c.rect(0,H-110,W,110,fill=1,stroke=0);c.setFillColor(HexColor('#ffffff'));c.setFont('Helvetica-Bold',23);c.drawString(42,H-48,'Practice, scan, review.');c.setFont('Helvetica',11);c.drawString(42,H-72,'Six original questions | '+('Teacher answer key' if key else 'Question paper'))
  c.setFillColor(HexColor('#20382d'));c.setFont('Helvetica',10);c.drawString(42,H-136,'12 marks total | +2 correct | -0.5 wrong | 0 blank | 15 minutes')
  c.setFont('Helvetica',10);c.drawString(42,H-154,'Use a personalized OMR sheet from Try sample files in the app.')
  y=H-194
  for i in range(page*3,page*3+3):
   q=qs[i];c.setFont('Helvetica-Bold',11)
   for line in simpleSplit(str(i+1)+'. '+q['stem'],'Helvetica-Bold',11,W-92):c.drawString(42,y,line);y-=16
   c.setFont('Helvetica',10)
   if key:
    c.drawString(56,y,'Answer: '+' + '.join(q['key']));y-=18
    for line in simpleSplit(q['solution'],'Helvetica',10,W-108):c.drawString(56,y,line);y-=15
   else:
    for o in q['options']:c.drawString(56,y,o['id']+'. '+o['text']);y-=17
    if q['type']=='multiple':c.setFont('Helvetica-Oblique',9);c.drawString(56,y,'Choose all correct options.');y-=14
   y-=28
  c.setStrokeColor(HexColor('#b7c9bf'));c.line(42,58,W-42,58);c.setFont('Helvetica',9);c.drawString(42,42,'Original CC0-1.0 demonstration material. For practice, not formal examinations.');c.drawRightString(W-42,26,str(page+1)+' / 2');c.showPage()
 c.save()
 Path("internal/academic/fixtures/"+name).write_bytes(Path("apps/web/public/samples/"+name).read_bytes())
