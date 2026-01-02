pipeline {
    agent any
    tools {
        go 'go_1.25'
    }

    stages {
        stage('Development') {
            steps {
                git 'https://github.com/dodderingstalwart/carnac.git'
            }
        }
        stage('Build') {
            steps {
                script {
                    sh 'go build -o carnac.exe ./cmd/carnac'
                }
            }
        }
    }
}
