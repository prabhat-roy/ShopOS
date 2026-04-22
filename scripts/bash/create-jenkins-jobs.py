#!/usr/bin/env python3
"""Create all 11 Jenkins pipeline jobs in execution order via Jenkins REST API."""
import urllib.request
import urllib.parse
import urllib.error
import http.cookiejar
import json
import sys
import time

JENKINS = "http://35.239.30.237:8080"
USER = "admin"
PASS = "admin"
REPO_URL = "https://github.com/prabhat-roy/ShopOS.git"
REPO_BRANCH = "*/main"

# Jobs in execution order — name → Jenkinsfile path in repo
JOBS = [
    ("01-install-tools",      "ci/jenkins/install-tools.Jenkinsfile"),
    ("02-k8s-infrastructure", "ci/jenkins/k8s-infra.Jenkinsfile"),
    ("03-cluster-bootstrap",  "ci/jenkins/cluster-bootstrap.Jenkinsfile"),
    ("04-registry",           "ci/jenkins/registry.Jenkinsfile"),
    ("05-databases",          "ci/jenkins/databases.Jenkinsfile"),
    ("06-messaging",          "ci/jenkins/messaging.Jenkinsfile"),
    ("07-streaming",          "ci/jenkins/streaming.Jenkinsfile"),
    ("08-networking",         "ci/jenkins/networking.Jenkinsfile"),
    ("09-security",           "ci/jenkins/security.Jenkinsfile"),
    ("10-observability",      "ci/jenkins/observability.Jenkinsfile"),
    ("11-docker-build",       "ci/jenkins/docker-build.Jenkinsfile"),
    ("12-deploy",             "ci/jenkins/deploy.Jenkinsfile"),
    ("13-post-deploy",        "ci/jenkins/post-deploy.Jenkinsfile"),
    ("14-gitops",             "ci/jenkins/gitops.Jenkinsfile"),
]

# Old jobs to delete before creating updated numbered ones
OLD_JOBS = [
    "01-install-tools", "02-k8s-infrastructure", "03-cluster-bootstrap", "04-registry",
    "05-messaging", "06-networking", "07-security", "08-observability",
    "09-deploy", "10-post-deploy", "11-gitops",
]

def auth_header():
    import base64
    token = base64.b64encode(f"{USER}:{PASS}".encode()).decode()
    return {"Authorization": f"Basic {token}"}

# Shared opener with cookie jar so session persists across requests
_cj = http.cookiejar.CookieJar()
_opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(_cj))

def open_url(req, timeout=15):
    return _opener.open(req, timeout=timeout)

def get_crumb():
    url = f"{JENKINS}/crumbIssuer/api/json"
    req = urllib.request.Request(url, headers=auth_header())
    with open_url(req) as r:
        d = json.loads(r.read())
    return {d["crumbRequestField"]: d["crumb"]}

def job_exists(name, crumb):
    url = f"{JENKINS}/job/{urllib.parse.quote(name)}/api/json"
    req = urllib.request.Request(url, headers={**auth_header(), **crumb})
    try:
        open_url(req, timeout=10)
        return True
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return False
        raise

def delete_job(name, crumb):
    url = f"{JENKINS}/job/{urllib.parse.quote(name)}/doDelete"
    req = urllib.request.Request(url, data=b"", headers={**auth_header(), **crumb}, method="POST")
    try:
        with open_url(req, timeout=10) as r:
            return r.status
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return 404
        raise

