package ports

// HelmClient defines the interface for interacting with Helm.
type HelmClient interface {
	// Template renders a helm chart and returns the manifests as YAML.
	Template(name, chartPath, namespace string, args []string) ([]byte, error)
	// UpgradeFromManifests installs/upgrades using pre-rendered manifests in a wrapper chart.
	UpgradeFromManifests(name, namespace, wrapperChartPath string) error
	// Uninstall removes a helm release.
	Uninstall(name, namespace string) error
	// List returns release names matching the label selector.
	List(labelSelector, namespace string) ([]string, error)
}
