pipeline {
    agent any
    tools {
        go 'Go-1.20'
    }

    stages {
        stage('Build') {
            steps {
                echo 'Building...'
                sh 'go buiild -o carnac.exe main.go'
            }
        }
        stage('Test') {
            steps {
                echo 'Testing...'
            }
        }
        stage('Deploy') {
            steps {
                echo 'Deploying...'
            }
        }
    }
}

post {
    always {
        echo 'This will always run after the stages.'
    }
    success {
        echo 'Pipeline completed successfully.'
    }
    failure {
        echo 'Pipeline failed.'
    }
}