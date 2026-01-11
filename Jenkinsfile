pipeline {
    agent any

    tools {
        go 'go-1.25'
        nodejs 'nodejs-25.2.1'
    }

    environment {
        REGISTRY = 'registry.manty.co.kr'
        BACKEND_IMAGE = "${REGISTRY}/mmessenger-backend"
        FRONTEND_IMAGE = "${REGISTRY}/mmessenger-frontend"
        KUBECONFIG = credentials('kubeconfig')
    }

    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 30, unit: 'MINUTES')
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Test Backend') {
            steps {
                sh 'go test ./... -v -cover'
            }
        }

        stage('Test Frontend') {
            steps {
                dir('frontend') {
                    sh 'npm ci'
                    sh 'npm run lint || true'
                }
            }
        }

        stage('Build Backend Image') {
            steps {
                script {
                    docker.build("${BACKEND_IMAGE}:${BUILD_NUMBER}", ".")
                    docker.build("${BACKEND_IMAGE}:latest", ".")
                }
            }
        }

        stage('Build Frontend Image') {
            steps {
                script {
                    dir('frontend') {
                        docker.build("${FRONTEND_IMAGE}:${BUILD_NUMBER}", ".")
                        docker.build("${FRONTEND_IMAGE}:latest", ".")
                    }
                }
            }
        }

        stage('Push Images') {
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY}", 'docker-registry-credentials') {
                        docker.image("${BACKEND_IMAGE}:${BUILD_NUMBER}").push()
                        docker.image("${BACKEND_IMAGE}:latest").push()
                        docker.image("${FRONTEND_IMAGE}:${BUILD_NUMBER}").push()
                        docker.image("${FRONTEND_IMAGE}:latest").push()
                    }
                }
            }
        }

        stage('Deploy to Kubernetes') {
            steps {
                script {
                    sh """
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/namespace.yaml
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/configmap.yaml
                        # Note: secret.yaml must be created manually before first deployment
                        # kubectl create secret generic mmessenger-secret --from-literal=DB_USER=xxx --from-literal=DB_PASS=xxx --from-literal=JWT_SECRET=xxx -n messenger
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/backend-deployment.yaml
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/backend-service.yaml
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/frontend-deployment.yaml
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/frontend-service.yaml
                        kubectl --kubeconfig=\$KUBECONFIG apply -f k8s/ingress.yaml
                    """

                    // Update image tags
                    sh """
                        kubectl --kubeconfig=\$KUBECONFIG set image deployment/mmessenger-backend \
                            mmessenger-backend=${BACKEND_IMAGE}:${BUILD_NUMBER} \
                            -n messenger
                        kubectl --kubeconfig=\$KUBECONFIG set image deployment/mmessenger-frontend \
                            mmessenger-frontend=${FRONTEND_IMAGE}:${BUILD_NUMBER} \
                            -n messenger
                    """

                    // Wait for rollout
                    sh """
                        kubectl --kubeconfig=\$KUBECONFIG rollout status deployment/mmessenger-backend -n messenger --timeout=300s
                        kubectl --kubeconfig=\$KUBECONFIG rollout status deployment/mmessenger-frontend -n messenger --timeout=300s
                    """
                }
            }
        }
    }

    post {
        success {
            echo 'Deployment successful!'
        }
        failure {
            echo 'Deployment failed!'
        }
        always {
            cleanWs()
        }
    }
}
