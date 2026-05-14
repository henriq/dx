package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"pilot/internal/core/domain"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHelmChartOverrides_ResolvesAbsolutePath(t *testing.T) {
	chartDirectory := t.TempDir()

	overrides, err := ParseHelmChartOverrides([]string{"api:" + chartDirectory})
	require.NoError(t, err)

	directory, ok := overrides.LookupChartDirectory("api")
	require.True(t, ok)
	assert.Equal(t, chartDirectory, directory)
}

func TestParseHelmChartOverrides_EmptyInputIsEmpty(t *testing.T) {
	overrides, err := ParseHelmChartOverrides(nil)
	require.NoError(t, err)
	assert.True(t, overrides.IsEmpty())
}

func TestParseHelmChartOverrides_RejectsMissingSeparator(t *testing.T) {
	_, err := ParseHelmChartOverrides([]string{"api"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service:path")
	assert.Contains(t, err.Error(), `"api"`)
}

func TestParseHelmChartOverrides_RejectsEmptyService(t *testing.T) {
	chartDirectory := t.TempDir()

	_, err := ParseHelmChartOverrides([]string{":" + chartDirectory})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service name cannot be empty")
}

func TestParseHelmChartOverrides_RejectsEmptyPath(t *testing.T) {
	_, err := ParseHelmChartOverrides([]string{"api:"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chart path cannot be empty")
}

func TestParseHelmChartOverrides_RejectsMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := ParseHelmChartOverrides([]string{"api:" + missing})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestParseHelmChartOverrides_RejectsControlCharacterInPath(t *testing.T) {
	_, err := ParseHelmChartOverrides([]string{"api:\x1b[31m/charts"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "control character")
}

func TestParseHelmChartOverrides_RejectsFileInsteadOfDirectory(t *testing.T) {
	tempDirectory := t.TempDir()
	filePath := filepath.Join(tempDirectory, "Chart.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte("name: x"), 0o600))

	_, err := ParseHelmChartOverrides([]string{"api:" + filePath})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestParseHelmChartOverrides_LastWinsOnDuplicateService(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()

	overrides, err := ParseHelmChartOverrides([]string{"api:" + first, "api:" + second})
	require.NoError(t, err)

	directory, ok := overrides.LookupChartDirectory("api")
	require.True(t, ok)
	assert.Equal(t, second, directory)
}

func TestParseImageSourceOverrides_ResolvesAbsolutePath(t *testing.T) {
	sourceDirectory := t.TempDir()

	overrides, err := ParseImageSourceOverrides([]string{"api:" + sourceDirectory})
	require.NoError(t, err)

	path, ok := overrides.LookupSourcePath("api")
	require.True(t, ok)
	assert.Equal(t, sourceDirectory, path)
}

func TestParseImageSourceOverrides_EmptyInputIsEmpty(t *testing.T) {
	overrides, err := ParseImageSourceOverrides(nil)
	require.NoError(t, err)
	assert.True(t, overrides.IsEmpty())
}

func TestParseImageSourceOverrides_RejectsMissingSeparator(t *testing.T) {
	_, err := ParseImageSourceOverrides([]string{"api"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image:path")
	assert.Contains(t, err.Error(), `"api"`)
}

func TestParseImageSourceOverrides_RejectsEmptyImage(t *testing.T) {
	sourceDirectory := t.TempDir()

	_, err := ParseImageSourceOverrides([]string{":" + sourceDirectory})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image name cannot be empty")
}

func TestParseImageSourceOverrides_RejectsEmptyPath(t *testing.T) {
	_, err := ParseImageSourceOverrides([]string{"api:"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source path cannot be empty")
}

func TestParseImageSourceOverrides_RejectsMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := ParseImageSourceOverrides([]string{"api:" + missing})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestParseImageSourceOverrides_RejectsControlCharacterInPath(t *testing.T) {
	_, err := ParseImageSourceOverrides([]string{"api:\x1b[31m/src"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "control character")
}

func TestParseImageSourceOverrides_RejectsFileInsteadOfDirectory(t *testing.T) {
	tempDirectory := t.TempDir()
	filePath := filepath.Join(tempDirectory, "Dockerfile")
	require.NoError(t, os.WriteFile(filePath, []byte("FROM scratch"), 0o600))

	_, err := ParseImageSourceOverrides([]string{"api:" + filePath})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestParseImageSourceOverrides_LastWinsOnDuplicateImage(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()

	overrides, err := ParseImageSourceOverrides([]string{"api:" + first, "api:" + second})
	require.NoError(t, err)

	path, ok := overrides.LookupSourcePath("api")
	require.True(t, ok)
	assert.Equal(t, second, path)
}

func TestImageSourceOverrideImageCompletions_OffersImagesWithColonSuffix(t *testing.T) {
	services := []domain.Service{
		{Name: "svc-a", DockerImages: []domain.DockerImage{{Name: "image-a"}, {Name: "image-b"}}},
		{Name: "svc-b", DockerImages: []domain.DockerImage{{Name: "image-a"}, {Name: "image-c"}}},
	}

	completions := imageSourceOverrideImageCompletions(services)

	assert.ElementsMatch(t, []cobra.Completion{"image-a:", "image-b:", "image-c:"}, completions)
}

func TestHelmChartOverrideServiceCompletions_OffersServicesWithColonSuffix(t *testing.T) {
	services := []domain.Service{{Name: "api"}, {Name: "worker"}}

	completions := helmChartOverrideServiceCompletions(services)

	assert.Equal(t, []cobra.Completion{"api:", "worker:"}, completions)
}

func TestHelmChartOverrideServiceCompletions_EmptyServicesYieldsEmptyCompletions(t *testing.T) {
	completions := helmChartOverrideServiceCompletions(nil)

	assert.Empty(t, completions)
}

func TestHelmChartOverrideDirectoryCompletions_ListsMatchingSubdirectoriesWithServicePrefix(t *testing.T) {
	rootDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "charts"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "chimera"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "other"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(rootDirectory, "chart.yaml"), []byte("x"), 0o600))

	completions := helmChartOverrideDirectoryCompletions("api:", rootDirectory+string(filepath.Separator)+"ch")

	separator := string(filepath.Separator)
	require.Len(t, completions, 2)
	assert.ElementsMatch(t, []cobra.Completion{
		"api:" + rootDirectory + separator + "charts" + separator,
		"api:" + rootDirectory + separator + "chimera" + separator,
	}, completions)
}

func TestHelmChartOverrideDirectoryCompletions_ListsCurrentDirectoryWhenNoPath(t *testing.T) {
	rootDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "charts"), 0o750))
	t.Chdir(rootDirectory)

	completions := helmChartOverrideDirectoryCompletions("api:", "")

	assert.Equal(t, []cobra.Completion{"api:charts" + string(filepath.Separator)}, completions)
}

func TestHelmChartOverrideDirectoryCompletions_ReturnsNilWhenDirectoryDoesNotExist(t *testing.T) {
	missingDirectory := filepath.Join(t.TempDir(), "does-not-exist") + string(filepath.Separator)

	completions := helmChartOverrideDirectoryCompletions("api:", missingDirectory)

	assert.Nil(t, completions)
}

func TestHelmChartOverrideDirectoryCompletions_ReturnsEmptyWhenNoEntriesMatchPrefix(t *testing.T) {
	rootDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "charts"), 0o750))

	completions := helmChartOverrideDirectoryCompletions("api:", rootDirectory+string(filepath.Separator)+"zz")

	assert.NotNil(t, completions)
	assert.Empty(t, completions)
}

func TestHelmChartOverrideDirectoryCompletions_FollowsSymlinksToDirectories(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Symlink requires elevated privileges on Windows")
	}
	rootDirectory := t.TempDir()
	realChartDirectory := filepath.Join(rootDirectory, "real-charts")
	require.NoError(t, os.Mkdir(realChartDirectory, 0o750))
	symlinkPath := filepath.Join(rootDirectory, "linked-charts")
	require.NoError(t, os.Symlink(realChartDirectory, symlinkPath))

	completions := helmChartOverrideDirectoryCompletions("api:", rootDirectory+string(filepath.Separator)+"linked")

	assert.Equal(t, []cobra.Completion{
		"api:" + rootDirectory + string(filepath.Separator) + "linked-charts" + string(filepath.Separator),
	}, completions)
}

func TestHelmChartOverrideDirectoryCompletions_HidesDotDirectoriesUnlessPrefixStartsWithDot(t *testing.T) {
	rootDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, ".cache"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "charts"), 0o750))

	completionsWithoutDot := helmChartOverrideDirectoryCompletions("api:", rootDirectory+string(filepath.Separator))
	assert.Equal(t, []cobra.Completion{
		"api:" + rootDirectory + string(filepath.Separator) + "charts" + string(filepath.Separator),
	}, completionsWithoutDot)

	completionsWithDot := helmChartOverrideDirectoryCompletions("api:", rootDirectory+string(filepath.Separator)+".")
	assert.Equal(t, []cobra.Completion{
		"api:" + rootDirectory + string(filepath.Separator) + ".cache" + string(filepath.Separator),
	}, completionsWithDot)
}

func TestHelmChartOverrideDirectoryCompletions_ExpandsLeadingTildeSlash(t *testing.T) {
	homeDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDirectory, "charts"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(homeDirectory, "chimera"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(homeDirectory, "other"), 0o750))
	t.Setenv("HOME", homeDirectory)
	t.Setenv("USERPROFILE", homeDirectory)

	completions := helmChartOverrideDirectoryCompletions("api:", "~/ch")

	separator := string(filepath.Separator)
	assert.ElementsMatch(t, []cobra.Completion{
		"api:~/charts" + separator,
		"api:~/chimera" + separator,
	}, completions)
}

