package code

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/sirupsen/logrus"
)

func createArchive(path string) (*tar.Writer, *zstd.Encoder, *os.File, error) {
	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
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
	defer fileData.Close()

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
			defer fileData.Close()

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
	defer file.Close()
	defer zstdWriter.Close()
	defer archiveWriter.Close()

	logrus.Debug("Generating global tarball")
	includedDirRegex := regexp.MustCompile("^(inventory/groups|roles|variables/groups).+")
	return addDirectoryToArchive(codeFolder, archiveWriter, includedDirRegex)
}

func GenerateNodesArchives(codeFolder, outputFolder string) error {
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return err
	}

	logrus.Debug("Getting list of node in the inventory")
	nodesInventoryFile, err := os.ReadDir(filepath.Join(codeFolder, "inventory/nodes"))
	if err != nil {
		return err
	}

	logrus.Debug("Starting process of generating tarball for each nodes")
	for _, nodeInvFile := range nodesInventoryFile {
		nodeFileInfo, err := nodeInvFile.Info()
		if err != nil {
			return err
		}

		cleanNodeName := strings.TrimSuffix(nodeInvFile.Name(), filepath.Ext(nodeInvFile.Name()))
		logrus.Debugf("Generating node tarball for node '%s'", cleanNodeName)

		archiveWriter, zstdWriter, file, err := createArchive(filepath.Join(outputFolder, cleanNodeName+TarballExtension))
		if err != nil {
			return err
		}
		defer file.Close()
		defer zstdWriter.Close()
		defer archiveWriter.Close()

		err = addFileToArchive(codeFolder, filepath.Join(codeFolder, "inventory/nodes", nodeInvFile.Name()), archiveWriter, nodeFileInfo)
		if err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Join(codeFolder, "variables", cleanNodeName)); os.IsNotExist(err) {
			continue
		}

		err = addDirectoryToArchive(
			codeFolder,
			archiveWriter,
			regexp.MustCompile(fmt.Sprintf("^variables/nodes/%s", cleanNodeName)),
		)
		if err != nil {
			return err
		}
		logrus.Debugf("Finished generating node tarball for node '%s'", cleanNodeName)
	}
	logrus.Debug("Finished process of generating tarball for each nodes")

	return nil
}

func DecompressArchive(tarballPath string, outputDir string) error {
	tarball, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("Could not read tarball due the following error : %s", err.Error())
	}
	defer tarball.Close()

	zstdReader, err := zstd.NewReader(tarball)
	if err != nil {
		return fmt.Errorf("Could not create a Zstd reader due to the following error : %s", err.Error())
	}
	defer zstdReader.Close()

	tarReader := tar.NewReader(zstdReader)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("Could not create output dir due to the following error : %s", err.Error())
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("An error happened while to read tar entry : %s", err.Error())
		}

		target := filepath.Join(outputDir, filepath.Clean("/"+header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("Could not create directory '%s' locally : %s", target, err.Error())
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("Could not create parent directory '%s' locally : %s", target, err.Error())
			}

			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("An error happened while trying to create file '%s' locally : %s", target, err.Error())
			}

			if _, err := io.Copy(out, tarReader); err != nil {
				out.Close()
				return fmt.Errorf("An error happened while writing to file '%s' : %s", target, err.Error())
			}
			out.Close()
		default:
			return fmt.Errorf("Unsupported tar entry type '%v' for file '%s'", header.Typeflag, header.Name)
		}
	}

	return nil
}
