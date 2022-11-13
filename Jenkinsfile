#!groovy
pipeline {
    agent none
    stages {
        stage('Build') {
            agent any
            steps {
                sh 'docker compose -f docker-compose.prod.yml build  --build-arg user=bysoft'
            }
        }
    }
}