# Kubernetes 배포 가이드

## 1. TLS 인증서 생성

Mutating Webhook은 TLS를 사용하므로 인증서가 필요합니다.

```bash
# 인증서 생성
openssl genrsa -out tls.key 2048
openssl req -new -x509 -key tls.key -out tls.crt -days 365 \
  -subj "/CN=seal-automate.default.svc"

# Kubernetes Secret 생성
kubectl create secret tls seal-automate-tls \
  --cert=tls.crt \
  --key=tls.key \
  --dry-run=client -o yaml > tls-secret.yaml
```

## 2. Docker 이미지 빌드 및 푸시

```bash
# 이미지 빌드
docker build -t seal-automate:latest .

# 레지스트리에 푸시 (선택사항)
# docker tag seal-automate:latest <registry>/seal-automate:latest
# docker push <registry>/seal-automate:latest
```

## 3. Kubernetes 배포

```bash
# TLS Secret 생성
kubectl apply -f k8s/tls-secret.yaml

# Deployment 및 Service 배포
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# MutatingWebhookConfiguration 배포
kubectl apply -f k8s/mutating-webhook.yaml
```

## 4. 배포 확인

```bash
# Pod 상태 확인
kubectl get pods -l app=seal-automate

# 로그 확인
kubectl logs -l app=seal-automate

# Service 확인
kubectl get svc seal-automate
```

## 5. 테스트

`auto-seal: "true"` annotation이 있는 Secret을 생성하여 테스트:

```bash
kubectl create secret generic test-secret \
  --from-literal=username=admin \
  --from-literal=password=secret \
  --dry-run=client -o yaml | \
  kubectl annotate --local -f - auto-seal=true -o yaml | \
  kubectl apply -f -
```

## 주의사항

- MutatingWebhookConfiguration의 `namespaceSelector`를 사용하여 특정 네임스페이스에만 적용할 수 있습니다.
- TLS 인증서의 CN(Common Name)은 Service의 FQDN과 일치해야 합니다: `<service-name>.<namespace>.svc`
- Webhook이 실패하면 Secret 생성이 실패할 수 있으므로 `failurePolicy`를 적절히 설정하세요.