func TestHelmChartOverrideDirectoryCompletions_ExpandsBareTilde(t *testing.T) {
	homeDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(homeDirectory, "charts"), 0o750))
	t.Setenv("HOME", homeDirectory)
	t.Setenv("USERPROFILE", homeDirectory)

	completions := helmChartOverrideDirectoryCompletions("api:", "~")

	separator := string(filepath.Separator)
	assert.Equal(t, []cobra.Completion{
		"api:~" + separator + "charts" + separator,
	}, completions)
}

func TestHelmChartOverrideDirectoryCompletions_LeavesNonLeadingTildeLiteral(t *testing.T) {
	rootDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "~tilde"), 0o750))

	completions := helmChartOverrideDirectoryCompletions("api:", rootDirectory+string(filepath.Separator)+"~")

	separator := string(filepath.Separator)
	assert.Equal(t, []cobra.Completion{
		"api:" + rootDirectory + separator + "~tilde" + separator,
	}, completions)
}

func TestParseHelmChartOverrides_ExpandsLeadingTilde(t *testing.T) {
	homeDirectory := t.TempDir()
	chartDirectoryName := "charts"
	require.NoError(t, os.Mkdir(filepath.Join(homeDirectory, chartDirectoryName), 0o750))
	t.Setenv("HOME", homeDirectory)
	t.Setenv("USERPROFILE", homeDirectory)

	overrides, err := ParseHelmChartOverrides([]string{"api:~/" + chartDirectoryName})
	require.NoError(t, err)

	directory, ok := overrides.LookupChartDirectory("api")
	require.True(t, ok)
	assert.Equal(t, filepath.Join(homeDirectory, chartDirectoryName), directory)
}

