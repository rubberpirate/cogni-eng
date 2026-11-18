"""Academic API schemas and operations, loaded by contract.py. Go tagged structs are the DTO source."""
import re
names={}
for source in ['internal/academic/model.go','internal/academic/import.go','internal/academic/omr.go','internal/academic/grid.go','internal/academic/key.go','internal/scholar/service.go','internal/scholar/import.go','internal/scholar/paper.go','internal/scholar/sitting.go','internal/scholar/scan.go','internal/scholar/jobs.go','internal/scholar/tutor.go','internal/scholar/evidence.go','internal/scholar/samples.go']:
 for name,body in re.findall(r'type (\w+) struct \{\n(.*?)\n\}',Path(source).read_text(),re.S):
  names[name]=body

def dtype(t):
 t=t.removeprefix('academic.')
 if t.startswith('*'):return dtype(t[1:])
 if t.startswith('[]'):return {'type':'array','items':dtype(t[2:])}
 if t.startswith('map[string]'):return {'type':'object','additionalProperties':dtype(t[11:])}
 if t in ('string','time.Time'):return S
 if t in ('int','int64'):return I
 if t=='float64':return {'type':'number'}
 if t=='bool':return B
 if t=='json.RawMessage':return {}
 return ref('Academic'+t)
for name,body in names.items():
 fields={};optional=[]
 for typ,tag in re.findall(r'^\s*\w+\s+([^`\n]+?)\s+`json:"([^"]+)"`',body,re.M):
  key=tag.split(',')[0]
  if key=='-':continue
  fields[key]=dtype(typ.strip())
  if 'omitempty' in tag or typ.startswith('*'):optional.append(key)
 schemas['Academic'+name]=obj(fields,optional)
# Types with only one-line Go declarations are explicitly described here.
schemas['AcademicGeneration']=obj({'topic':ref('AcademicTopic'),'count':I,'difficulty':S})
schemas.update({
 'AcademicVersion':obj({'version':I}),
 'AcademicQuestionWrite':obj({'question':ref('AcademicQuestion'),'version':I},('version',)),
 'AcademicQuestionCreated':obj({'question':ref('AcademicQuestion'),'duplicate':B}),
 'AcademicQuestionReview':obj({'state':S,'note':S,'version':I}),
 'AcademicImportChange':obj({'index':I,'question':ref('AcademicQuestion'),'version':I,'selected':{'type':'array','items':I}},('index','question','selected')),
 'AcademicAssign':obj({'paper_id':S,'learners':{'type':'array','items':S},'opens_at':S,'closes_at':S,'extra_minutes':I}),
 'AcademicStart':obj({'takeover':B}),
 'AcademicSave':obj({'operation_id':S,'version':I,'answers':{'type':'object','additionalProperties':ref('AcademicAnswer')}}),
 'AcademicReviewMarks':obj({'reason':S,'revision':I,'manual_marks':{'type':'object','additionalProperties':ref('AcademicManualMark')}}),
 'AcademicNewLayout':obj({'paper_id':S,'columns':I}),
 'AcademicNewBatch':obj({'paper_id':S,'layout_id':S,'sitting_at':S}),
 'AcademicScanReview':obj({'assignment_id':S,'reason':S,'identity_confirmed':B,'version':I,'corrections':{'type':'object','additionalProperties':ref('AcademicAnswer')},'commit':B}),
 'AcademicGenerate':obj({'topic_id':S,'count':I,'difficulty':S}),
 'AcademicStrings':{'type':'array','items':S},
 'AcademicEmpty':obj({}),
 'AcademicSaved':obj({'saved':B,'published':B,'released':B,'sitting_id':S},('saved','published','released','sitting_id')),
})
base='/institutions/{org}/education/'
for path,entity,query in [('courses','Course','campus'),('topics','Topic','campus'),('questions','Question','topic'),('imports','Import','topic'),('papers','PaperRecord','campus'),('assignments','Assigned','paper'),('layouts','SheetLayout','paper'),('batches','Batch','paper'),('scan-sheets','ScanSheet','batch'),('jobs','Job','campus')]:
 route(base+path,'get','Academic'+entity,listed=True)
 op=paths['/api/v1'+base+path]['get'];op['parameters']=[p for p in op.get('parameters',[]) if p['in']!='query']+[{'name':query,'in':'query','schema':S}]
