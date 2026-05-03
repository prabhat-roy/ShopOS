// Jenkins pipeline — runs `buf breaking` against the main branch on every PR.
// Failing breaks the build and prevents merge.
pipeline {
  agent { kubernetes { yamlFile 'ci/jenkins/agents/buf-agent.yaml' } }
  options { timestamps(); disableConcurrentBuilds() }
  stages {
    stage('Checkout') { steps { checkout scm } }
    stage('Buf format + lint') {
      steps {
        sh '''
          cd proto
          buf format --diff --exit-code
          buf lint
        '''
      }
    }
    stage('Buf breaking-change check') {
      steps {
        sh '''
          cd proto
          buf breaking --against "https://github.com/prabhat-roy/ShopOS.git#branch=main,subdir=proto"
        '''
      }
    }
    stage('Generate bindings (sanity)') {
      steps {
        sh 'cd proto && buf generate --template buf.gen.yaml'
      }
    }
  }
  post {
    failure {
      slackSend channel: '#ci-failures',
                color: 'danger',
                message: ":no_entry: buf breaking-change check failed on ${env.BRANCH_NAME} — review proto edits"
    }
  }
}
