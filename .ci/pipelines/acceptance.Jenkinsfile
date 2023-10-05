#!/usr/bin/env groovy

node('docker && gobld/machineType:n1-highcpu-8') {
    String DOCKER_IMAGE = "golang:1.21"
    String APP_PATH = "/go/src/github.com/elastic/terraform-provider-ec"

    stage('Checkout from GitHub') {
	    checkout scm
    }
    withCredentials([
        string(credentialsId: 'vault-addr', variable: 'VAULT_ADDR'),
        string(credentialsId: 'vault-secret-id', variable: 'VAULT_SECRET_ID'),
        string(credentialsId: 'vault-role-id', variable: 'VAULT_ROLE_ID')
    ]) {
        stage("Get EC_API_KEY from vault") {
            withEnv(["VAULT_SECRET_ID=${VAULT_SECRET_ID}", "VAULT_ROLE_ID=${VAULT_ROLE_ID}", "VAULT_ADDR=${VAULT_ADDR}"]) {
                sh 'make -C .ci .apikey'
            }
        }
    }
    docker.image("${DOCKER_IMAGE}").inside("-u root:root -v ${pwd()}:${APP_PATH} -w ${APP_PATH}") {
        try {
            stage("Download dependencies") {
                sh 'make vendor'
            }
            stage("Run acceptance tests") {
                sh 'make testacc-ci'
            }
        } catch (Exception err) {
            throw err
        } finally {
            stage("Clean up") {
                // Sweeps any deployments older than 1h.
                sh 'make sweep-ci'
                sh 'make -C .ci clean'
                sh 'rm -rf reports bin'
            }
        }
    }
}
