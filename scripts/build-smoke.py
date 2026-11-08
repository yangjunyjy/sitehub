from pathlib import Path
import subprocess,concurrent.futures,os
root=Path(__file__).resolve().parents[1];out=root/'.tmp/smoke/bin';out.mkdir(parents=True,exist_ok=True)
targets={f'{d}-{kind}':f'./service/{d}/{kind}' for d in ['rbac','blog','product','file'] for kind in ['api','rpc']}
targets.update(scheduler='./service/scheduler/cronjob',bootstrap='./tools/bootstrap',redis='./tests/e2e/redis',gatewayctl='./tools/gatewayctl',gateway='./gateway')
def build(item):
    name,path=item;r=subprocess.run(['go','build','-o',str(out/(name+('.exe' if os.name=='nt' else ''))),path],cwd=root,capture_output=True,text=True,encoding='utf-8',errors='replace')
    return name,r.returncode,r.stderr
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    results=list(pool.map(build,targets.items()))
for name,code,error in results:
    print(name,'FAIL' if code else 'OK',error.strip())
if any(code for _,code,_ in results):raise SystemExit(1)
