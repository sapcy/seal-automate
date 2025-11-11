package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := review.Request
	var patchBytes []byte

	if req.Kind.Kind == "Secret" && req.Operation == admissionv1.Create {
		var secret corev1.Secret
		if err := json.Unmarshal(req.Object.Raw, &secret); err != nil {
			// error 처리
		}
		if val, ok := secret.Annotations["auto-seal"]; ok && val == "true" {
			// 여기서 암호화 + SealedSecret 생성 로직 들어감
			// patchBytes = … (JSONPatch)
		}
	}

	review.Response = &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}
	if patchBytes != nil {
		patchType := admissionv1.PatchTypeJSONPatch
		review.Response.PatchType = &patchType
		review.Response.Patch = patchBytes
	}

	respBytes, _ := json.Marshal(review)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func main() {
	// Health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/mutate", handler)

	// TLS certificate paths (can be overridden by environment variables)
	certPath := getEnv("TLS_CERT_PATH", "/tls/tls.crt")
	keyPath := getEnv("TLS_KEY_PATH", "/tls/tls.key")

	fmt.Printf("Serving on :8443 (cert: %s, key: %s)\n", certPath, keyPath)
	if err := http.ListenAndServeTLS(":8443", certPath, keyPath, nil); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
