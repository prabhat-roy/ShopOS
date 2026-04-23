// dashboard-links.groovy
// Shared helper — returns a formatted string of all ShopOS dashboard URLs.
// Usage in a Jenkinsfile stage:
//   def d = load 'scripts/groovy/dashboard-links.groovy'
//   echo d.call(envMap, title, contextVars)
//
// envMap     : Map parsed from infra.env
// title      : String — pipeline-specific header line
// contextVars: Map of optional context (service, tag, domain, etc.)

def call(Map envMap = [:], String title = 'DASHBOARD LINKS', Map ctx = [:]) {

    // ── Core URLs (from infra.env with sane cluster-DNS defaults) ────────────
    def grafana    = envMap['GRAFANA_URL']            ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'
    def prom       = envMap['PROMETHEUS_URL']          ?: 'http://prometheus-prometheus.prometheus.svc.cluster.local:9090'
    def argocd     = envMap['ARGOCD_URL']              ?: 'http://argocd-server.argocd.svc.cluster.local:80'
    def harbor     = envMap['HARBOR_URL']              ?: 'harbor.shopos.local'
    def sonar      = envMap['SONARQUBE_URL']           ?: 'http://sonarqube-sonarqube.sonarqube.svc.cluster.local:9000'
    def dojo       = envMap['DEFECTDOJO_URL']          ?: 'http://defectdojo-defectdojo.defectdojo.svc.cluster.local:8080'
    def deptrack   = envMap['DEPENDENCY_TRACK_URL']    ?: 'http://dependency-track.dependency-track.svc.cluster.local:8080'
    def jaeger     = envMap['JAEGER_URL']              ?: 'http://jaeger-query.tracing.svc.cluster.local:16686'
    def loki       = envMap['LOKI_URL']                ?: 'http://loki.loki.svc.cluster.local:3100'
    def alertmgr   = envMap['ALERTMANAGER_URL']        ?: 'http://alertmanager.prometheus.svc.cluster.local:9093'
    def pyrra      = envMap['PYRRA_URL']               ?: 'http://pyrra.monitoring.svc.cluster.local:9099'
    def kiali      = envMap['KIALI_URL']               ?: 'http://kiali.istio-system.svc.cluster.local:20001'
    def uptime     = envMap['UPTIME_KUMA_URL']         ?: 'http://uptime-kuma.monitoring.svc.cluster.local:3001'
    def pyroscope  = envMap['PYROSCOPE_URL']           ?: 'http://pyroscope.monitoring.svc.cluster.local:4040'
    def signoz     = envMap['SIGNOZ_URL']              ?: 'http://signoz.monitoring.svc.cluster.local:3301'
    def zipkin     = envMap['ZIPKIN_URL']              ?: 'http://zipkin.tracing.svc.cluster.local:9411'
    def pact       = envMap['PACT_BROKER_URL']         ?: 'http://pact-broker.testing.svc.cluster.local:9292'
    def akhq       = envMap['AKHQ_URL']                ?: 'http://akhq.kafka.svc.cluster.local:8080'
    def kafkaui    = envMap['KAFKA_UI_URL']            ?: 'http://kafka-ui.kafka.svc.cluster.local:8080'
    def backstage  = envMap['BACKSTAGE_URL']           ?: 'http://backstage.backstage.svc.cluster.local:7007'
    def nexus      = envMap['NEXUS_URL']               ?: 'http://nexus.registry.svc.cluster.local:8081'
    def gitea      = envMap['GITEA_URL']               ?: 'http://gitea.registry.svc.cluster.local:3000'
    def vault      = envMap['VAULT_URL']               ?: 'http://vault.vault.svc.cluster.local:8200'
    def keycloak   = envMap['KEYCLOAK_URL']            ?: 'http://keycloak.keycloak.svc.cluster.local:8080'
    def temporal   = envMap['TEMPORAL_UI_URL']         ?: 'http://temporal-ui.temporal.svc.cluster.local:8080'
    def mlflow     = envMap['MLFLOW_URL']              ?: 'http://mlflow.mlops.svc.cluster.local:5000'
    def unleash    = envMap['UNLEASH_URL']             ?: 'http://unleash.platform.svc.cluster.local:4242'
    def prefect    = envMap['PREFECT_URL']             ?: 'http://prefect.analytics.svc.cluster.local:4200'
    def dagster    = envMap['DAGSTER_URL']             ?: 'http://dagster.analytics.svc.cluster.local:3000'
    def airflow    = envMap['AIRFLOW_URL']             ?: 'http://airflow.analytics.svc.cluster.local:8080'
    def superset   = envMap['SUPERSET_URL']            ?: 'http://superset.analytics.svc.cluster.local:8088'
    def marquez    = envMap['MARQUEZ_URL']             ?: 'http://marquez-web.analytics.svc.cluster.local:3000'
    def mimir      = envMap['MIMIR_URL']               ?: 'http://grafana-mimir.monitoring.svc.cluster.local:9009'
    def victlogs   = envMap['VICTORIA_LOGS_URL']       ?: 'http://victoria-logs.logging.svc.cluster.local:9428'
    def netdata    = envMap['NETDATA_URL']             ?: 'http://netdata.monitoring.svc.cluster.local:19999'
    def perses     = envMap['PERSES_URL']              ?: 'http://perses.monitoring.svc.cluster.local:8080'
    def rabbitmq   = envMap['RABBITMQ_MGMT_URL']       ?: 'http://rabbitmq.messaging.svc.cluster.local:15672'
    def nats       = envMap['NATS_MONITORING_URL']     ?: 'http://nats.messaging.svc.cluster.local:8222'
    def conduktor  = envMap['CONDUKTOR_URL']           ?: 'http://conduktor.kafka.svc.cluster.local:8080'
    def apisix     = envMap['APISIX_DASHBOARD_URL']    ?: 'http://apisix-dashboard.api.svc.cluster.local:9000'
    def hasura     = envMap['HASURA_URL']              ?: 'http://hasura.api.svc.cluster.local:8080'
    def tyk        = envMap['TYK_DASHBOARD_URL']       ?: 'http://tyk.api.svc.cluster.local:3000'
    def nomad      = envMap['NOMAD_URL']               ?: 'http://nomad.platform.svc.cluster.local:4646'
    def oncall     = envMap['GRAFANA_ONCALL_URL']      ?: 'http://grafana-oncall.monitoring.svc.cluster.local:8080'
    def cachet     = envMap['CACHET_URL']              ?: 'http://cachet.monitoring.svc.cluster.local:8000'
    def pgadmin    = envMap['PGADMIN_URL']             ?: 'http://pgadmin.databases.svc.cluster.local:80'
    def mongoexp   = envMap['MONGO_EXPRESS_URL']       ?: 'http://mongo-express.databases.svc.cluster.local:8081'
    def rediscmdr  = envMap['REDIS_COMMANDER_URL']     ?: 'http://redis-commander.databases.svc.cluster.local:8081'
    def bytebase   = envMap['BYTEBASE_URL']            ?: 'http://bytebase.databases.svc.cluster.local:8080'
    def neo4j      = envMap['NEO4J_BROWSER_URL']       ?: 'http://neo4j.databases.svc.cluster.local:7474'
    def cockroach  = envMap['COCKROACHDB_URL']         ?: 'http://cockroachdb.databases.svc.cluster.local:8080'
    def atlas      = envMap['ATLAS_URL']               ?: 'http://atlas.governance.svc.cluster.local:21000'
    def surreal    = envMap['SURREALDB_URL']           ?: 'http://surrealdb.databases.svc.cluster.local:8000'
    def meilisrch  = envMap['MEILISEARCH_URL']         ?: 'http://meilisearch.databases.svc.cluster.local:7700'
    def consul     = envMap['CONSUL_URL']              ?: 'http://consul.consul.svc.cluster.local:8500'
    def traefik    = envMap['TRAEFIK_URL']             ?: 'http://traefik.traefik.svc.cluster.local:9000'
    def linkerd    = envMap['LINKERD_VIZ_URL']         ?: 'http://web.linkerd-viz.svc.cluster.local:8084'
    def argoflows  = envMap['ARGO_WORKFLOWS_URL']      ?: 'http://argo-workflows-server.argo.svc.cluster.local:2746'
    def argoroll   = envMap['ARGO_ROLLOUTS_URL']       ?: 'http://argo-rollouts-dashboard.argo.svc.cluster.local:3100'
    def weaveui    = envMap['FLUX_UI_URL']             ?: 'http://weave-gitops.gitops.svc.cluster.local:9001'
    def falcoui    = envMap['FALCO_SIDEKICK_URL']      ?: 'http://falco-falcosidekick-ui.falco.svc.cluster.local:2802'
    def wazuh      = envMap['WAZUH_URL']               ?: 'http://wazuh-dashboard.wazuh.svc.cluster.local:5601'
    def authentik  = envMap['AUTHENTIK_URL']           ?: 'http://authentik.authentik.svc.cluster.local:9000'
    def opensrchdb = envMap['OPENSEARCH_DASHBOARDS_URL'] ?: 'http://opensearch-dashboards.logging.svc.cluster.local:5601'
    def kibana     = envMap['KIBANA_URL']              ?: 'http://kibana.logging.svc.cluster.local:5601'
    def wiremock   = envMap['WIREMOCK_URL']            ?: 'http://wiremock.testing.svc.cluster.local:8080'
    def chaosmesh  = envMap['CHAOS_MESH_URL']          ?: 'http://chaos-mesh.chaos.svc.cluster.local:2333'
    def litmus     = envMap['LITMUS_URL']              ?: 'http://litmus.litmus.svc.cluster.local:9091'
    def locust     = envMap['LOCUST_URL']              ?: 'http://locust.testing.svc.cluster.local:8089'
    def openreplay = envMap['OPENREPLAY_URL']          ?: ''
    def plausible  = envMap['PLAUSIBLE_URL']           ?: ''
    def glitchtip  = envMap['GLITCHTIP_URL']           ?: 'http://glitchtip.monitoring.svc.cluster.local:8000'
    def reportsPrtal = envMap['REPORTS_PORTAL_URL']    ?: 'http://reports-portal-service.platform.svc.cluster.local:8300'

    // Context variables (service name, tag, etc.)
    def svc    = ctx['service'] ?: 'shopos'
    def tag    = ctx['tag']     ?: 'latest'
    def domain = ctx['domain']  ?: 'platform'
    def regPrj = ctx['project'] ?: 'shopos'

    return """
╔═══════════════════════════════════════════════════════════════════════════════╗
║  SHOPOS — ${title.padRight(60)}║
╠═══════════════════════════════════════════════════════════════════════════════╣
║  Service: ${svc.padRight(30)} Tag: ${tag.take(24).padRight(24)} ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║  OBSERVABILITY — METRICS & ALERTING                                           ║
║  Grafana (main)             : ${grafana}/dashboards
║  Grafana (service board)    : ${grafana}/d/services/service-overview?var-service=${svc}
║  Grafana (CI/CD board)      : ${grafana}/d/cicd/cicd-overview
║  Grafana (SLO board)        : ${grafana}/d/slo/slo-overview
║  Grafana (load test board)  : ${grafana}/d/k6/k6-load-test-results
║  Prometheus                 : ${prom}
║  Alertmanager               : ${alertmgr}
║  Pyrra (SLO budget)         : ${pyrra}
║  Grafana Mimir              : ${mimir}
║  VictoriaMetrics            : http://victoria-metrics.monitoring.svc.cluster.local:8428
║  Netdata (real-time)        : ${netdata}
║  Perses (dashboards-as-code): ${perses}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  OBSERVABILITY — TRACING                                                      ║
║  Jaeger                     : ${jaeger}/search?service=${svc}
║  Zipkin                     : ${zipkin}
║  Grafana Tempo              : ${grafana}/explore (select datasource=Tempo)
╠═══════════════════════════════════════════════════════════════════════════════╣
║  OBSERVABILITY — LOGS                                                         ║
║  Loki (via Grafana)         : ${grafana}/explore (select datasource=Loki)
║  Kibana                     : ${kibana}
║  OpenSearch Dashboards      : ${opensrchdb}
║  VictoriaLogs               : ${victlogs}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  OBSERVABILITY — PROFILING, ERRORS & SESSION                                  ║
║  Pyroscope (profiling)      : ${pyroscope}
║  SigNoz (full-stack OTel)   : ${signoz}
║  GlitchTip/Sentry (errors)  : ${glitchtip}
║  Uptime Kuma (status)       : ${uptime}
║  Cachet (public status page): ${cachet}${openreplay ? '\n║  OpenReplay (sessions)      : ' + openreplay : ''}${plausible ? '\n║  Plausible (analytics)      : ' + plausible : ''}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  MESSAGING & STREAMING                                                        ║
║  AKHQ (Kafka browser)       : ${akhq}
║  Kafka UI                   : ${kafkaui}
║  RabbitMQ Management        : ${rabbitmq}
║  NATS Monitoring            : ${nats}
║  Conduktor Gateway          : ${conduktor}
║  ksqlDB REST API            : http://ksqldb.kafka.svc.cluster.local:8088
╠═══════════════════════════════════════════════════════════════════════════════╣
║  GITOPS & CI/CD                                                               ║
║  ArgoCD (all apps)          : ${argocd}/applications
║  ArgoCD (${svc.take(20).padRight(20)}) : ${argocd}/applications/${svc}
║  Argo Workflows             : ${argoflows}
║  Argo Rollouts              : ${argoroll}
║  Flux / Weave GitOps        : ${weaveui}
║  Jenkins (CI)               : http://jenkins.ci.svc.cluster.local:8080
╠═══════════════════════════════════════════════════════════════════════════════╣
║  REGISTRY                                                                     ║
║  Harbor (image registry)    : https://${harbor}/harbor/projects
║  Harbor project             : https://${harbor}/harbor/projects/${regPrj}
║  Harbor image               : https://${harbor}/harbor/projects/${regPrj}/repositories/${svc}
║  Nexus (artifacts)          : ${nexus}
║  Gitea (source)             : ${gitea}
║  Rekor (transparency log)   : https://rekor.sigstore.dev
╠═══════════════════════════════════════════════════════════════════════════════╣
║  SECURITY — IAM & ACCESS                                                      ║
║  Keycloak (IAM/SSO)         : ${keycloak}
║  Vault (secrets/PKI)        : ${vault}/ui
║  Authentik (IdP)            : ${authentik}
║  Pomerium (zero-trust proxy): http://pomerium.pomerium.svc.cluster.local
║  OpenFGA (authz)            : http://openfga.openfga.svc.cluster.local:8080
╠═══════════════════════════════════════════════════════════════════════════════╣
║  SECURITY — VULNERABILITY MANAGEMENT                                          ║
║  DefectDojo (findings)      : ${dojo}/finding
║  DefectDojo (engagements)   : ${dojo}/engagement
║  Dependency-Track (SBOMs)   : ${deptrack}/projects
║  SonarQube (code quality)   : ${sonar}/dashboard?id=${svc}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  SECURITY — RUNTIME & NETWORK                                                 ║
║  Wazuh (SIEM/XDR)           : ${wazuh}
║  Falco Sidekick UI          : ${falcoui}
║  Kiali (service mesh)       : ${kiali}/kiali/graph
║  Traefik (edge router)      : ${traefik}/dashboard/
║  Consul (service discovery) : ${consul}/ui
║  Linkerd Viz                : ${linkerd}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  DATABASES                                                                    ║
║  pgAdmin 4 (PostgreSQL)     : ${pgadmin}
║  Mongo Express (MongoDB)    : ${mongoexp}
║  Redis Commander            : ${rediscmdr}
║  Bytebase (schema mgmt)     : ${bytebase}
║  Apache Superset (BI)       : ${superset}
║  Neo4j Browser              : ${neo4j}
║  CockroachDB Admin          : ${cockroach}
║  Apache Atlas (data catalog): ${atlas}
║  SurrealDB Studio           : ${surreal}
║  Meilisearch (search)       : ${meilisrch}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  PLATFORM & DEVELOPER TOOLS                                                   ║
║  Reports Portal (all reports): ${reportsPrtal}
║  Backstage (service catalog): ${backstage}/catalog/default/component/${svc}
║  Temporal UI (workflows)    : ${temporal}
║  MLflow (experiments)       : ${mlflow}
║  APISIX Dashboard           : ${apisix}
║  Hasura Console (GraphQL)   : ${hasura}
║  Tyk Dashboard (API mgmt)   : ${tyk}
║  Unleash (feature flags)    : ${unleash}
║  Nomad (workloads)          : ${nomad}/ui
║  Grafana OnCall             : ${oncall}
║  Pact Broker (contracts)    : ${pact}
║  WireMock (API mocks)       : ${wiremock}/__admin/
╠═══════════════════════════════════════════════════════════════════════════════╣
║  DATA & ANALYTICS PIPELINES                                                   ║
║  Apache Airflow (DAGs)      : ${airflow}
║  Prefect (workflows)        : ${prefect}
║  Dagster (assets/pipelines) : ${dagster}
║  Marquez (data lineage)     : ${marquez}
║  Apache Superset (BI)       : ${superset}
╠═══════════════════════════════════════════════════════════════════════════════╣
║  CHAOS ENGINEERING & LOAD TESTING                                             ║
║  Chaos Mesh Dashboard       : ${chaosmesh}
║  Litmus Chaos Center        : ${litmus}
║  Locust UI (during test)    : ${locust}
║  Gatling reports            : reports/load/gatling/ (Jenkins artifact)
╠═══════════════════════════════════════════════════════════════════════════════╣
║  KUBERNETES & NETWORKING                                                      ║
║  Argo Rollouts dashboard    : ${argoroll}
║  Kiali (mesh topology)      : ${kiali}/kiali/graph?namespaces=${domain}
║  k8sGPT                     : kubectl get k8sgpts -A
║  Botkube                    : (Slack channel alerts)
╚═══════════════════════════════════════════════════════════════════════════════╝"""
}

return this
