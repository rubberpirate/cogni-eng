"""Author the v1 contract; regenerate TypeScript with bun run generate."""
import json
from pathlib import Path
S={'type':'string'}; I={'type':'integer'}; B={'type':'boolean'}
def obj(fields,optional=()):return {'type':'object','additionalProperties':False,'properties':fields,'required':[x for x in fields if x not in optional]}
def ref(name):return {'$ref':'#/components/schemas/'+name}
def arr(name):return {'type':'array','items':ref(name)}
schemas={
 'Error':obj({'code':S,'message':S,'request_id':S}),
 'Outcome':obj({'state':S,'id':S,'reference':S,'message':S,'delivery':S},('id','reference','message','delivery','state')),
 'Me':obj({'id':S,'name':S,'email':S,'csrf':S,'mfa_enabled':B,'session_id':S,'environment':S}),
 'Workspace':obj({'id':S,'name':S,'timezone':S,'locale':S,'role':S,'campus_id':S}),
 'Campus':obj({'id':S,'name':S,'code':S,'address':S,'state':S,'version':I}),
 'Person':obj({'membership_id':S,'id':S,'name':S,'email':S,'role':S,'campus_id':S,'state':S,'access_state':S,'version':I}),
 'Course':obj({'id':S,'campus_id':S,'name':S,'academic_year':S}),
 'Enrollment':obj({'id':S,'person_id':S,'name':S,'campus_id':S,'course_id':S,'course':S,'reference':S,'state':S,'version':I}),
 'Family':obj({'id':S,'guardian_id':S,'guardian':S,'learner_id':S,'learner':S,'relationship':S,'state':S}),
 'Responsibility':obj({'id':S,'staff_id':S,'staff':S,'course_id':S,'course':S,'subject':S,'starts_on':S,'ends_on':S}),
 'Audit':obj({'id':I,'actor':S,'action':S,'resource_id':S,'reason':S,'request_id':S,'time':S}),
 'Session':obj({'id':S,'device':S,'created_at':S,'last_seen':S,'expires_at':S,'current':B}),
 'Delivery':obj({'id':S,'recipient':S,'kind':S,'state':S,'attempts':I,'error':S,'created_at':S}),
 'SignIn':obj({'email':S,'password':S,'code':S,'remember':B},('code','remember')),
 'Reset':obj({'token':S,'password':S,'code':S,'accept_policy':B},('code','accept_policy')),
 'Invite':obj({'name':S,'email':S,'role':{'type':'string','enum':['coordinator','instructor','learner','guardian','reviewer']},'campus_id':S}),
 'NewInstitution':obj({'name':S,'timezone':S,'locale':{'type':'string','enum':['en','hi']}}),
 'NewCampus':obj({'name':S,'code':S,'address':S}),
 'NewCourse':obj({'name':S,'campus_id':S,'academic_year':S}),
 'NewEnrollment':obj({'person_id':S,'campus_id':S,'course_id':S,'start_date':S}),
 'NewFamily':obj({'guardian_id':S,'learner_id':S,'campus_id':S,'relationship':S}),
 'NewResponsibility':obj({'staff_id':S,'course_id':S,'campus_id':S,'subject':S,'starts_on':S,'ends_on':S}),
 'MembershipChange':obj({'state':{'type':'string','enum':['active','suspended','archived']},'reason':S,'version':I}),
 'Factor':obj({'code':S}), 'MFAEnrollment':obj({'secret':S,'url':S}),
 'RecoveryCodes':obj({'recovery_codes':{'type':'array','items':S}}), 'RecoveryRequest':obj({'email':S})
}
schemas.update({
 'Profile':obj({'name':S,'email':S,'phone':S,'emergency_contact':S}),
 'ProfileChange':obj({'name':S,'phone':S,'emergency_contact':S}),
 'PasswordChange':obj({'current_password':S,'password':S,'code':S}),
 'EmailChange':obj({'email':S,'password':S,'code':S}),
 'Token':obj({'token':S}),
 'AccountEvent':obj({'id':I,'action':S,'reason':S,'time':S}),
 'Settings':obj({'name':S,'timezone':S,'locale':{'type':'string','enum':['en','hi']},'state':S,'support_email':S,'academic_year':S,'require_staff_mfa':B,'version':I}),
 'OwnerTransfer':obj({'person_id':S,'password':S,'code':S,'reason':S}),
 'PlacementChange':obj({'campus_id':S,'course_id':S,'state':{'type':'string','enum':['active','withdrawn']},'effective_on':S,'reason':S,'version':I}),
})
paths={}
def route(path,method,response='Outcome',body=None,public=False,listed=False):
 path='/api/v1'+path
 op={'operationId':method+'_'+path.replace('/','_').replace('{','').replace('}',''), 'summary':method.upper()+' '+path,'security':[] if public else [{'session':[]}], 'responses':{}}
 for code in ('400','401','403','404','409','429','500'):
  op['responses'][code]={'description':'Structured error','content':{'application/json':{'schema':ref('Error')}}}
 op['responses']['200']={'description':'Successful committed operation','content':{'application/json':{'schema':arr(response) if listed else ref(response)}}}
 params=[]
 for p in ('org','id'):
  if '{'+p+'}' in path:params.append({'name':p,'in':'path','required':True,'schema':S})
 if listed and '{org}' in path:
  params += [{'name':p,'in':'query','schema':S,'description':'Bounded results; 200 maximum. Empty campus requires institution-wide entitlement.'} for p in ('campus','q')]
 if method!='get' and not public:params.append({'name':'X-CSRF-Token','in':'header','required':True,'schema':S})
 if params:op['parameters']=params
 if body:op['requestBody']={'required':True,'content':{'application/json':{'schema':ref(body)}}}
 paths.setdefault(path,{})[method]=op
