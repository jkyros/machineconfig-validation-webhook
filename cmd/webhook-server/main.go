package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	ctrlcommon "github.com/openshift/machine-config-operator/pkg/controller/common"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

func validateAndWhodunnit(req *admission.AdmissionRequest) ([]patchOperation, error) {

	mcResource := metav1.GroupVersionResource{
		Group:    "machineconfiguration.openshift.io",
		Version:  "v1",
		Resource: "machineconfigs",
	}
	if req.Resource != mcResource {
		log.Printf("expect resource to be %s", mcResource)
		return nil, nil
	}

	mc := mcfgv1.MachineConfig{}
	if _, _, err := universalDeserializer.Decode(req.Object.Raw, nil, &mc); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	var skipKey = "machineconfiguration.openshift.io/skip-validation"
	if _, skip := mc.Annotations[skipKey]; skip {
		log.Printf("Ignition validation was manually skipped for %s", mc.Name)
		return nil, nil
	}

	var patches []patchOperation
	if err := ctrlcommon.ValidateMachineConfig(mc.Spec); err != nil {
		return nil, fmt.Errorf("MachineConfig '%s' contains invalid ignition: %w", mc.Name, err)
	}

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/metadata/annotations/machineconfiguration.openshift.io~1last-modified-by",
		Value: req.UserInfo.Username,
	})

	if podname, haspodname := req.UserInfo.Extra["authentication.kubernetes.io/pod-name"]; haspodname {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/annotations/machineconfiguration.openshift.io~1last-modified-by-pod",
			Value: podname,
		})

	}

	// Make it clear we didn't skip validation on the way in
	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/metadata/annotations/machineconfiguration.openshift.io~1ignition-validated",
		Value: "true",
	})

	return patches, nil
}

func main() {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(validateAndWhodunnit))

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}
