/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"context"
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps Helm action configuration and provides methods for chart operations
type Client struct {
	actionConfig *action.Configuration
	kubeClient   kubernetes.Interface
	restConfig   *rest.Config
	namespace    string
}

// NewClient creates a new Helm client
func NewClient(restConfig *rest.Config, kubeClient kubernetes.Interface, namespace string) (*Client, error) {
	actionConfig := new(action.Configuration)

	settings := cli.New()
	settings.SetNamespace(namespace)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", func(format string, v ...interface{}) {
		// Log function for Helm (can be customized)
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action config: %w", err)
	}

	return &Client{
		actionConfig: actionConfig,
		kubeClient:   kubeClient,
		restConfig:   restConfig,
		namespace:    namespace,
	}, nil
}

// InstallOrUpgrade installs or upgrades a Helm chart
func (c *Client) InstallOrUpgrade(ctx context.Context, releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
	// Check if release exists
	histClient := action.NewHistory(c.actionConfig)
	histClient.Max = 1
	_, err := histClient.Run(releaseName)

	exists := err == nil

	if exists {
		// Upgrade existing release
		return c.upgrade(ctx, releaseName, chartPath, values)
	}

	// Install new release
	return c.install(ctx, releaseName, chartPath, values)
}

// install installs a new Helm chart
func (c *Client) install(ctx context.Context, releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
	installAction := action.NewInstall(c.actionConfig)
	installAction.ReleaseName = releaseName
	installAction.Namespace = c.namespace
	installAction.CreateNamespace = true
	installAction.Timeout = 5 * time.Minute
	installAction.Wait = true
	installAction.WaitForJobs = true

	// Load chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Install
	rel, err := installAction.RunWithContext(ctx, chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart: %w", err)
	}

	return rel, nil
}

// upgrade upgrades an existing Helm chart
func (c *Client) upgrade(ctx context.Context, releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
	upgradeAction := action.NewUpgrade(c.actionConfig)
	upgradeAction.Namespace = c.namespace
	upgradeAction.Timeout = 5 * time.Minute
	upgradeAction.Wait = true
	upgradeAction.WaitForJobs = true

	// Load chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Upgrade
	rel, err := upgradeAction.RunWithContext(ctx, releaseName, chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade chart: %w", err)
	}

	return rel, nil
}

// Uninstall uninstalls a Helm release
func (c *Client) Uninstall(ctx context.Context, releaseName string) error {
	uninstallAction := action.NewUninstall(c.actionConfig)
	uninstallAction.Timeout = 5 * time.Minute
	uninstallAction.Wait = true

	_, err := uninstallAction.RunWithContext(ctx, releaseName)
	if err != nil {
		if errors.IsNotFound(err) {
			// Release not found, consider it already uninstalled
			return nil
		}
		return fmt.Errorf("failed to uninstall release: %w", err)
	}

	return nil
}

// GetReleaseStatus gets the status of a Helm release
func (c *Client) GetReleaseStatus(releaseName string) (string, error) {
	getAction := action.NewGetValues(c.actionConfig)
	_, err := getAction.Run(releaseName)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return "NotFound", nil
		}
		return "", err
	}

	statusAction := action.NewStatus(c.actionConfig)
	rel, err := statusAction.Run(releaseName)
	if err != nil {
		return "", err
	}

	return string(rel.Info.Status), nil
}

// ListReleases lists all Helm releases in the namespace
func (c *Client) ListReleases() ([]*release.Release, error) {
	listAction := action.NewList(c.actionConfig)
	listAction.AllNamespaces = false
	listAction.SetStateMask()

	releases, err := listAction.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	return releases, nil
}
