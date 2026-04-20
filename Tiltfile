# Tiltfile — ShopOS local dev hot-reload (alongside Skaffold)
# Usage: tilt up [service-name]
# Requires: kubectl configured to local cluster (kind/minikube/k3d)

load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://namespace', 'namespace_create', 'namespace_inject')

# ── Config ────────────────────────────────────────────────────────────────────
config.define_string_list('services', usage='Services to run (default: core set)')
cfg = config.parse()
services_to_run = cfg.get('services', [
    'api-gateway',
    'auth-service',
    'product-catalog-service',
    'order-service',
    'cart-service',
    'payment-service',
    'inventory-service',
    'notification-orchestrator',
])

# ── Namespaces ────────────────────────────────────────────────────────────────
namespace_create('platform')
namespace_create('identity')
namespace_create('catalog')
namespace_create('commerce')

# ── Infrastructure (minimal local) ───────────────────────────────────────────
helm_repo('bitnami', 'https://charts.bitnami.com/bitnami')

helm_resource('postgres',   'bitnami/postgresql', namespace='infra', flags=['--set=auth.postgresPassword=localdev'])
helm_resource('redis',      'bitnami/redis',      namespace='infra', flags=['--set=auth.enabled=false'])
helm_resource('kafka',      'bitnami/kafka',      namespace='infra', flags=['--set=kraft.enabled=true', '--set=replicaCount=1'])
helm_resource('mongodb',    'bitnami/mongodb',    namespace='infra', flags=['--set=auth.rootPassword=localdev'])

# ── Service hot-reload helper ─────────────────────────────────────────────────
def shopos_service(name, domain, port, lang='go', deps=None):
    if name not in services_to_run:
        return

    src_dir = 'src/{}/{}'.format(domain, name)

    if lang == 'go':
        local_resource(
            name + '-build',
            cmd='cd {} && go build -o /tmp/{} .'.format(src_dir, name),
            deps=[src_dir],
            labels=[domain],
        )
        docker_build(
            'shopos/' + name,
            src_dir,
            live_update=[
                sync(src_dir, '/app'),
                run('cd /app && go build -o /tmp/{} .'.format(name)),
            ],
        )
    elif lang == 'node':
        docker_build(
            'shopos/' + name,
            src_dir,
            live_update=[
                sync(src_dir, '/app'),
                run('cd /app && npm install --prefer-offline'),
            ],
        )
    elif lang == 'python':
        docker_build(
            'shopos/' + name,
            src_dir,
            live_update=[
                sync(src_dir, '/app'),
                run('pip install -r /app/requirements.txt'),
            ],
        )
    else:
        docker_build('shopos/' + name, src_dir)

    k8s_yaml(helm('helm/charts/{}'.format(name), name=name, namespace=domain, set=[
        'image.repository=shopos/' + name,
        'image.tag=dev',
        'image.pullPolicy=Never',
    ]))
    k8s_resource(name, port_forwards=[port], labels=[domain], resource_deps=deps or [])


# ── Core services ─────────────────────────────────────────────────────────────
shopos_service('api-gateway',             'platform',  8080, deps=['postgres', 'redis'])
shopos_service('auth-service',            'identity',  50060, lang='rust', deps=['postgres'])
shopos_service('product-catalog-service', 'catalog',   50070, deps=['mongodb'])
shopos_service('inventory-service',       'catalog',   50074, deps=['postgres'])
shopos_service('order-service',           'commerce',  50082, deps=['postgres', 'kafka'])
shopos_service('cart-service',            'commerce',  50080, deps=['redis'])
shopos_service('payment-service',         'commerce',  50083, deps=['postgres'])
shopos_service('notification-orchestrator', 'communications', 0, lang='node', deps=['kafka'])
