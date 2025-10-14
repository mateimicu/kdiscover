// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/mateimicu/kdiscover/internal/aws"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	backupKubeconfig bool
)

func backupKubeConfig(kubeconfigPath string) (string, error) {
	bName, err := generateBackupName(kubeconfigPath)
	if err != nil {
		log.WithFields(log.Fields{
			"kubeconfig-path": kubeconfigPath,
			"err":             err.Error(),
		}).Info("Can't generate backup file name ")
	}
	err = copyFs(kubeconfigPath, bName)
	if err != nil {
		return "", err
	}
	return bName, nil
}

func newUpdateCommand() *cobra.Command {
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update all EKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(cmd.Short)

			remoteEKSClusters := aws.GetEKSClusters(awsRegions)
			log.Info(remoteEKSClusters)

			cmd.Printf("Found %v clusters remote\n", len(remoteEKSClusters))

			kubeconfigPath := GetKubeconfigPath()
			if backupKubeconfig && fileExists(kubeconfigPath) {
				bName, err := backupKubeConfig(kubeconfigPath)
				if err != nil {
					return err
				}
				cmd.Printf("Backup kubeconfig to %v\n", bName)
			}
			kubeconfig, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			if err != nil {
				return err
			}

			for _, cls := range remoteEKSClusters {
				ctxName, err := cls.PrettyName(alias)
				if err != nil {
					log.WithFields(log.Fields{
						"cluster": cls,
						"error":   err,
					}).Info("Can't generate alias for the cluster")
					continue
				}
				kubeconfig.AddCluster(cls, ctxName)
			}
			err = kubeconfig.Persist(kubeconfigPath)
			if err != nil {
				cmd.Printf("Failed to persist kubeconfig %v", err.Error())
			}
			return err
		},
	}

	updateCommand.Flags().BoolVar(&backupKubeconfig, "backup-kubeconfig", true, "Backup cubeconfig before update")

	return updateCommand
}

func copyFs(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("file %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		return err
	}

	//nolint: gomnd
	buf := make([]byte, 1000000)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func generateBackupName(origin string) (string, error) {
	fi, err := os.Stat(origin)
	if err != nil {
		return "", err
	}
	if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", origin)
	}
	oName := path.Base(origin)
	oDir := path.Dir(origin)
	for {
		if fileExists(path.Join(oDir, oName)) {
			oName += ".bak"
		} else {
			break
		}
	}
	return path.Join(oDir, oName), nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
