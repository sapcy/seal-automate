# SAPCY Sealed Secret Automate

Kubernetes Mutating Webhook을 사용하여 Secret을 자동으로 SealedSecret으로 변환하는 도구입니다.

## 기능

- Kubernetes Admission Controller로 동작
- `auto-seal: "true"` annotation이 있는 Secret을 자동으로 SealedSecret으로 변환
- kubeseal과 동일한 암호화 방식 사용

## 프로젝트 구조

```
kubeseal-automate/
├── main.go                    # 메인 애플리케이션
├── Dockerfile                 # Docker 이미지 빌드 파일
├── .dockerignore             # Docker 빌드 시 제외할 파일
├── go.mod                     # Go 모듈 정의
├── go.sum                     # 의존성 체크섬
└── k8s/                       # Kubernetes 매니페스트
    ├── deployment.yaml        # Deployment
    ├── service.yaml           # Service
    ├── mutating-webhook.yaml  # MutatingWebhookConfiguration
    ├── tls-secret.yaml.example # TLS Secret 예시
    └── README.md              # 배포 가이드
```

## 로컬 빌드

```bash
go build -o seal-automate main.go
```

## Docker 빌드

```bash
docker build -t seal-automate:latest .
```

## Kubernetes 배포

자세한 배포 가이드는 [k8s/README.md](k8s/README.md)를 참조하세요.

### 빠른 시작

```bash
# 1. TLS 인증서 생성
openssl genrsa -out tls.key 2048
openssl req -new -x509 -key tls.key -out tls.crt -days 365 \
  -subj "/CN=seal-automate.default.svc"

# 2. Kubernetes Secret 생성
kubectl create secret tls seal-automate-tls \
  --cert=tls.crt \
  --key=tls.key

# 3. 배포
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/mutating-webhook.yaml
```

## 사용법

Secret에 `auto-seal: "true"` annotation을 추가하면 자동으로 SealedSecret으로 변환됩니다:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  annotations:
    auto-seal: "true"
data:
  username: YWRtaW4=
  password: c2VjcmV0
```

## 개발

### 로컬 테스트

```bash
# TLS 인증서 생성 (로컬 테스트용)
openssl genrsa -out tls.key 2048
openssl req -new -x509 -key tls.key -out tls.crt -days 365 \
  -subj "/CN=localhost"

# 인증서를 /tls 디렉토리에 복사
mkdir -p tls
cp tls.crt tls/tls.crt
cp tls.key tls/tls.key

# 애플리케이션 실행
./sapcy_sealed_secret_automate
```
