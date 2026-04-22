#!/usr/bin/env python3
"""Resume Jenkins setup from post-install configuration."""

import paramiko
import sys

HOST = '192.168.168.158'
USER = 'prabhat'
PASS = '123456'

SCRIPT = r"""
set -euo pipefail
log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# Fix any half-configured packages
log "Fixing dpkg state..."
DEBIAN_FRONTEND=noninteractive dpkg --configure -a

# Stop Jenkins if running
systemctl stop jenkins 2>/dev/null || true
sleep 3

# Disable setup wizard
log "Disabling setup wizard..."
mkdir -p /etc/systemd/system/jenkins.service.d
cat > /etc/systemd/system/jenkins.service.d/override.conf << 'EOF'
[Service]
Environment="JAVA_OPTS=-Djenkins.install.runSetupWizard=false"
EOF
systemctl daemon-reload

# Admin password via groovy init
log "Setting admin password..."
mkdir -p /var/lib/jenkins/init.groovy.d
cat > /var/lib/jenkins/init.groovy.d/01-security.groovy << 'GROOVY'
import jenkins.model.*
import hudson.security.*
def instance = Jenkins.getInstance()
def realm = new HudsonPrivateSecurityRealm(false)
realm.createAccount("admin", "admin")
instance.setSecurityRealm(realm)
def strategy = new FullControlOnceLoggedInAuthorizationStrategy()
strategy.setAllowAnonymousRead(false)
instance.setAuthorizationStrategy(strategy)
instance.save()
GROOVY
chown -R jenkins:jenkins /var/lib/jenkins/init.groovy.d

# Passwordless sudo for jenkins
log "Granting jenkins user passwordless sudo..."
echo "jenkins ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/jenkins
chmod 440 /etc/sudoers.d/jenkins

# Start Jenkins
log "Starting Jenkins..."
systemctl enable jenkins
systemctl start jenkins

# Wait for Jenkins
log "Waiting for Jenkins to be ready..."
for i in $(seq 1 30); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login || true)
  if [ "$STATUS" = "200" ]; then
    log "Jenkins is up!"
    break
  fi
  log "  Attempt $i/30 -- status: $STATUS. Retrying in 15s..."
  sleep 15
done

# Install plugins
log "Installing Jenkins plugins..."
JENKINS_WAR=$(find /usr/share/jenkins -name "jenkins.war" 2>/dev/null | head -1)
PLUGIN_CLI_JAR=$(find /usr/share/jenkins -name "jenkins-plugin-manager*.jar" 2>/dev/null | head -1)

if [ -z "$PLUGIN_CLI_JAR" ]; then
  log "Downloading jenkins-plugin-manager..."
  PLUGIN_CLI_VERSION=$(curl -s https://api.github.com/repos/jenkinsci/plugin-installation-manager-tool/releases/latest \
    | python3 -c 'import sys,json; print(json.load(sys.stdin)["tag_name"])' 2>/dev/null || echo "2.13.2")
  curl -sfL "https://github.com/jenkinsci/plugin-installation-manager-tool/releases/download/${PLUGIN_CLI_VERSION}/jenkins-plugin-manager-${PLUGIN_CLI_VERSION#v}.jar" \
    -o /tmp/jenkins-plugin-manager.jar
  PLUGIN_CLI_JAR=/tmp/jenkins-plugin-manager.jar
fi

PLUGINS="workflow-aggregator pipeline-stage-view pipeline-graph-analysis pipeline-utility-steps \
  git git-client github github-branch-source gitlab-plugin bitbucket ssh-agent \
  maven-plugin gradle nodejs ant \
  docker-workflow docker-plugin kubernetes kubernetes-credentials kubernetes-cli \
  terraform ansible \
  sonar dependency-check-jenkins-plugin warnings-ng jacoco htmlpublisher junit cobertura \
  nexus-artifact-uploader artifactory \
  credentials credentials-binding ssh-credentials plain-credentials role-strategy matrix-auth \
  email-ext slack mailer \
  blueocean timestamper ansicolor build-timeout ws-cleanup dashboard-view \
  parameterized-trigger rebuild copy-artifact throttle-concurrents \
  generic-webhook-trigger job-dsl configuration-as-code prometheus monitoring \
  multibranch-scan-webhook-trigger console-column-plugin"

java -jar "$PLUGIN_CLI_JAR" \
  --war "$JENKINS_WAR" \
  --plugin-download-directory /var/lib/jenkins/plugins \
  --plugins $PLUGINS \
  --verbose 2>&1 | tail -30

chown -R jenkins:jenkins /var/lib/jenkins/plugins

# Restart to activate plugins
log "Restarting Jenkins to activate plugins..."
systemctl restart jenkins

log "Waiting for Jenkins to come back..."
for i in $(seq 1 30); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login || true)
  if [ "$STATUS" = "200" ]; then
    log "Jenkins is ready!"
    break
  fi
  log "  Attempt $i/30 -- status: $STATUS. Retrying in 15s..."
  sleep 15
done

touch /var/lib/jenkins/jenkins-setup-complete

LOCAL_IP=$(hostname -I | awk '{print $1}')
log "============================================================"
log "Jenkins setup complete!"
log "  URL:      http://${LOCAL_IP}:8080"
log "  Username: admin"
log "  Password: admin"
log "============================================================"
"""

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
print(f"Connecting to {USER}@{HOST}...")
client.connect(HOST, username=USER, password=PASS, timeout=30)
print("Connected. Running post-install setup...\n")

cmd = f"echo '{PASS}' | sudo -S bash -s"
stdin, stdout, stderr = client.exec_command(cmd, get_pty=True, timeout=3600)
stdin.write(SCRIPT)
stdin.channel.shutdown_write()

for line in iter(stdout.readline, ''):
    print(line.encode('ascii', errors='replace').decode('ascii'), end='', flush=True)

exit_code = stdout.channel.recv_exit_status()
if exit_code != 0:
    print(f"\nERROR: exit code {exit_code}", file=sys.stderr)
    sys.exit(exit_code)

print("\nDone.")
client.close()
