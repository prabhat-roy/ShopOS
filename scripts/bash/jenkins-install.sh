#!/bin/bash
set -euo pipefail

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# ── System Update ─────────────────────────────────────────────────────────────
log "Updating and upgrading system..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get upgrade -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold"

# ── Java 21 ───────────────────────────────────────────────────────────────────
log "Installing Java 21..."
apt-get install -y fontconfig openjdk-21-jdk curl wget gnupg2

java -version

# ── Jenkins Repository ────────────────────────────────────────────────────────
log "Adding Jenkins repository..."
gpg --keyserver hkps://keyserver.ubuntu.com --recv-keys 7198F4B714ABFC68
gpg --export 7198F4B714ABFC68 | tee /usr/share/keyrings/jenkins-keyring.gpg > /dev/null

echo "deb [signed-by=/usr/share/keyrings/jenkins-keyring.gpg] \
  https://pkg.jenkins.io/debian-stable binary/" \
  | tee /etc/apt/sources.list.d/jenkins.list > /dev/null

apt-get update -y

# ── Jenkins Install ───────────────────────────────────────────────────────────
log "Installing Jenkins..."
apt-get install -y jenkins

# ── Stop Jenkins (auto-started by apt) ───────────────────────────────────────
log "Stopping Jenkins to apply config before first real start..."
systemctl stop jenkins 2>/dev/null || true
sleep 5

# ── Disable Setup Wizard ──────────────────────────────────────────────────────
log "Disabling setup wizard..."
mkdir -p /etc/systemd/system/jenkins.service.d
cat > /etc/systemd/system/jenkins.service.d/override.conf << 'EOF'
[Service]
Environment="JAVA_OPTS=-Djenkins.install.runSetupWizard=false"
EOF
systemctl daemon-reload

# ── Admin Password via Groovy Init Script ─────────────────────────────────────
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

# ── Passwordless sudo for jenkins user ───────────────────────────────────────
log "Granting jenkins user passwordless sudo..."
echo "jenkins ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/jenkins
chmod 440 /etc/sudoers.d/jenkins

# ── Start Jenkins ─────────────────────────────────────────────────────────────
log "Starting Jenkins..."
systemctl enable jenkins
systemctl start jenkins

# ── Wait for Jenkins ──────────────────────────────────────────────────────────
log "Waiting for Jenkins to be ready..."
for i in $(seq 1 30); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login || true)
  if [ "$STATUS" = "200" ]; then
    log "Jenkins is up!"
    break
  fi
  log "  Attempt $i/30 — status: $STATUS. Retrying in 15s..."
  sleep 15
done

# ── Install Plugins via jenkins-plugin-cli (synchronous) ─────────────────────
log "Installing Jenkins plugins via jenkins-plugin-cli..."

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
  --verbose 2>&1 | tail -20

chown -R jenkins:jenkins /var/lib/jenkins/plugins

# ── Restart to Activate Plugins ───────────────────────────────────────────────
log "Restarting Jenkins to activate all plugins..."
systemctl restart jenkins

log "Waiting for Jenkins to come back..."
for i in $(seq 1 30); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login || true)
  if [ "$STATUS" = "200" ]; then
    log "Jenkins is ready!"
    break
  fi
  log "  Attempt $i/30 — status: $STATUS. Retrying in 15s..."
  sleep 15
done

PUBLIC_IP=$(curl -s https://ipv4.icanhazip.com)
log "============================================================"
log "Jenkins setup complete!"
log "  URL:      http://${PUBLIC_IP}:8080"
log "  Username: admin"
log "  Password: admin"
log "============================================================"

touch /var/lib/jenkins/jenkins-setup-complete
