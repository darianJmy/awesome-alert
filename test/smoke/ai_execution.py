import json, os, subprocess, tempfile, threading, time, urllib.request, socket
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
state={'calls':0,'fail':False}
class Model(BaseHTTPRequestHandler):
    def log_message(self,*args): pass
    def do_POST(self):
        body=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        state['calls']+=1
        if state['fail']:
            self.send_response(400);self.end_headers();self.wfile.write(b'private-provider-error');return
        if len(body['messages'])==1:
            reply={'type':'tool_call','tool':'query_prometheus_instant','params':{'datasource_id':99999,'promql':'up'}}
        else:
            reply={'type':'final','answer':'未找到指定数据源，请检查数据源配置。'}
        payload=json.dumps({'content':[{'type':'text','text':json.dumps(reply,ensure_ascii=False)}]}).encode()
        self.send_response(200);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(payload)
model=ThreadingHTTPServer(('127.0.0.1',0),Model)
threading.Thread(target=model.serve_forever,daemon=True).start()
with socket.socket() as sock:
    sock.bind(('127.0.0.1',0));port=sock.getsockname()[1]
http=urllib.request.build_opener(urllib.request.ProxyHandler({}))
def request(path,data=None,token=None):
    headers={'Content-Type':'application/json'}
    if token: headers['Authorization']='Bearer '+token
    req=urllib.request.Request(f'http://127.0.0.1:{port}'+path,data=None if data is None else json.dumps(data).encode(),headers=headers)
    return json.loads(http.open(req,timeout=15).read())
try:
 with tempfile.TemporaryDirectory(prefix='guanlan-ai-') as tmp:
    env={k:v for k,v in os.environ.items() if not k.startswith('APP_')}
    env.update(APP_SERVER_ADDR=f'127.0.0.1:{port}',APP_AI_BASE_URL=f'http://127.0.0.1:{model.server_port}',APP_AI_API_KEY='mock-key',APP_DATABASE_DSN=str(Path(tmp)/'data.db'))
    with open(Path(tmp)/'server.log','w') as log:
     proc=subprocess.Popen([str(Path(__file__).resolve().parents[2] / 'guanlan')],cwd=tmp,env=env,stdout=log,stderr=log)
     try:
      for _ in range(100):
       try: token=request('/api/login',{'username':'admin','password':'admin123'})['data']['token'];break
       except OSError: time.sleep(.1)
      else: raise AssertionError('server not ready')
      message=request('/api/chat',{'message':'查询数据源99999的状态'},token)['data']
      trace=json.loads(message['metadata'])['execution']
      assert trace['status']=='succeeded' and len(trace['steps'])==3,trace
      assert trace['steps'][1]['error_code']=='DATASOURCE_ERROR',trace
      conversation=message['conversation_id']
      saved=request(f'/api/conversations/{conversation}/messages',token=token)['data']
      assert json.loads(saved[-1]['metadata'])['execution']['run_id']==trace['run_id']
      before=state['calls'];state['fail']=True
      failed=request('/api/chat',{'conversation_id':conversation,'message':'帮我新增一个Prometheus数据源'},token)['data']
      metadata=json.loads(failed['metadata'])
      assert metadata['execution']['status']=='failed' and state['calls']==before+1,metadata
      assert 'private-provider-error' not in json.dumps(failed)
      assert 'pending_action' not in metadata
      print('PASS: authenticated chat, tool trace, persisted history, failed-run trace, no legacy fallback, no provider-error leak')
     finally:
      proc.terminate();proc.wait(timeout=10)
finally:
 model.shutdown();model.server_close()