def create_job(name, jenkinsfile_path, crumb):
    config_xml = f"""<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <description>ShopOS pipeline: {name}</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty/>
    <hudson.model.BuildDiscarderProperty>
      <strategy class="hudson.tasks.LogRotator">
        <daysToKeepStr>-1</daysToKeepStr>
        <numToKeepStr>10</numToKeepStr>
        <artifactDaysToKeepStr>-1</artifactDaysToKeepStr>
        <artifactNumToKeepStr>-1</artifactNumToKeepStr>
      </strategy>
    </hudson.model.BuildDiscarderProperty>
  </properties>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition" plugin="workflow-cps">
    <scm class="hudson.plugins.git.GitSCM" plugin="git">
      <configVersion>2</configVersion>
      <userRemoteConfigs>
        <hudson.plugins.git.UserRemoteConfig>
          <url>{REPO_URL}</url>
        </hudson.plugins.git.UserRemoteConfig>
      </userRemoteConfigs>
      <branches>
        <hudson.plugins.git.BranchSpec>
          <name>{REPO_BRANCH}</name>
        </hudson.plugins.git.BranchSpec>
      </branches>
      <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
      <submoduleCfg class="empty-list"/>
      <extensions/>
    </scm>
    <scriptPath>{jenkinsfile_path}</scriptPath>
    <lightweight>true</lightweight>
  </definition>
  <triggers/>
  <disabled>false</disabled>
</flow-definition>"""

    url = f"{JENKINS}/createItem?name={urllib.parse.quote(name)}"
    data = config_xml.encode("utf-8")
    headers = {
        **auth_header(),
        **crumb,
        "Content-Type": "application/xml",
        "Content-Length": str(len(data)),
    }
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")
    try:
        with open_url(req, timeout=15) as r:
            return r.status
    except urllib.error.HTTPError as e:
        print(f"  HTTP {e.code}: {e.read().decode()[:200]}")
        raise

def update_job(name, jenkinsfile_path, crumb):
    config_xml = f"""<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <description>ShopOS pipeline: {name}</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty/>
    <hudson.model.BuildDiscarderProperty>
      <strategy class="hudson.tasks.LogRotator">
        <daysToKeepStr>-1</daysToKeepStr>
        <numToKeepStr>10</numToKeepStr>
      </strategy>
    </hudson.model.BuildDiscarderProperty>
  </properties>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition" plugin="workflow-cps">
    <scm class="hudson.plugins.git.GitSCM" plugin="git">
      <configVersion>2</configVersion>
      <userRemoteConfigs>
        <hudson.plugins.git.UserRemoteConfig>
          <url>{REPO_URL}</url>
        </hudson.plugins.git.UserRemoteConfig>
      </userRemoteConfigs>
      <branches>
        <hudson.plugins.git.BranchSpec>
          <name>{REPO_BRANCH}</name>
        </hudson.plugins.git.BranchSpec>
      </branches>
      <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
      <submoduleCfg class="empty-list"/>
      <extensions/>
    </scm>
    <scriptPath>{jenkinsfile_path}</scriptPath>
    <lightweight>true</lightweight>
  </definition>
  <triggers/>
  <disabled>false</disabled>
</flow-definition>"""

    url = f"{JENKINS}/job/{urllib.parse.quote(name)}/config.xml"
    data = config_xml.encode("utf-8")
    headers = {
        **auth_header(),
        **crumb,
        "Content-Type": "application/xml",
        "Content-Length": str(len(data)),
    }
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")
    with urllib.request.urlopen(req, timeout=15) as r:
        return r.status

def main():
    print(f"Connecting to Jenkins at {JENKINS}...")
    crumb = get_crumb()

    print(f"Deleting {len(OLD_JOBS)} old unnumbered jobs...\n")
    for name in OLD_JOBS:
        if job_exists(name, crumb):
            delete_job(name, crumb)
            print(f"  Deleted: {name}")
            crumb = get_crumb()
        else:
            print(f"  Skip (not found): {name}")

    print(f"\nCreating {len(JOBS)} numbered jobs in order...\n")

    for i, (name, jenkinsfile) in enumerate(JOBS, 1):
        print(f"[{i:2d}/{len(JOBS)}] {name}")
        print(f"        Jenkinsfile: {jenkinsfile}")
        try:
            if job_exists(name, crumb):
                update_job(name, jenkinsfile, crumb)
                print(f"        Updated existing job.")
            else:
                create_job(name, jenkinsfile, crumb)
                print(f"        Created.")
            # Refresh crumb after each job (sessions can expire)
            crumb = get_crumb()
        except Exception as e:
            print(f"        ERROR: {e}")
            sys.exit(1)

    print(f"\nAll {len(JOBS)} jobs created successfully.")
    print(f"\nJenkins URL: {JENKINS}")
    print("Jobs (in execution order):")
    for i, (name, _) in enumerate(JOBS, 1):
        print(f"  {i:2d}. {JENKINS}/job/{name}/")

if __name__ == "__main__":
    main()
