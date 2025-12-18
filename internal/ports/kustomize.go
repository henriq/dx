package ports

// KustomizeClient applies kustomize patches to Kubernetes manifests.
type KustomizeClient interface {
	// Apply takes raw YAML manifests and applies patches, returning patched YAML.
	// workDir is the directory where kustomize files will be written for inspection.
	// If a patch target is not found in the manifests, it logs a warning and continues.
	Apply(manifests []byte, patches []Patch, workDir string) ([]byte, error)
}

// Patch represents a kustomize patch configuration.
type Patch struct {
	// Target selects which resources to patch
	Target PatchTarget
	// Operations are JSON patch operations to apply
	Operations []PatchOperation
}

// PatchTarget identifies which Kubernetes resources to patch.
type PatchTarget struct {
	Kind string
	Name string // Empty means match all of this kind
}

// PatchOperation is a single JSON patch operation.
type PatchOperation struct {
	Op    string      // "add", "replace", "remove"
	Path  string      // JSON pointer path
	Value interface{} // Value for add/replace
}
