#!/usr/bin/env groovy

node('docker && gobld/machineType:n1-highcpu-8') {
    String DOCKER_IMAGE = "golang:1.16"
    String APP_PATH = "/go/src/github.com/elastic/terraform-provider-ec"

    stage('Checkout from GitHub') {
	    checkout scm
    }
    withCredentials([
        string(credentialsId: 'vault-addr', variable: 'VAULT_ADDR'),
        string(credentialsId: 'vault-secret-id', variable: 'VAULT_SECRET_ID'),
        string(credentialsId: 'vault-role-id', variable: 'VAULT_ROLE_ID')
    ]) {
        stage("Get secrets from vault") {
            withEnv(["VAULT_SECRET_ID=${VAULT_SECRET_ID}", "VAULT_ROLE_ID=${VAULT_ROLE_ID}", "VAULT_ADDR=${VAULT_ADDR}"]) {
                sh 'make -C .ci .gpg_private .gpg_passphrase .github_token .gpg_fingerprint'
            }
        }
    }
    docker.image("${DOCKER_IMAGE}").inside("-u root:root -v ${pwd()}:${APP_PATH} -w ${APP_PATH}") {
        try {
            stage("Download dependencies") {
                sh 'make vendor'
            }
            stage("Import gpg key") {
                sh 'make -C .ci import-gpg-key'
            }
            stage("Cache GPG key and release the binaries") {
                sh '. .ci/.env; make -C .ci cache-gpg-passphrase; make release'
            }
        } catch (Exception err) {
            throw err
        } finally {
            stage("Clean up") {
                sh 'make -C .ci clean'
                sh 'rm -rf dist bin'
            }
        }
    }
}
