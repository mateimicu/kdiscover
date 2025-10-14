// Package gcp provides function for discovering GCP projects and zones
package gcp

import (
	"context"
	"os"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"golang.org/x/oauth2/google"
	log "github.com/sirupsen/logrus"
)

// GetProjectsAndZones gets all accessible projects and their zones
// If projects are provided, it uses those. Otherwise, it discovers all accessible projects.
func GetProjectsAndZones(projects []string) []ProjectZone {
	if len(projects) == 0 {
		projects = discoverProjects()
	}

	var projectZones []ProjectZone
	for _, project := range projects {
		zones := getZonesForProject(project)
		for _, zone := range zones {
			projectZones = append(projectZones, ProjectZone{
				ProjectID: project,
				Zone:      zone,
			})
		}
	}

	log.WithFields(log.Fields{
		"project-zones-count": len(projectZones),
		"projects-count":      len(projects),
	}).Info("Found project/zone combinations")

	return projectZones
}

// discoverProjects attempts to discover all accessible GCP projects
func discoverProjects() []string {
	ctx := context.Background()

	// Check if we have a specific project from environment
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		log.WithFields(log.Fields{
			"project": projectID,
		}).Info("Using project from GOOGLE_CLOUD_PROJECT environment variable")
		return []string{projectID}
	}

	// Try to use default credentials to discover projects
	creds, err := google.FindDefaultCredentials(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to find default credentials for project discovery")
		return []string{}
	}

	service, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to create resource manager service")
		return []string{}
	}

	resp, err := service.Projects.List().Do()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to list projects")
		return []string{}
	}

	var projects []string
	for _, project := range resp.Projects {
		if project.LifecycleState == "ACTIVE" {
			projects = append(projects, project.ProjectId)
		}
	}

	log.WithFields(log.Fields{
		"projects-count": len(projects),
	}).Info("Discovered projects")

	return projects
}

// getZonesForProject gets all zones for a specific project
func getZonesForProject(projectID string) []string {
	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		log.WithFields(log.Fields{
			"project": projectID,
			"error":   err.Error(),
		}).Error("Failed to find default credentials for zone discovery")
		return []string{}
	}

	service, err := compute.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.WithFields(log.Fields{
			"project": projectID,
			"error":   err.Error(),
		}).Error("Failed to create compute service")
		return []string{}
	}

	resp, err := service.Zones.List(projectID).Do()
	if err != nil {
		log.WithFields(log.Fields{
			"project": projectID,
			"error":   err.Error(),
		}).Error("Failed to list zones")
		return []string{}
	}

	var zones []string
	for _, zone := range resp.Items {
		// Filter only zones that are typically used for GKE
		if strings.Contains(zone.Name, "-a") || strings.Contains(zone.Name, "-b") || strings.Contains(zone.Name, "-c") ||
			zone.Name == "us-central1" || zone.Name == "us-east1" || zone.Name == "us-west1" ||
			zone.Name == "europe-west1" || zone.Name == "asia-east1" {
			zones = append(zones, zone.Name)
		}
	}

	log.WithFields(log.Fields{
		"project":     projectID,
		"zones-count": len(zones),
	}).Info("Found zones for project")

	return zones
}

// GetDefaultProjectsAndZones returns commonly used project/zone combinations
// This is a fallback when automatic discovery fails
func GetDefaultProjectsAndZones() []ProjectZone {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Warn("No GOOGLE_CLOUD_PROJECT environment variable set and auto-discovery failed")
		return []ProjectZone{}
	}

	// Return some common zones
	commonZones := []string{
		"us-central1-a", "us-central1-b", "us-central1-c",
		"us-east1-a", "us-east1-b", "us-east1-c",
		"us-west1-a", "us-west1-b", "us-west1-c",
		"europe-west1-a", "europe-west1-b", "europe-west1-c",
		"asia-east1-a", "asia-east1-b", "asia-east1-c",
	}

	var projectZones []ProjectZone
	for _, zone := range commonZones {
		projectZones = append(projectZones, ProjectZone{
			ProjectID: projectID,
			Zone:      zone,
		})
	}

	return projectZones
}