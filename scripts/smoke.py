"""Real process smoke test: four API/RPC pairs, Redis, isolated databases and scheduler."""
from pathlib import Path
import subprocess,sys,os,json,socket,time,urllib.request,urllib.error,uuid,sqlite3,hmac,hashlib,datetime,traceback
root=Path(__file__).resolve().parents[1]
sys.stdout.reconfigure(encoding='utf-8')
if '--skip-build' not in sys.argv:subprocess.run([sys.executable,str(root/'scripts/build-smoke.py')],cwd=root,check=True)
binary=root/'.tmp/smoke/bin';run=root/'.tmp/smoke/runs'/uuid.uuid4().hex;run.mkdir(parents=True)
governance=None
processes=[];logs=[];api_configs=[];env=os.environ.copy();key='native-smoke-service-key-32-bytes-long';secret='native-smoke-auth-secret-32-bytes-long';ports={};checks=[]
def port():
    with socket.socket() as s:s.bind(('127.0.0.1',0));return s.getsockname()[1]
for d in ['rbac','blog','product','file']:
    for kind in ['api','rpc']:ports[d+'-'+kind]=port()
def config(name,data):
    if governance is not None:data=governance.configure(name,data)
    p=run/(name+'.json');p.write_text(json.dumps(data),encoding='utf-8');return p
def spawn(name,cfg=None):
    log=open(run/(name+'.log'),'wb');logs.append(log)
    p=subprocess.Popen([str(binary/(name+('.exe' if os.name=='nt' else '')))]+(['-f',str(cfg)] if cfg else []),cwd=root,env=env,stdout=log,stderr=log,creationflags=getattr(subprocess,'CREATE_NO_WINDOW',0));processes.append(p);return p
def request(d,path,body=None,token=None,method=None,headers=None,expect=200,raw=False):
    url=f'http://127.0.0.1:{ports[d+"-api"]}'+path;data=None;h=dict(headers or {})
    if body is not None:
        data=body if isinstance(body,bytes) else json.dumps(body).encode();h.setdefault('Content-Type','application/json')
    if token:h['Authorization']='Bearer '+token
    req=urllib.request.Request(url,data=data,headers=h,method=method or ('POST' if data is not None else 'GET'))
    try:
        with urllib.request.urlopen(req,timeout=15) as response:status=response.status;payload=response.read();response_headers=dict(response.headers)
    except urllib.error.HTTPError as e:status=e.code;payload=e.read();response_headers=dict(e.headers)
    if status!=expect:raise AssertionError(f'{d} {path}: expected {expect}, got {status}: {payload[:1000]!r}')
    return (payload,response_headers) if raw else (json.loads(payload).get('data') if payload and status<400 else None)
def check(name):checks.append(name);print('PASS',name,flush=True)
def upload(d,path,name,data,token,kind=''):
    boundary='native'+uuid.uuid4().hex
    body=(f'--{boundary}\r\nContent-Disposition: form-data; name="kind"\r\n\r\n{kind}\r\n--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{name}"\r\nContent-Type: application/octet-stream\r\n\r\n').encode()+data+f'\r\n--{boundary}--\r\n'.encode()
    return request(d,path,body,token,headers={'Content-Type':'multipart/form-data; boundary='+boundary})
