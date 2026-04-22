def call() {
    def cloud = ''
    if (env.CLOUD_PROVIDER) {
        cloud = env.CLOUD_PROVIDER
    } else if (fileExists('/var/lib/jenkins/infra.env')) {
        cloud = readFile('/var/lib/jenkins/infra.env').trim().split('\n')
            .find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: ''
    } else if (fileExists('infra.env')) {
        cloud = readFile('infra.env').trim().split('\n')
            .find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: ''
    }
    switch (cloud) {
        case 'AWS':   return 'gp2'
        case 'AZURE': return 'managed-csi'
        default:      return 'standard-rwo'   // GCP default
    }
}
return this
