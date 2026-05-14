package domain

import (
	"fmt"
	"sort"
	"strings"
)

// DockerImageSourceOverrides maps a Docker image name to a local directory that
// should be used as the build source in place of cloning the image's configured
// git repository.
type DockerImageSourceOverrides struct {
	sourcePathByImage map[string]string
}

func NewDockerImageSourceOverrides(sourcePathByImage map[string]string) DockerImageSourceOverrides {
	copied := make(map[string]string, len(sourcePathByImage))
	for imageName, sourcePath := range sourcePathByImage {
		copied[imageName] = sourcePath
	}
	return DockerImageSourceOverrides{sourcePathByImage: copied}
}

func (o DockerImageSourceOverrides) IsEmpty() bool {
	return len(o.sourcePathByImage) == 0
}

func (o DockerImageSourceOverrides) LookupSourcePath(imageName string) (string, bool) {
	path, ok := o.sourcePathByImage[imageName]
	return path, ok
}

// FindUnusedOverrides returns the image names with overrides that are not
// present in the given list, sorted lexically.
func (o DockerImageSourceOverrides) FindUnusedOverrides(imagesInScope []DockerImage) []string {
	if o.IsEmpty() {
		return nil
	}
	inScope := map[string]struct{}{}
	for _, image := range imagesInScope {
		inScope[image.Name] = struct{}{}
	}
	var unused []string
	for imageName := range o.sourcePathByImage {
		if _, ok := inScope[imageName]; !ok {
			unused = append(unused, imageName)
		}
	}
	sort.Strings(unused)
	return unused
}

// ValidateAgainstImages returns an error if any override targets an image name
// not present in the given list.
func (o DockerImageSourceOverrides) ValidateAgainstImages(images []DockerImage) error {
	if o.IsEmpty() {
		return nil
	}

	knownImages := map[string]struct{}{}
	for _, image := range images {
		knownImages[image.Name] = struct{}{}
	}

	var unknownImages []string
	for imageName := range o.sourcePathByImage {
		if _, ok := knownImages[imageName]; !ok {
			unknownImages = append(unknownImages, imageName)
		}
	}
	if len(unknownImages) == 0 {
		return nil
	}
	sort.Strings(unknownImages)
	return fmt.Errorf(
		"image(s) not found: %s; available images:\n%s",
		strings.Join(quoteAll(unknownImages), ", "),
		availableImageNames(images),
	)
}

func availableImageNames(images []DockerImage) string {
	names := make([]string, 0, len(images))
	for _, image := range images {
		names = append(names, image.Name)
	}
	sort.Strings(names)
	for index, name := range names {
		names[index] = "  - " + name
	}
	return strings.Join(names, "\n")
}
