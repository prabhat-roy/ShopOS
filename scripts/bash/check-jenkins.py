#!/usr/bin/env python3
import paramiko

HOST = '192.168.168.158'
USER = 'prabhat'
PASS = '123456'

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASS, timeout=30)

cmds = [
    "systemctl is-active jenkins 2>/dev/null || echo 'not running'",
    "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/login 2>/dev/null || echo 'unreachable'",
    "cat /var/lib/jenkins/jenkins-setup-complete 2>/dev/null && echo 'SETUP COMPLETE' || echo 'setup not complete'",
    "ls /var/lib/jenkins/plugins 2>/dev/null | wc -l | xargs echo 'plugins installed:'",
    "hostname -I | awk '{print $1}' | xargs -I{} echo 'URL: http://{}:8080'",
]

for cmd in cmds:
    _, stdout, _ = client.exec_command(cmd)
    out = stdout.read().decode().strip()
    print(f"$ {cmd}")
    print(out)
    print('---')

client.close()