route('/sign-in','post',body='SignIn',public=True);route('/sign-out','post');route('/me','get','Me')
route('/sessions','get','Session',listed=True);route('/sessions/{id}','delete')
route('/recovery','post',body='RecoveryRequest',public=True);route('/reset','post',body='Reset',public=True);route('/accept-invitation','post',body='Reset',public=True)
route('/mfa/enroll','post','MFAEnrollment');route('/mfa/confirm','post','RecoveryCodes','Factor')
route('/workspaces','get','Workspace',listed=True);route('/institutions','post',body='NewInstitution')
for plural,entity,new in [('campuses','Campus','NewCampus'),('people','Person',None),('courses','Course','NewCourse'),('enrollments','Enrollment','NewEnrollment'),('families','Family','NewFamily'),('responsibilities','Responsibility','NewResponsibility'),('audit','Audit',None),('deliveries','Delivery',None)]:
 route('/institutions/{org}/'+plural,'get',entity,listed=True)
 if new:route('/institutions/{org}/'+plural,'post',body=new)
route('/institutions/{org}/invitations','post',body='Invite');route('/institutions/{org}/memberships/{id}','patch',body='MembershipChange');route('/institutions/{org}/families/{id}','delete');route('/institutions/{org}/deliveries/{id}/retry','post')
route('/profile','get','Profile');route('/profile','patch',body='ProfileChange')
route('/password','post',body='PasswordChange');route('/email-change','post',body='EmailChange');route('/email-change/confirm','post',body='Token')
route('/account-activity','get','AccountEvent',listed=True)
route('/institutions/{org}/settings','get','Settings');route('/institutions/{org}/settings','patch',body='Settings')
route('/institutions/{org}/campuses/{id}','patch',body='Campus')
route('/institutions/{org}/owner-transfer','post',body='OwnerTransfer')
route('/institutions/{org}/enrollments/{id}','patch',body='PlacementChange')
doc={'openapi':'3.1.0','info':{'title':'Learning Workspace API','version':'1.0.0','description':'Go-owned business API. Mutations require exact configured Origin, opaque cookie and CSRF header. No private response may be publicly cached. Limits and incomplete functionality are tracked in IMPLEMENTATION-STATUS.md.'},'servers':[{'url':'http://localhost:3000'}],'paths':paths,'components':{'securitySchemes':{'session':{'type':'apiKey','in':'cookie','name':'workspace_session'}},'schemas':schemas}}
Path('contracts/openapi.json').write_text(json.dumps(doc,indent=2)+'\n')
