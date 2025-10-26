package container_orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"dx/internal/core"
	"dx/internal/core/domain"
	"dx/internal/ports"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubernetes represents a client for interacting with Kubernetes
type Kubernetes struct {
	configRepository  core.ConfigRepository
	secretsRepository core.SecretsRepository
	templater         ports.Templater
	clientSet         *kubernetes.Clientset
	postRendererPath  string
}

func ProvideKubernetes(
	configRepository core.ConfigRepository,
	secretsRepository core.SecretsRepository,
	templater ports.Templater,
) (*Kubernetes, error) {
	// Try to load the kubeConfig from the default location
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}
	kubeConfigPath := filepath.Join(home, ".kube", "config")

	// Create the config from the kubeConfig file
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
	}

	// Create the clientSet
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	configContextName, err := configRepository.LoadCurrentContextName()
	if err != nil {
		return nil, err
	}

	postRendererPath := filepath.Join(home, ".dx", configContextName, "helm-post-renderer.sh")
	if runtime.GOOS == "windows" {
		postRendererPath = filepath.Join(home, ".dx", configContextName, "helm-post-renderer.bat")
	}

	return &Kubernetes{
		configRepository:  configRepository,
		secretsRepository: secretsRepository,
		templater:         templater,
		clientSet:         clientSet,
		postRendererPath:  postRendererPath,
	}, nil
}

// CreateClusterEnvironmentKey creates a string that is used to uniquely identify the cluster and namespace
func (k *Kubernetes) CreateClusterEnvironmentKey() (string, error) {
	// Get cluster ID from kube-system namespace UID
	kubeSystemNS, err := k.clientSet.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace: %v", err)
	}
	clusterUID := string(kubeSystemNS.UID)

	// Get the current namespace from context
	namespace := ""
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err == nil && config.CurrentContext != "" {
		if currentContext, ok := config.Contexts[config.CurrentContext]; ok && currentContext.Namespace != "" {
			namespace = currentContext.Namespace
		}
	}

	// Fail if no namespace is set
	if namespace == "" {
		return "", fmt.Errorf("no namespace set in current context")
	}

	// Create a deterministic key based only on cluster UID and namespace
	key := fmt.Sprintf("%s-%s", clusterUID, namespace)

	hash := sha256.New()
	hash.Write([]byte(key))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil)), nil
}

// InstallService Installs a service using helm
func (k *Kubernetes) InstallService(service *domain.Service) error {
	fmt.Println("Installing service", service.Name)
	// Construct the helm install command
	cmd := exec.Command(
		"helm",
		"upgrade",
		"--install",
		"--labels",
		"managed-by=dx",
		service.Name,
		fmt.Sprintf("%s/%s", service.HelmPath, service.HelmChartRelativePath),
		fmt.Sprintf("--post-renderer=%s", k.postRendererPath),
	)
	templateValues, err := core.CreateTemplatingValues(k.configRepository, k.secretsRepository)
	if err != nil {
		return err
	}

	for i, arg := range service.HelmArgs {
		renderedArg, err := k.templater.Render(arg, fmt.Sprintf("helm-args.%d", i), templateValues)
		if err != nil {
			return err
		}
		cmd.Args = append(cmd.Args, renderedArg)
	}

	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install helm chart: %v, output: %s", err, string(output))
	}

	return nil
}

func (k *Kubernetes) InstallDevProxy(service *domain.Service) error {
	fmt.Println("Installing service", service.Name)
	// Construct the helm install command
	cmd := exec.Command(
		"helm",
		"upgrade",
		"--install",
		"--labels",
		"managed-by=dx",
		service.Name,
		fmt.Sprintf("%s/%s", service.HelmPath, service.HelmChartRelativePath),
	)
	templateValues, err := core.CreateTemplatingValues(k.configRepository, k.secretsRepository)
	if err != nil {
		return err
	}

	for i, arg := range service.HelmArgs {
		renderedArg, err := k.templater.Render(arg, fmt.Sprintf("helm-args.%d", i), templateValues)
		if err != nil {
			return err
		}
		cmd.Args = append(cmd.Args, renderedArg)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install helm chart: %v, output: %s", err, string(output))
	}

	return nil
}

// UninstallService deletes a service using helm uninstall
func (k *Kubernetes) UninstallService(service *domain.Service) error {
	fmt.Println("Uninstalling service", service.Name)
	// Construct the helm uninstall command
	cmd := exec.Command("helm", "uninstall", service.Name)

	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall helm chart: %v, output: %s", err, string(output))
	}

	return nil
}

func (k *Kubernetes) HasDeployedServices() (bool, error) {
	cmd := exec.Command("helm", "list", "-l", "managed-by=dx", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to list helm charts: %v, output: %s", err, string(output))
	}
	return len(strings.Split(strings.TrimSpace(string(output)), "\n")) > 1, nil
}
