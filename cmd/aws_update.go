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

	"github.com/mateimicu/kdiscover/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	backupKubeconfig bool
	alias            string
)

func newUpdateCommand() *cobra.Command {
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update all EKS Clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(cmd.Short)
			remoteEKSClusters := internal.GetEKSClusters(awsRegions)
			log.Info(remoteEKSClusters)
			fmt.Printf("Found %v clusters remote\n", len(remoteEKSClusters))
			if backupKubeconfig && fileExists(kubeconfigPath) {
				bName, err := generateBackupName(kubeconfigPath)
				if err != nil {
					log.WithFields(log.Fields{
						"kubeconfig-path": kubeconfigPath,
					}).Info("Can't generate backup file name ")
				}
				fmt.Printf("Backup kubeconfig to %v\n", bName)
				err = copy(kubeconfigPath, bName)
				if err != nil {
					return err
				}
			}
			err := internal.UpdateKubeconfig(remoteEKSClusters, kubeconfigPath, contextName{templateValue: alias})
			if err != nil {
				return err
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

type contextName struct {
	templateValue string
}

func (c contextName) GetContextName(cls internal.Cluster) (string, error) {
	tmpl, err := template.New("context-name").Parse(c.templateValue)
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
