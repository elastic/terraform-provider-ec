#!/usr/bin/env groovy

node('docker && gobld/machineType:n1-highcpu-8') {
    String DOCKER_IMAGE = "golang:1.19"
    String APP_PATH = "/go/src/github.com/elastic/terraform-provider-ec"
    String STABLE_TF_VERSION = "1.2.9"

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
            matrix {
                axes {
                    axis {
                        name 'TF_VERSION'
                        values "${STABLE_TF_VERSION}" 'latest'
                    }
                }
                stage("Run acceptance tests") {
                    acc_env = TF_VERSION != 'latest' ? ["TF_ACC_TERRAFORM_VERSION=${TF_VERSION}"] : []
                    withEnv(acc_env) {
                        sh "${TF_ACC_TERRAFORM_VERSION} make testacc-ci"
                    }
                }
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
