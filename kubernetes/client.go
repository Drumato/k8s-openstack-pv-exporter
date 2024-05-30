package kubernetes

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgok8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubeconfigEnvKey = "KUBECONFIG"
)

type Client interface {
	ListPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error)
}

// NewDefaultClient initializes a Client that uses in-cluster config to authenticate the kube-apiserver.
// if this fails to get in-cluster config, this tries to find the kubeconfig file and use it instead.
func NewDefaultClient() (Client, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		clientset, err := clientgok8s.NewForConfig(config)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return &DefaultClient{clientset}, nil
	}

	kubeconfigFilePath := os.Getenv(kubeconfigEnvKey)
	if kubeconfigFilePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		kubeconfigFilePath = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigFilePath)
	if err != nil {
		// both trying failed.
		return nil, errors.WithStack(err)
	}

	clientset, err := clientgok8s.NewForConfig(config)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &DefaultClient{clientset}, nil
}

type DefaultClient struct {
	clientset *clientgok8s.Clientset
}

func (dc *DefaultClient) ListPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error) {
	pvlist, err := dc.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return pvlist, nil
}
