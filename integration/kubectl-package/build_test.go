/* #nosec */

package kubectlpackage

import (
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("build subcommand",
	testSubCommand("build"),
	Entry("given no path",
		subCommandTestCase{
			ExpectedExitCode: 1,
		},
	),
	Entry("given an invalid path",
		subCommandTestCase{
			Args:             []string{"dne"},
			ExpectedExitCode: 1,
		},
	),
	Entry("given the path of a valid package",
		subCommandTestCase{
			Args:             []string{sourcePathFixture("valid_without_config")},
			ExpectedExitCode: 0,
		},
	),
	Entry("given the path of a package with an invalid manifest",
		subCommandTestCase{
			Args:                []string{sourcePathFixture("invalid_bad_manifest")},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"spec.scopes: Required value"},
		},
	),
	Entry("given the path of a package with images, but no lock file",
		subCommandTestCase{
			Args:                []string{sourcePathFixture("invalid_missing_lock_file")},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"manifest.lock.yaml is missing"},
		},
	),
	// TODO: Add test registry and fixture with stale lock file
	// Entry("given the path of a package with images, but lock file is stale",
	// 	TestCase{
	// 		Args:                []string{filepath.Join("testdata", "")},
	// 		ExpectedExitCode:    1,
	// 		ExpectedErrorOutput: []string{""},
	// 	},
	// ),
	// TODO: Add test registry and fixture with lock file containing bad digests
	// Entry("given the path of a package with images, but lock file has invalid image ref(s)",
	// 	TestCase{
	// 		Args:                []string{filepath.Join("testdata", "")},
	// 		ExpectedExitCode:    1,
	// 		ExpectedErrorOutput: []string{""},
	// 	},
	// ),
	Entry("given the path of a package with images, but no lock file",
		subCommandTestCase{
			Args:                []string{sourcePathFixture("invalid_missing_lock_file")},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"manifest.lock.yaml is missing"},
		},
	),
	Entry("given '--output' without tags",
		subCommandTestCase{
			Args: []string{
				"--output", filepath.Join("dne", "dne"),
				sourcePathFixture("valid_without_config"),
			},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"output or push is requested but no tags are set"},
		},
	),
	Entry("given '--output' with an invalid path",
		subCommandTestCase{
			Args: []string{
				"--output", filepath.Join("dne", "dne"),
				"--tag", "test",
				sourcePathFixture("valid_without_config"),
			},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"no such file or directory"},
		},
	),
	Entry("given '--output' with an invalid tag",
		subCommandTestCase{
			Args: []string{
				"--output", filepath.Join("dne", "dne"),
				"--tag", "****",
				sourcePathFixture("valid_without_config"),
			},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"invalid tag specified as parameter"},
		},
	),
	Entry("given '--push' with no tags",
		subCommandTestCase{
			Args: []string{
				"--push",
				sourcePathFixture("valid_without_config"),
			},
			ExpectedExitCode:    1,
			ExpectedErrorOutput: []string{"output or push is requested but no tags are set"},
		},
	),
	Entry("given '--output' with valid path",
		subCommandTestCase{
			Args: []string{
				sourcePathFixture("valid_without_config"),
				"--output", filepath.Join(outputPath, "valid_build.tar"),
				"--tag", "valid-build",
			},
			ExpectedExitCode: 0,
			AdditionalValidations: func() {
				Expect(filepath.Join(outputPath, "valid_build.tar")).To(BeAnExistingFile())
			},
		},
	),
	Entry("given '--push' with valid tag",
		subCommandTestCase{
			Args: []string{
				"--push",
				"--insecure",
				"--tag", path.Join(registryPlaceholder, "valid-package"),
				sourcePathFixture("valid_without_config"),
			},
			ExpectedExitCode: 0,
			AdditionalValidations: func() {
				Expect(path.Join(_registryDomain, "valid-package")).To(ExistOnRegistry())
			},
		},
	),
)
