#!/usr/bin/env python3
"""Copy jenkins-install.sh to remote host and run it via paramiko."""

import paramiko
import sys
import time

HOST = '192.168.168.158'
USER = 'prabhat'
PASS = '123456'
LOCAL_SCRIPT = 'scripts/bash/jenkins-install.sh'
REMOTE_SCRIPT = '/tmp/jenkins-install.sh'

def main():
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    print(f"Connecting to {USER}@{HOST}...")
    client.connect(HOST, username=USER, password=PASS, timeout=30)
    print("Connected.")

    # Copy script
    sftp = client.open_sftp()
    sftp.put(LOCAL_SCRIPT, REMOTE_SCRIPT)
    sftp.chmod(REMOTE_SCRIPT, 0o755)
    sftp.close()
    print(f"Copied {LOCAL_SCRIPT} -> {REMOTE_SCRIPT}")

    # Run with sudo
    cmd = f"echo '{PASS}' | sudo -S bash {REMOTE_SCRIPT}"
    print("Running jenkins-install.sh (this will take several minutes)...\n")
    stdin, stdout, stderr = client.exec_command(cmd, get_pty=True, timeout=3600)

    for line in iter(stdout.readline, ''):
        print(line, end='', flush=True)

    exit_code = stdout.channel.recv_exit_status()
    if exit_code != 0:
        print(f"\nERROR: script exited with code {exit_code}", file=sys.stderr)
        for line in stderr:
            print(line, end='', flush=True)
        sys.exit(exit_code)

    print("\nDone.")
    client.close()

if __name__ == '__main__':
    main()
