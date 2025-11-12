package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		log.Printf("Error unmarshaling AdmissionReview: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// AdmissionReview v1beta1 지원을 위한 버전 설정
	if review.APIVersion == "" {
		review.APIVersion = "admission.k8s.io/v1"
	}
	if review.Kind == "" {
		review.Kind = "AdmissionReview"
	}

	req := review.Request
	if req == nil {
		log.Printf("Error: AdmissionReview.Request is nil")
		http.Error(w, "AdmissionReview.Request is nil", http.StatusBadRequest)
		return
	}

	log.Printf("Processing request: UID=%s, Kind=%s.%s, Operation=%s, Namespace=%s, Name=%s",
		req.UID, req.Kind.Kind, req.Kind.Group, req.Operation, req.Namespace, req.Name)

	var patchBytes []byte

	if req.Kind.Kind == "Secret" && req.Operation == admissionv1.Create {
		var secret corev1.Secret
		if err := json.Unmarshal(req.Object.Raw, &secret); err != nil {
			log.Printf("Error unmarshaling Secret: %v", err)
			// 에러가 있어도 계속 진행 (Secret이 아닐 수 있음)
		} else {
			log.Printf("Secret annotations: %v", secret.Annotations)
			if val, ok := secret.Annotations["auto-seal"]; ok && val == "true" {
				log.Printf("Found auto-seal=true annotation, processing...")
				// 여기서 암호화 + SealedSecret 생성 로직 들어감
				// patchBytes = … (JSONPatch)
				// TODO: 실제 SealedSecret 변환 로직 구현
			} else {
				log.Printf("No auto-seal annotation or value is not 'true'")
			}
		}
	} else {
		log.Printf("Skipping: Kind=%s, Operation=%s", req.Kind.Kind, req.Operation)
	}

	// AdmissionResponse 생성
	response := &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	if patchBytes != nil {
		patchType := admissionv1.PatchTypeJSONPatch
		response.PatchType = &patchType
		response.Patch = patchBytes
		log.Printf("Returning patch with %d bytes", len(patchBytes))
	} else {
		log.Printf("No patch to apply, allowing request")
	}

	// AdmissionReview 응답 생성 (Request는 nil로 설정)
	reviewResponse := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: review.APIVersion,
			Kind:       review.Kind,
		},
		Response: response,
	}

	respBytes, err := json.Marshal(reviewResponse)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Sending response: %d bytes", len(respBytes))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func main() {
	// 로그 설정
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting seal-automate webhook server...")

	// Health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Health check from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/mutate", handler)

	// TLS certificate paths (can be overridden by environment variables)
	certPath := getEnv("TLS_CERT_PATH", "/tls/tls.crt")
	keyPath := getEnv("TLS_KEY_PATH", "/tls/tls.key")

	log.Printf("Serving on :8443 (cert: %s, key: %s)", certPath, keyPath)
	if err := http.ListenAndServeTLS(":8443", certPath, keyPath, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