func TestHelmChartOverrideCompletion_ReturnsDirectoryCompletionsAfterSeparator(t *testing.T) {
	rootDirectory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(rootDirectory, "charts"), 0o750))

	completions, directive := HelmChartOverrideCompletion(nil, nil, "api:"+rootDirectory+string(filepath.Separator))

	assert.Equal(t, []cobra.Completion{
		"api:" + rootDirectory + string(filepath.Separator) + "charts" + string(filepath.Separator),
	}, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoSpace, directive)
}

func TestAnchorBareVolumeToRoot(t *testing.T) {
	cases := []struct {
		name          string
		path          string
		volumeName    string
		pathSeparator string
		expected      string
	}{
		{
			name:          "windows bare drive letter is anchored to drive root",
			path:          "C:",
			volumeName:    "C:",
			pathSeparator: `\`,
			expected:      `C:\`,
		},
		{
			name:          "windows UNC volume is anchored to share root",
			path:          `\\server\share`,
			volumeName:    `\\server\share`,
			pathSeparator: `\`,
			expected:      `\\server\share\`,
		},
		{
			name:          "windows drive-rooted path is left unchanged",
			path:          `C:\charts`,
			volumeName:    "C:",
			pathSeparator: `\`,
			expected:      `C:\charts`,
		},
		{
			name:          "unix path is left unchanged",
			path:          "/charts",
			volumeName:    "",
			pathSeparator: "/",
			expected:      "/charts",
		},
		{
			name:          "empty path is left unchanged",
			path:          "",
			volumeName:    "",
			pathSeparator: "/",
			expected:      "",
		},
		{
			name:          "tilde-expanded path is left unchanged",
			path:          "~/charts",
			volumeName:    "",
			pathSeparator: "/",
			expected:      "~/charts",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := anchorBareVolumeToRoot(testCase.path, testCase.volumeName, testCase.pathSeparator)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
