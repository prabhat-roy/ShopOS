#!/bin/bash
set -euo pipefail

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# ── System Update ─────────────────────────────────────────────────────────────
log "Updating system..."
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

# ── Stop Jenkins ──────────────────────────────────────────────────────────────
systemctl stop jenkins 2>/dev/null || true
sleep 3

# ── Disable Setup Wizard ──────────────────────────────────────────────────────
log "Disabling setup wizard..."
mkdir -p /etc/systemd/system/jenkins.service.d
cat > /etc/systemd/system/jenkins.service.d/override.conf << 'EOF'
[Service]
Environment="JAVA_OPTS=-Djenkins.install.runSetupWizard=false"
EOF
systemctl daemon-reload

# ── Admin User via Groovy Init ────────────────────────────────────────────────
log "Setting admin credentials..."
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

# ── Passwordless sudo for jenkins ─────────────────────────────────────────────
log "Granting jenkins passwordless sudo..."
echo "jenkins ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/jenkins
chmod 440 /etc/sudoers.d/jenkins

# ── Start Jenkins ─────────────────────────────────────────────────────────────
log "Starting Jenkins..."
systemctl enable jenkins
systemctl start jenkins

# ── Wait for Jenkins to be ready ─────────────────────────────────────────────
log "Waiting for Jenkins to be ready..."
for i in $(seq 1 30); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login || true)
  if [ "$STATUS" = "200" ]; then
    log "Jenkins is up!"
    break
  fi
  log "  Attempt $i/30 — status: $STATUS. Retrying in 10s..."
  sleep 10
done

# ── Mark setup complete ───────────────────────────────────────────────────────
touch /var/lib/jenkins/jenkins-setup-complete

PUBLIC_IP=$(curl -s -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip 2>/dev/null || echo "unknown")

log "============================================================"
log "Jenkins ready! Plugins will be installed separately."
log "  URL:      http://${PUBLIC_IP}:8080"
log "  Username: admin"
log "  Password: admin"
log "============================================================"
