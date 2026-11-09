"""Verify only: save application -> enqueue -> start worker -> consume and log."""
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import time
from urllib.error import URLError
from urllib.request import Request, urlopen
import uuid

ROOT = Path(__file__).resolve().parent.parent


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def request(base, path, payload=None):
    body = json.dumps(payload).encode() if payload is not None else None
    with urlopen(Request(base+path, data=body, headers={"Content-Type":"application/json"}), timeout=8) as response:
        return json.load(response)


def main():
    mysql_host=os.environ.get("MYSQL_TEST_HOST","127.0.0.1")
    mysql_port=os.environ.get("MYSQL_TEST_PORT","3306")
    mysql_user=os.environ.get("MYSQL_TEST_USER","root")
    mysql_password=os.environ.get("MYSQL_TEST_PASSWORD","")
    database="ms_test_"+uuid.uuid4().hex+"_mes"
    redis_port, http_port=free_port(),free_port()
    base=f"http://127.0.0.1:{http_port}"
    env=dict(os.environ, LC_ALL="C", REDIS_ADDR=f"127.0.0.1:{redis_port}", REDIS_PASSWORD="", HTTP_ADDR=f"127.0.0.1:{http_port}", MES_MYSQL_DSN=f"{mysql_user}:{mysql_password}@tcp({mysql_host}:{mysql_port})/{database}")
    processes=[]
    with tempfile.TemporaryDirectory(prefix="mes-queue-test-") as temp:
        logfiles=[]
        def start(name,args,cwd):
            log=open(Path(temp)/(name+".log"),"w+");logfiles.append(log)
            process=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT)
            processes.append(process)
            return log
        try:
            subprocess.run([str(ROOT/".bin/mes"),"migrate"],cwd=ROOT/"mes",env=env,check=True)
            start("redis",["redis-server","--bind","127.0.0.1","--port",str(redis_port),"--save","","--appendonly","no"],temp)
            for _ in range(100):
                try:
                    with socket.create_connection(("127.0.0.1",redis_port),timeout=.2):break
                except OSError:time.sleep(.1)
            else:raise AssertionError("Redis did not start")
            start("mes",[str(ROOT/".bin/mes")],ROOT/"mes")
            for _ in range(100):
                try:
                    if request(base,"/health")["code"]==0:break
                except (URLError,OSError):pass
                time.sleep(.1)
            else:raise AssertionError("MES did not start")
            payload={"clientRequestId":str(uuid.uuid4()),"processId":"PR20261109001","warehouseId":"WH001","date":"2026-11-09","remark":"入队与消费验收","items":[{"materialCode":"RM001","quantity":"12.2500","remark":""}]}
            result=request(base,"/api/v1/material-requisitions",payload)
            assert result["code"]==0,result
            record=result["data"]["record"]
            assert record["status"]=="queued",record
            duplicate=request(base,"/api/v1/material-requisitions",payload)
            assert duplicate["data"]["record"]["id"]==record["id"]
            # Explicitly keep the worker stopped so the queued state is observable.
            time.sleep(1)
            records=request(base,"/api/v1/material-requisitions")["data"]["records"]
            assert len(records)==1 and records[0]["status"]=="queued"
            print("OK worker 停止时：申请保存成功，任务保持排队",flush=True)
            worker_log=start("worker",[str(ROOT/".bin/mes"),"worker"],ROOT/"mes")
            for _ in range(100):
                record=request(base,"/api/v1/material-requisitions")["data"]["records"][0]
                if record["status"]=="consumed":break
                time.sleep(.1)
            else:raise AssertionError(record)
            assert record["items"][0]["quantity"]=="12.2500"
            worker_log.flush();worker_log.seek(0);output=worker_log.read()
            assert "[领料任务消费]" in output and payload["clientRequestId"] in output
            print("OK 启动 worker 后：消费任务、读取 MySQL 申请、打印日志",flush=True)
        except Exception:
            for log in logfiles:
                log.flush();log.seek(0);print(log.read()[-3000:])
            raise
        finally:
            for process in reversed(processes):
                if process.poll() is None:process.terminate()
                try:process.wait(timeout=15)
                except subprocess.TimeoutExpired:process.kill();process.wait()
            for log in logfiles:log.close()
            subprocess.run(["mysql","--protocol=TCP","--host="+mysql_host,"--port="+mysql_port,"--user="+mysql_user,"--connect-timeout=3","-e",f"DROP DATABASE IF EXISTS `{database}`"],env=dict(os.environ,MYSQL_PWD=mysql_password),check=True)


if __name__ == "__main__":
    main()
