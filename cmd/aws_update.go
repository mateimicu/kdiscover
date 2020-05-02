// Package cmd offers CLI functionality
package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/mateimicu/kdiscover/internal/aws"
	"github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	backupKubeconfig bool
	alias            string
)

func backupKubeConfig(kubeconfigPath string) (string, error) {
	bName, err := generateBackupName(kubeconfigPath)
	if err != nil {
		log.WithFields(log.Fields{
			"kubeconfig-path": kubeconfigPath,
			"err":             err.Error(),
		}).Info("Can't generate backup file name ")
	}
	err = copy(kubeconfigPath, bName)
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

			if backupKubeconfig && fileExists(kubeconfigPath) {
				bName, err := backupKubeConfig(kubeconfigPath)
				if err != nil {
					return err
				}
				cmd.Printf("Backup kubeconfig to %v\n", bName)
			}
			kubeconfig, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
			return nil
			if err != nil {
				return err
			}

			for _, cls := range remoteEKSClusters {
				ctxName, err := cls.GetContextName(alias)
				if err != nil {
					log.WithFields(log.Fields{
						"cluster": cls,
						"error":   err,
					}).Info("Can't generate alias for the cluster")
					continue
				}
				kubeconfig.AddCluster(&cls, ctxName)
			}
			return nil
		},
	}

	updateCommand.Flags().BoolVar(&backupKubeconfig, "backup-kubeconfig", true, "Backup cubeconfig before update")
	updateCommand.Flags().StringVar(
		&alias,
		"context-name-alias",
		"{{.Name}}",
		"Template for the context name. Has acces to Cluster type")

	return updateCommand
}

func GetContextName(cls cluster.Cluster, templateValue string) (string, error) {
	tmpl, err := template.New("context-name").Parse(templateValue)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, cls)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func copy(src, dst string) error {
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