for path,entity in [('imports','Import'),('papers','PaperRecord'),('sittings','Sitting'),('jobs','Job')]:route(base+path+'/{id}','get','Academic'+entity)
route(base+'recipients/{id}','get','AcademicRecipient',listed=True)
for path,response,body in [
 ('topics','Topic','Topic'),('questions','QuestionCreated','QuestionWrite'),('questions/{id}','Question','QuestionWrite'),('questions/{id}/review','Saved','QuestionReview'),
 ('imports/{id}/edit','Import','ImportChange'),('imports/{id}/commit','Import','ImportChange'),('papers','PaperRecord','Compose'),('papers/{id}/publish','Saved','Version'),
 ('blueprint','Strings','Generate'),('assignments','Strings','Assign'),('assignments/{id}/start','Sitting','Start'),('sittings/{id}/save','Sitting','Save'),('sittings/{id}/submit','Sitting','Empty'),('sittings/{id}/release','Saved','Empty'),('sittings/{id}/review','Sitting','ReviewMarks'),
 ('layouts','SheetLayout','NewLayout'),('batches','Batch','NewBatch'),('scan-sheets/{id}/review','Saved','ScanReview'),('generate','Job','Generate'),('jobs/{id}/retry','Saved','Empty'),('jobs/{id}/cancel','Saved','Empty')]:route(base+path,'post','Academic'+response,'Academic'+body)
for path,entity,fields in [('imports/{id}/upload','Import',{'file':{'type':'string','format':'binary'},'topic_id':S,'provenance':S,'license':S}),('batches/{id}/upload','Job',{'file':{'type':'string','format':'binary'}}),('assets','Asset',{'file':{'type':'string','format':'binary'},'campus_id':S,'course_id':S,'purpose':S})]:
 route(base+path,'post','Academic'+entity)
 paths['/api/v1'+base+path]['post']['requestBody']={'required':True,'content':{'multipart/form-data':{'schema':obj(fields)}}}
for path in ['assets/{id}','sheet-image/{id}']:
 route(base+path,'get')
 op=paths['/api/v1'+base+path]['get'];op['responses']['200']={'description':'Authorized private binary','content':{'image/png':{'schema':{'type':'string','format':'binary'}},'application/pdf':{'schema':{'type':'string','format':'binary'}}}}
 if path.startswith('sheet'):op['parameters'].append({'name':'assignment','in':'query','required':True,'schema':S})

schemas['AcademicParseDocument']=obj({'text':S,'provenance':S,'license':S,'version':I})
route(base+'imports/{id}/parse','post','AcademicImport','AcademicParseDocument')
schemas['AcademicTutorAsk']=obj({'sitting_id':S,'question_id':S,'text':S})
route(base+'tutor','post','AcademicThread','AcademicTutorAsk')
route(base+'tutor','get','AcademicThread')
paths['/api/v1'+base+'tutor']['get']['parameters'] += [{'name':p,'in':'query','required':True,'schema':S} for p in ['sitting','question']]

route(base+'evidence/{id}','get','AcademicTopicEvidence',listed=True)
route(base+'result-history/{id}','get','AcademicResultHistory',listed=True)

route(base+'print-paper/{id}','get')
paths['/api/v1'+base+'print-paper/{id}']['get']['parameters'].append({'name':'answers','in':'query','schema':{'type':'string','enum':['0','1']}})
paths['/api/v1'+base+'print-paper/{id}']['get']['responses']['200']={'description':'Authorized printable booklet or approved key','content':{'text/html':{'schema':S}}}

schemas['AcademicSampleRequest']=obj({'course_id':S,'learner_id':S})
route(base+'sample-kits','post','AcademicSampleKit','AcademicSampleRequest')
route(base+'course-learners/{id}','get','AcademicRecipient',listed=True)
route(base+'sample-files/{id}','get')
paths['/api/v1'+base+'sample-files/{id}']['get']['parameters'].append({'name':'assignment','in':'query','required':True,'schema':S})
paths['/api/v1'+base+'sample-files/{id}']['get']['responses']['200']={'description':'Private original practice test files with assigned registered sheets','content':{'application/zip':{'schema':{'type':'string','format':'binary'}}}}

schemas['AcademicApplyKey']=obj({'job_id':S,'version':I})
route(base+'imports/{id}/apply-key','post','AcademicImport','AcademicApplyKey')
for path,response,fields in [('imports/{id}/key-upload','Job',{'file':{'type':'string','format':'binary'}}),('grid-preview','GridReading',{'file':{'type':'string','format':'binary'},'course_id':S,'calibration':S})]:
 route(base+path,'post','Academic'+response)
 paths['/api/v1'+base+path]['post']['requestBody']={'required':True,'content':{'multipart/form-data':{'schema':obj(fields)}}}
