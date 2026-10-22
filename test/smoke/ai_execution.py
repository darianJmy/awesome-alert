import json, os, subprocess, tempfile, threading, time, urllib.request, socket
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
state={'calls':0,'fail':False,'models':[]}
class Model(BaseHTTPRequestHandler):
    def log_message(self,*args): pass
    def do_POST(self):
        body=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        state['calls']+=1
        state['models'].append(body['model'])
        if state['fail']:
            self.send_response(400);self.end_headers();self.wfile.write(b'private-provider-error');return
        assert body.get('tools'), 'native tools missing'
        if body['messages'][-1]['role']=='user':
            reply={'role':'assistant','tool_calls':[{'id':'call_1','type':'function','function':{'name':'query_prometheus_instant','arguments':json.dumps({'datasource_id':99999,'promql':'up'})}}]}
        else:
            assert body['messages'][-1]['role']=='tool'
            assert body['messages'][-1]['tool_call_id']=='call_1'
            reply={'role':'assistant','content':'未找到指定数据源，请检查数据源配置。'}
        payload=json.dumps({'choices':[{'message':reply,'finish_reason':'tool_calls' if 'tool_calls' in reply else 'stop'}]}).encode()
        self.send_response(200);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(payload)
model=ThreadingHTTPServer(('127.0.0.1',0),Model)
threading.Thread(target=model.serve_forever,daemon=True).start()
with socket.socket() as sock:
    sock.bind(('127.0.0.1',0));port=sock.getsockname()[1]
http=urllib.request.build_opener(urllib.request.ProxyHandler({}))
def request(path,data=None,token=None,method=None):
    headers={'Content-Type':'application/json'}
    if token: headers['Authorization']='Bearer '+token
    req=urllib.request.Request(f'http://127.0.0.1:{port}'+path,data=None if data is None else json.dumps(data).encode(),headers=headers,method=method)
    return json.loads(http.open(req,timeout=15).read())
try:
 with tempfile.TemporaryDirectory(prefix='guanlan-ai-') as tmp:
    env={k:v for k,v in os.environ.items() if not k.startswith('APP_')}
    env.update(APP_SERVER_ADDR=f'127.0.0.1:{port}',APP_DATABASE_DSN=str(Path(tmp)/'data.db'))
    with open(Path(tmp)/'server.log','w') as log:
     proc=subprocess.Popen([os.environ.get('GUANLAN_BINARY',str(Path(__file__).resolve().parents[2] / 'guanlan'))],cwd=tmp,env=env,stdout=log,stderr=log)
     try:
      for _ in range(100):
       try: token=request('/api/login',{'username':'admin','password':'admin123'})['data']['token'];break
       except OSError: time.sleep(.1)
      else: raise AssertionError('server not ready')
      assert request('/api/chat/models',token=token)['data']==[]
      config={'name':'模型一','provider':'newapi','base_url':f'http://127.0.0.1:{model.server_port}','model':'model-one','api_key':'mock-key','status':1}
      first=request('/api/ai-models',config,token)['data']
      config.update(name='模型二',model='model-two')
      second=request('/api/ai-models',config,token)['data']
      assert 'mock-key' not in json.dumps(first)
      message=request('/api/chat',{'message':'查询数据源99999的状态','model_id':second['id']},token)['data']
      assert state['models']==['model-two','model-two'],state
      assert json.loads(message['metadata'])['model_id']==second['id']
      assert json.loads(message['metadata'])['engine']=='eino'
      trace=json.loads(message['metadata'])['execution']
      assert trace['status']=='succeeded' and len(trace['steps'])==3,trace
      assert trace['steps'][1]['error_code']=='DATASOURCE_ERROR',trace
      conversation=message['conversation_id']
      saved=request(f'/api/conversations/{conversation}/messages',token=token)['data']
      assert json.loads(saved[-1]['metadata'])['execution']['run_id']==trace['run_id']
      conversations=request('/api/conversations',token=token)['data']
      assert conversations[0]['model_id']==second['id']
      switched=request('/api/chat',{'conversation_id':conversation,'message':'再检查一次','model_id':first['id']},token)['data']
      assert state['models'][-1]=='model-one'
      before=state['calls'];state['fail']=True
      failed=request('/api/chat',{'conversation_id':conversation,'message':'帮我新增一个Prometheus数据源'},token)['data']
      metadata=json.loads(failed['metadata'])
      assert metadata['execution']['status']=='failed' and state['calls']==before+1,metadata
      assert 'private-provider-error' not in json.dumps(failed)
      assert 'pending_action' not in metadata
      assert state['models'][-1]=='model-one'
      config.update(status=0,api_key='')
      request(f"/api/ai-models/{second['id']}",config,token,method='PUT')
      try: request('/api/chat',{'conversation_id':conversation,'message':'不可用模型','model_id':second['id']},token)
      except urllib.error.HTTPError as err: assert err.code==400
      else: raise AssertionError('disabled model accepted')
      if os.environ.get('AI_UI_CHECK'):
       state['fail']=False
       subprocess.run(['node',os.environ['AI_UI_CHECK'],f'http://127.0.0.1:{port}'],check=True,env={**os.environ,'AI_TEST_TOKEN':token})
      print('PASS: DB model CRUD, selection, switch, disabled rejection;  authenticated chat, tool trace, persisted history, failed-run trace, no legacy fallback, no provider-error leak')
     finally:
      proc.terminate();proc.wait(timeout=10)
finally:
 model.shutdown();model.server_close()
