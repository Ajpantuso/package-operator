package packagecontent

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	manifestsv1alpha1 "package-operator.run/apis/manifests/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"package-operator.run/internal/packages"
)

func PackageFromFiles(ctx context.Context, scheme *runtime.Scheme, files Files, component string) (pkg *Package, err error) {
	componentsEnabled, err := areComponentsEnabled(ctx, scheme, files)
	if err != nil {
		return nil, err
	}
	if !componentsEnabled {
		if component != "" {
			return nil, packages.ViolationError{Reason: packages.ViolationReasonComponentsNotEnabled}
		}
		return buildPackageFromFiles(ctx, scheme, files)
	}
	return buildPackageFromFiles(ctx, scheme, filterComponentFiles(files, component))
}

func buildPackageFromFiles(ctx context.Context, scheme *runtime.Scheme, files Files) (pkg *Package, err error) {
	pkg = &Package{nil, nil, map[string][]unstructured.Unstructured{}}
	for path, content := range files {
		switch {
		case strings.HasPrefix(filepath.Base(path), "_"):
			// skip template helper files.
			continue
		case !packages.IsYAMLFile(path):
			// skip non YAML files
			continue

		case packages.IsManifestFile(path):
			if pkg.PackageManifest != nil {
				err = packages.ViolationError{
					Reason: packages.ViolationReasonPackageManifestDuplicated,
					Path:   path,
				}

				return
			}
			pkg.PackageManifest, err = manifestFromFile(ctx, scheme, path, content)
			if err != nil {
				return nil, err
			}

			continue
		case packages.IsManifestLockFile(path):
			if pkg.PackageManifestLock != nil {
				err = packages.ViolationError{
					Reason: packages.ViolationReasonPackageManifestLockDuplicated,
					Path:   path,
				}

				return
			}
			pkg.PackageManifestLock, err = manifestLockFromFile(ctx, scheme, path, content)
			if err != nil {
				return nil, err
			}

			continue
		}

		// Trim empty starting and ending objects
		objects := []unstructured.Unstructured{}

		// Split for every included yaml document.
		for idx, yamlDocument := range bytes.Split(bytes.Trim(content, "---\n"), []byte("---\n")) {
			obj := unstructured.Unstructured{}
			if err = yaml.Unmarshal(yamlDocument, &obj); err != nil {
				err = packages.ViolationError{
					Reason:  packages.ViolationReasonInvalidYAML,
					Details: err.Error(),
					Path:    path,
					Index:   packages.Index(idx),
				}

				return
			}

			if len(obj.Object) != 0 {
				objects = append(objects, obj)
			}
		}
		if len(objects) != 0 {
			pkg.Objects[path] = objects
		}
	}

	if pkg.PackageManifest == nil {
		err = packages.ErrManifestNotFound
		return
	}

	return
}

func areComponentsEnabled(ctx context.Context, scheme *runtime.Scheme, files Files) (result bool, err error) {
	var manifest *manifestsv1alpha1.PackageManifest
	for path, content := range files {
		if packages.IsManifestFile(path) {
			if manifest != nil {
				return false, packages.ViolationError{
					Reason: packages.ViolationReasonPackageManifestDuplicated,
					Path:   path,
				}
			}
			manifest, err = manifestFromFile(ctx, scheme, path, content)
			if err != nil {
				return false, err
			}
		}
	}
	if manifest == nil {
		return false, packages.ErrManifestNotFound
	}
	return manifest.Spec.Component != nil, nil
}

func filterComponentFiles(files Files, component string) Files {
	var filtered Files = make(map[string][]byte)
	for path := range files {
		if isComponentFile(path, component) {
			filtered[path] = files[path]
		}
	}
	return filtered
}

func isComponentFile(path string, component string) bool {
	if component == "" {
		return !strings.HasPrefix(path, "components/")
	}
	return strings.HasPrefix(path, fmt.Sprintf("components/%s/", component))
}
