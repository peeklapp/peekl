package code

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/klauspost/compress/zstd"
	"github.com/peeklapp/peekl/internal/inventory"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

func createArchive(path string) (*tar.Writer, *zstd.Encoder, *os.File, error) {
	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	zstdWriter, err := zstd.NewWriter(f)
	if err != nil {
		return nil, nil, f, err
	}

	archiveWriter := tar.NewWriter(zstdWriter)

	return archiveWriter, zstdWriter, f, nil
}

func addFileToArchive(rootDirectory string, path string, archiveWriter *tar.Writer, info os.FileInfo) error {
	relPath, err := filepath.Rel(rootDirectory, path)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = relPath
	if err := archiveWriter.WriteHeader(header); err != nil {
		return err
	}

	fileData, err := os.Open(path)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(fileData)

	_, err = io.Copy(archiveWriter, fileData)
	if err != nil {
		return err
	}

	return nil
}

func addDirectoryToArchive(rootDirectory string, archiveWriter *tar.Writer, directoryRegex *regexp.Regexp) error {
	return filepath.Walk(rootDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootDirectory, path)
		if err != nil {
			return err
		}

		if directoryRegex != nil && !directoryRegex.Match([]byte(relPath)) {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath
		if err := archiveWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			fileData, err := os.Open(path)
			if err != nil {
				return err
			}
			defer utils.CloseWithoutError(fileData)

			if _, err := io.Copy(archiveWriter, fileData); err != nil {
				return err
			}
		}

		return nil
	})
}

func GenerateCodeArchive(codeFolder string, outputFolder string) error {
	logrus.Debug("Creating output folder if it doesn't exist")
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return err
	}

	logrus.Debug("Creating empty tarball")
	archiveWriter, zstdWriter, file, err := createArchive(filepath.Join(outputFolder, CodeTarballName))
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(file)
	defer utils.CloseWithoutError(zstdWriter)
	defer utils.CloseWithoutError(archiveWriter)

	// Only roles are shared by every node. Inventory groups and their
	// variables are shipped in each node archive, for the groups that
	// node belongs to (see generateNodeArchive).
	logrus.Debug("Generating global tarball")
	includedDirRegex := regexp.MustCompile("^roles/.+")
	if err := addDirectoryToArchive(codeFolder, archiveWriter, includedDirRegex); err != nil {
		return err
	}

	return closeArchive(archiveWriter, zstdWriter, file)
}

// closeArchive flushes and closes the archive writers in order, so that
// a failed flush is reported instead of leaving a truncated archive.
func closeArchive(archiveWriter *tar.Writer, zstdWriter *zstd.Encoder, file *os.File) error {
	if err := archiveWriter.Close(); err != nil {
		return err
	}
	if err := zstdWriter.Close(); err != nil {
		return err
	}
	return file.Close()
}

// entityDirectoryRegex matches everything under `<prefix>/<name>/`, and
// nothing under a directory that merely starts with name.
func entityDirectoryRegex(prefix string, name string) *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.QuoteMeta(prefix+"/"+name) + "(/|$)")
}

// addEntityToArchive adds an inventory file and, when it exists, the
// matching variables directory (e.g. `variables/nodes/<name>`).
func addEntityToArchive(codeFolder string, archiveWriter *tar.Writer, sourcePath string, variablesPrefix string, name string) error {
	sourceFullPath := filepath.Join(codeFolder, sourcePath)
	sourceStat, err := os.Stat(sourceFullPath)
	if err != nil {
		return err
	}
	if err := addFileToArchive(codeFolder, sourceFullPath, archiveWriter, sourceStat); err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(codeFolder, variablesPrefix, name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return addDirectoryToArchive(codeFolder, archiveWriter, entityDirectoryRegex(variablesPrefix, name))
}

func generateNodeArchive(
	codeFolder string,
	outputFolder string,
	node models.NodeInventory,
	groups map[string]models.GroupInventory,
) error {
	archiveWriter, zstdWriter, file, err := createArchive(filepath.Join(outputFolder, node.Name+TarballExtension))
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(file)
	defer utils.CloseWithoutError(zstdWriter)
	defer utils.CloseWithoutError(archiveWriter)

	if err := addEntityToArchive(codeFolder, archiveWriter, node.SourcePath, "variables/nodes", node.Name); err != nil {
		return err
	}

	for _, groupName := range node.Groups {
		group, present := groups[groupName]
		if !present {
			return fmt.Errorf("node '%s' : %w", node.Name, models.GroupNotFoundError{GroupName: groupName})
		}
		if err := addEntityToArchive(codeFolder, archiveWriter, group.SourcePath, "variables/groups", group.Name); err != nil {
			return err
		}
	}

	return closeArchive(archiveWriter, zstdWriter, file)
}

func GenerateNodesArchives(codeFolder string, outputFolder string) error {
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return err
	}

	logrus.Debug("Open code in root folder")
	codeRoot, err := os.OpenRoot(codeFolder)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(codeRoot)

	logrus.Debug("Loading entire inventory")
	nodes, groups, err := inventory.LoadInventory(codeRoot)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := generateNodeArchive(codeFolder, outputFolder, node, groups); err != nil {
			return fmt.Errorf("could not generate archive for node '%s' : %w", node.Name, err)
		}
	}

	return nil
}

func DecompressArchive(tarballPath string, outputDir string) error {
	tarball, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("could not read tarball due the following error : %s", err.Error())
	}
	defer utils.CloseWithoutError(tarball)

	zstdReader, err := zstd.NewReader(tarball)
	if err != nil {
		return fmt.Errorf("could not create a Zstd reader due to the following error : %s", err.Error())
	}
	defer zstdReader.Close()

	tarReader := tar.NewReader(zstdReader)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("could not create output dir due to the following error : %s", err.Error())
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("an error happened while to read tar entry : %s", err.Error())
		}

		target := filepath.Join(outputDir, filepath.Clean("/"+header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("could not create directory '%s' locally : %s", target, err.Error())
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("could not create parent directory '%s' locally : %s", target, err.Error())
			}

			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("an error happened while trying to create file '%s' locally : %s", target, err.Error())
			}
			defer utils.CloseWithoutError(out)

			if _, err := io.Copy(out, tarReader); err != nil {
				return fmt.Errorf("an error happened while writing to file '%s' : %s", target, err.Error())
			}
		default:
			return fmt.Errorf("unsupported tar entry type '%v' for file '%s'", header.Typeflag, header.Name)
		}
	}

	return err
}
