pipeline {
    agent any
    environment {
        GOROOT = "${tool type: 'go', name: 'go1.15.6'}/go"
    }
    stages {
        stage('build') {
            steps {
                echo 'building...'
                sh 'echo $GOROOT'
                sh '$GOROOT/bin/go build'
            }
        }
        stage('test') {
            steps {
                echo 'testing...'
            }
        }
        stage('deploy') {
            steps {
                echo 'deploying...'
                sh 'sudo systemctl stop go-chat'
                sh 'sudo cp go-chat /etc/go-chat/go-chat'
                sh 'sudo cp -r templates/* /etc/go-chat/templates'
                sh 'sudo systemctl start go-chat'
            }
        }
    }
    post {
        cleanup {
            deleteDir()
        }
    }
}