try:
    if '--governance' in sys.argv or '--frontends' in sys.argv:
        from governance_smoke import Harness
        governance=Harness(root,run,ports)
        governance.start()
    redis=subprocess.Popen([str(binary/('redis.exe' if os.name=='nt' else 'redis'))],stdout=subprocess.PIPE,stderr=subprocess.PIPE,creationflags=getattr(subprocess,'CREATE_NO_WINDOW',0));processes.append(redis);redisaddr=redis.stdout.readline().decode().strip()
    def redis_cmd(*args):
        host,p=redisaddr.split(':')
        with socket.create_connection((host,int(p)),timeout=5) as sock:
            sock.sendall(('*'+str(len(args))+'\r\n').encode()+b''.join(('$'+str(len(a.encode()))+'\r\n'+a+'\r\n').encode() for a in args));return sock.recv(1024)
    for d in ['rbac','file','blog','product']:
        db=run/(d+'.db');env[d.upper()+'_DB_DSN']=str(db)
        env.update(BOOTSTRAP_ADMIN_USERNAME='nativeadmin',BOOTSTRAP_ADMIN_PASSWORD='NativePassword123!',BOOTSTRAP_ADMIN_EMAIL='admin@example.test')
        result=subprocess.run([str(binary/('bootstrap.exe' if os.name=='nt' else 'bootstrap')),'-service',d,'-driver','sqlite']+(['-admin'] if d=='rbac' else []),cwd=root,env=env,capture_output=True)
        (run/(d+'-bootstrap.log')).write_bytes(result.stdout+result.stderr)
        if result.returncode:raise RuntimeError(d+' schema failed; see '+str(run/(d+'-bootstrap.log')))
        rpc={'Name':d+'-rpc','ListenOn':f'127.0.0.1:{ports[d+"-rpc"]}','Mode':'test','Timeout':30000,'ServiceKey':key,'Database':{'Driver':'sqlite','DSN':str(db),'MaxOpen':1,'MaxIdle':1},'Log':{'Mode':'console','Level':'error'}}
        downstream=lambda name:{'Endpoints':[f'127.0.0.1:{ports[name+"-rpc"]}'],'NonBlock':True,'Timeout':30000}
        if d=='rbac':rpc.update(Cache={'Addr':redisaddr},AuthSecret=secret,TOTP={},Mail={})
        else:rpc['RbacRPC']=downstream('rbac')
        if d in ['blog','product']:rpc['FileRPC']=downstream('file')
        if d=='file':rpc.update(StorageRoot=str(run/'files'),StorageSecret=secret)
        rpc_process=spawn(d+'-rpc',config(d+'-rpc',rpc))
        for attempt in range(300):
            if rpc_process.poll() is not None:raise RuntimeError(d+' RPC startup failed; see '+str(run/(d+'-rpc.log')))
            try:
                with socket.create_connection(('127.0.0.1',ports[d+'-rpc']),timeout=.2):break
            except OSError:time.sleep(.1)
        api={'Name':d+'-api','Host':'127.0.0.1','Port':ports[d+'-api'],'Mode':'test','Timeout':30000,'ServiceKey':key,'RPC':downstream(d),'Log':{'Mode':'console','Level':'error'}}
        if d=='file':api['ProductRPC']=downstream('product')
        api_configs.append((d+'-api',config(d+'-api',api)))
    for name,cfg in api_configs:spawn(name,cfg)
    for d in ['rbac','blog','product','file']:
        for attempt in range(80):
            try:request(d,'/health');break
            except Exception:
                if attempt==79:raise
                time.sleep(.15)
    check('all four native API -> RPC health calls')
    request('blog','/v1/blog/admin/articles',expect=401)
    request('rbac','/v1/rbac/users/detail?id=invalid',expect=400)
    check('authentication and HTTP parameter error classification')
    assert redis_cmd('SET','sysCaptcha:admin@example.test','123456','EX','300').startswith(b'+OK')
    login=request('rbac','/v1/rbac/auth/login',dict(username='nativeadmin',password='NativePassword123!',email='admin@example.test',code='123456'));admin=login['token']
    current=request('rbac','/v1/rbac/auth/me',token=admin);assert current['username']=='nativeadmin' and 'password' not in current
    check('administrator login and secret-free profile')
    once=request('rbac','/v1/rbac/auth/once-token',{},admin)['token']
    request('rbac','/v1/rbac/auth/once-token/consume',dict(token=once),admin)
    request('rbac','/v1/rbac/auth/once-token/consume',dict(token=once),admin,expect=401)
    check('one-time credentials cannot be replayed')
    request('rbac','/v1/rbac/blacklist/add',dict(ip='203.0.113.7',reason='native smoke'),admin)
    # In the governance topology the gateway and test client share loopback;
    # exercise header stripping at the real gateway trust boundary. The basic
    # smoke topology still verifies untrusted direct clients independently.
    if governance is not None:
        governance.setup_routes()
        assert governance.request('/v1/blog/articles',headers={'X-Real-IP':'203.0.113.7','X-Forwarded-For':'203.0.113.7'})[0]==200
    else:
        request('blog','/v1/blog/articles',headers={'X-Real-IP':'203.0.113.7'})
    check('untrusted forwarded IP headers are ignored')
    request('rbac','/v1/rbac/users/create',dict(username='',password='x',email='invalid',status=1),admin,expect=400)
    request('rbac','/v1/rbac/users/create',dict(username='operator',nickname='Operator',password='OperatorPassword123',email='operator@example.test',status=1,gender=0),admin)
    request('rbac','/v1/rbac/roles/create',dict(role_code='native-reader',role_name='Reader',status=1),admin)
    request('rbac','/v1/rbac/menus/create',dict(name='User list',code='system:user:list',type=3,status=1),admin)
    with sqlite3.connect(run/'rbac.db') as db:
        user_id=db.execute('SELECT id FROM sys_user WHERE username=?',('operator',)).fetchone()[0]
        role_id=db.execute('SELECT id FROM sys_role WHERE role_code=?',('native-reader',)).fetchone()[0]
        menu_id=db.execute('SELECT id FROM sys_menu WHERE menu_code=?',('system:user:list',)).fetchone()[0]
    request('rbac','/v1/rbac/roles/menus',dict(role_id=role_id,menu_ids=[menu_id]),admin)
    request('rbac','/v1/rbac/users/roles',dict(userId=user_id,roleIds=[role_id]),admin)
    redis_cmd('SET','sysCaptcha:operator@example.test','123456','EX','300')
    operator=request('rbac','/v1/rbac/auth/login',dict(username='operator',password='OperatorPassword123',email='operator@example.test',code='123456'))['token']
    request('rbac','/v1/rbac/users',token=operator)
    request('rbac','/v1/rbac/roles',token=operator,expect=403)
    request('rbac','/v1/rbac/users/roles',dict(userId=user_id,roleIds=[]),admin)
    request('rbac','/v1/rbac/users',token=operator,expect=403)
    check('RBAC creation, assignment, permission denial and live revocation')
    request('blog','/v1/blog/categories/create',dict(name='Native category',slug='native-category'),admin)
    category=int(request('blog','/v1/blog/categories?page=1&pageSize=20')['items'][0]['id'])
    request('blog','/v1/blog/tags/create',dict(name='Native',slug='native'),admin)
    tag=int(request('blog','/v1/blog/tags?page=1&pageSize=20')['items'][0]['id'])
    asset=upload('blog','/v1/blog/assets/upload','notes.txt',b'native private asset',admin)
    request('blog','/v1/blog/assets?id='+asset['id'],expect=404)
    raw,_=request('blog','/v1/blog/assets?id='+asset['id'],token=admin,raw=True);assert raw==b'native private asset'
    check('private article asset streams across blog and file')
    draft_input=dict(title='Native migration',slug='native-migration',content='# Native\n\n[asset]('+asset['url']+')',content_format='markdown',category_id=category,tag_ids=[tag])
    draft=request('blog','/v1/blog/admin/articles/draft',dict(requestID=uuid.uuid4().hex,input=draft_input),admin)
    article=int(draft['id']);version=int(draft['version'])
    request('blog',f'/v1/blog/articles/detail?id={article}',expect=404)
    request('blog','/v1/blog/admin/articles/action',dict(id=article,version=version+1,requestID=uuid.uuid4().hex,action='publish'),admin,expect=409)
    published=request('blog','/v1/blog/admin/articles/action',dict(id=article,version=version,requestID=uuid.uuid4().hex,action='publish'),admin)
    assert request('blog',f'/v1/blog/articles/detail?id={article}')['title']=='Native migration'
    request('blog','/v1/blog/tags/delete',dict(id=tag),admin,expect=409)
    check('article draft/publish, optimistic locking and reference integrity')
    edited=dict(draft_input);edited['title']='Unpublished edit';edited['content']='# Unpublished edit\n\n[asset]('+asset['url']+')'
    update=request('blog','/v1/blog/admin/articles/draft',dict(id=article,version=int(published['version']),requestID=uuid.uuid4().hex,input=edited),admin)
    assert request('blog',f'/v1/blog/articles/detail?id={article}')['title']=='Native migration'
    assert request('blog','/v1/blog/articles/search?keyword=Unpublished&page=1&pageSize=20')['items']==[]
    assert request('blog','/v1/blog/articles/hot?limit=5')['items'][0]['title']=='Native migration'
    check('unpublished edits never appear in public detail, search or hot lists')
    body,_=request('blog','/v1/blog/assets?id='+asset['id'],raw=True);assert body==b'native private asset'
    check('published asset becomes public through blog authorization')
    data=b'0123456789-native-release'
    release=upload('file','/v1/file/upload','native.tgz',data,admin,'release')
    release_admin='/v1/product/admin/releases'
    request('product',release_admin+'/groups/create',dict(id='native-tools',name='Native tools'),admin)
    package=request('product',release_admin+'/files/create',dict(id='native',groupId='native-tools',platform='Linux',product='Deploy',version='v1',url=release['url'],fileName='native.tgz'),admin)
    request('product',release_admin+'/files/publish',dict(id='native',revision=int(package['revision'])),admin)
    request('file','/v1/file/download?id=native',expect=401)
    request('file',release['url'],expect=404)
    email='member@example.test';code='654321';digest=hmac.new(secret.encode(),('product-register:'+email+':'+code).encode(),hashlib.sha256).hexdigest()
    with sqlite3.connect(run/'rbac.db') as db:db.execute('INSERT INTO product_member_verifications(email,code_hash,expires_at,attempts) VALUES (?,?,?,0)',(email,digest,(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(minutes=5)).isoformat()))
    request('rbac','/v1/rbac/members/register',dict(email=email,password='MemberPassword123',code=code))
    member=request('rbac','/v1/rbac/members/login',dict(email=email,password='MemberPassword123'))['token']
    body,headers=request('file','/v1/file/download?id=native',token=member,headers={'Range':'bytes=2-5'},expect=206,raw=True);assert body==data[2:6]
    body,headers=request('file','/v1/file/download?id=native',token=member,method='HEAD',raw=True);assert body==b'' and int(headers['Content-Length'])==len(data)
    assert 'native.tgz' in headers['Content-Disposition']
    request('rbac','/v1/rbac/users',token=member,expect=401)
    check('member-only downloads, Range/HEAD and admin/member separation')
    manuals='/v1/product/admin/manuals'
    request('product',manuals+'/groups/save',dict(version='v1',title='Native group'),admin)
    workspace=request('product',manuals+'/workspace?version=v1',token=admin)
    group=next(g for g in workspace['groups'] if g['title']=='Native group')
    document=dict(version='v1',id='native-doc',slug='native-document-slug',title='Native document',summary='summary',content='# Native documentation',groupId=int(group['id']),revision=0)
    saved=request('product',manuals+'/documents/save',document,admin)
    public='/v1/product/manuals/document?id=native-doc&version=v1'
    request('product',public,expect=404)
    request('product',manuals+'/documents/publish',dict(id='native-doc',version='v1',revision=int(saved['revision'])),admin)
    assert request('product',public)['content']=='# Native documentation'
    request('product',manuals+'/documents/save',document,admin,expect=409)
    history=request('product',manuals+'/history?id=native-doc&version=v1',token=admin)
    assert len(history)>0
    request('product',manuals+'/groups/delete',dict(version='v1',id=int(group['id']),revision=int(group['revision'])),admin,expect=409)
    check('product documents, publication, revision history, edit conflicts and group references')
    scheduled_input=dict(draft_input);scheduled_input['slug']='scheduled-native';scheduled_input['title']='Scheduled native'
    future=request('blog','/v1/blog/admin/articles/draft',dict(requestID=uuid.uuid4().hex,input=scheduled_input),admin)
    due=(datetime.datetime.now(datetime.timezone.utc)+datetime.timedelta(minutes=2)).isoformat()
    scheduled=request('blog','/v1/blog/admin/articles/action',dict(id=int(future['id']),version=int(future['version']),requestID=uuid.uuid4().hex,action='schedule',scheduledAt=due),admin)
    # Advance only the test fixture's due time; production scheduling still enforces its minimum delay.
    with sqlite3.connect(run/'blog.db') as db:db.execute('UPDATE articles SET scheduled_at=? WHERE id=?',((datetime.datetime.now(datetime.timezone.utc)-datetime.timedelta(seconds=2)).isoformat(),int(future['id'])))
    worker=spawn('scheduler',config('scheduler',dict(Name='scheduler-test',Mode='test',ServiceKey=key,BlogRPC=downstream('blog'),PublishSchedule='@every 1s',CleanupSchedule='@every 1h',JobTimeout='5s')))
    for attempt in range(40):
        if worker.poll() is not None:raise AssertionError('scheduler exited before receiving a shutdown signal')
        try:
            visible=request('blog','/v1/blog/articles/detail?id='+future['id']);assert visible['title']=='Scheduled native';break
        except AssertionError:
            if attempt==39:raise
            time.sleep(.2)
    check('standalone scheduler publishes through blog RPC and remains running')
    request('rbac','/v1/rbac/auth/logout',{},admin)
    request('rbac','/v1/rbac/auth/me',token=admin,expect=401)
    check('revoked sessions fail immediately')
    if governance is not None:
        governance.verify(member)
        check('real Go gateway/Nacos registration, atomic configuration reload, publish and rollback')
    if '--frontends' in sys.argv:
        redis_cmd('SET','sysCaptcha:admin@example.test','654321','EX','300')
        web_env=env.copy()
        web_env.update(governance.browser_env())
        web_env.update(FRONTEND_ARTICLE_ID=str(article),FRONTEND_ARTIFACT_DIR=str(run/'web'))
        subprocess.run(['node',str(root/'web/tests/native-browser.mjs')],cwd=root,env=web_env,check=True)
        check('both Node frontends: native login, navigation, editing and authenticated Range download')
    (run/'result.json').write_text(json.dumps({'checks':checks,'passed':True},indent=2),encoding='utf-8')
except Exception:
    (run/'result.json').write_text(json.dumps({'checks':checks,'passed':False,'error':traceback.format_exc()},indent=2),encoding='utf-8');traceback.print_exc();raise
finally:
    for p in reversed(processes):
        if p.poll() is None:p.terminate()
    for p in processes:
        try:p.wait(timeout=5)
        except subprocess.TimeoutExpired:p.kill();p.wait()
    for log in logs:log.close()
    if governance is not None:governance.close()
    print('Artifacts:',run,flush=True)
