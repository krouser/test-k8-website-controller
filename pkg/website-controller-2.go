// write Kubernetes controller, which watches the Kubernetes API server for website objects and runs an Nginx webserver for each of them.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.k8s.io/apimachinery/pkg/api/errors"

	"github.com/website-operator/pkg/apis/website/v1alpha1"
	"github.com/website-operator/pkg/controller/util"
)

// WebsiteController watches for Website objects and creates Nginx servers for each of them.
type WebsiteController struct {
	log logr.Logger
}

// NewWebsiteController creates a new WebsiteController.
func NewWebsiteController(log logr.Logger) *WebsiteController {
	return &WebsiteController{log: log}
}

// Run starts the WebsiteController.
func (c *WebsiteController) Run(ctx context.Context) error {
	// Watch for Website objects
	err := c.watch(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to watch for Website objects")
	}

	return nil
}

// watch watches for Website objects.
func (c *WebsiteController) watch(ctx context.Context) error {
	// Create a new Website object
	watch := util.NewWatch(ctx, &v1alpha1.Website{})

	// Watch for Website objects
	err := watch.Watch(func(event watch.Event) error {
		// Handle the event
		return c.handleEvent(event)
	})
	if err != nil {
		return errors.Wrap(err, "failed to watch for Website objects")
	}

	return nil
}

// handleEvent handles a watch event.
func (c *WebsiteController) handleEvent(event watch.Event) error {
	// Get the Website object
	website, ok := event.Object.(*v1alpha1.Website)
	if !ok {
		return errors.Errorf("object is not a Website: %T", event.Object)
	}

	// Handle the event type
	switch event.Type {
	case watch.Added:
		return c.handleAdded(website)
	case watch.Modified:
		return c.handleModified(website)
	case watch.Deleted:
		return c.handleDeleted(website)
	}

	return nil
}

// handleAdded handles an added Website object.
func (c *WebsiteController) handleAdded(website *v1alpha1.Website) error {
	// Create the Nginx server
	err := c.createNginxServer(website)
	if err != nil {
		return errors.Wrap(err, "failed to create Nginx server")
	}

	return nil
}

// handleModified handles a modified Website object.
func (c *WebsiteController) handleModified(website *v1alpha1.Website) error {
	// Update the Nginx server
	err := c.updateNginxServer(website)
	if err != nil {
		return errors.Wrap(err, "failed to update Nginx server")
	}

	return nil
}

// handleDeleted handles a deleted Website object.
func (c *WebsiteController) handleDeleted(website *v1alpha1.Website) error {
	// Delete the Nginx server
	err := c.deleteNginxServer(website)
	if err != nil {
		return errors.Wrap(err, "failed to delete Nginx server")
	}

	return nil
}

// createNginxServer creates an Nginx server for a Website object.
func (c *WebsiteController) createNginxServer(website *v1alpha1.Website) error {
	// Create the Nginx configuration
	config := c.createNginxConfig(website)

	// Write the Nginx configuration to a file
	configPath := filepath.Join("/etc/nginx/conf.d", fmt.Sprintf("%s.conf", website.Name))
	err := os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write Nginx configuration")
	}

	// Reload the Nginx configuration
	err = c.reloadNginx()
	if err != nil {
		return errors.Wrap(err, "failed to reload Nginx configuration")
	}

	return nil
}

// updateNginxServer updates an Nginx server for a Website object.
func (c *WebsiteController) updateNginxServer(website *v1alpha1.Website) error {
	// Create the Nginx configuration
	config := c.createNginxConfig(website)

	// Write the Nginx configuration to a file
	configPath := filepath.Join("/etc/nginx/conf.d", fmt.Sprintf("%s.conf", website.Name))
	err := os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write Nginx configuration")
	}

	// Reload the Nginx configuration
	err = c.reloadNginx()
	if err != nil {
		return errors.Wrap(err, "failed to reload Nginx configuration")
	}

	return nil
}

// deleteNginxServer deletes an Nginx server for a Website object.
func (c *WebsiteController) deleteNginxServer(website *v1alpha1.Website) error {
	// Delete the Nginx configuration file
	configPath := filepath.Join("/etc/nginx/conf.d", fmt.Sprintf("%s.conf", website.Name))
	err := os.Remove(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to delete Nginx configuration")
	}

	// Reload the Nginx configuration
	err = c.reloadNginx()
	if err != nil {
		return errors.Wrap(err, "failed to reload Nginx configuration")
	}

	return nil
}

// createNginxConfig creates an Nginx configuration for a Website object.
func (c *WebsiteController) createNginxConfig(website *v1alpha1.Website) string {
	return fmt.Sprintf(`
server {
	listen 80;
	server_name %s;
	location / {
		proxy_pass %s;
	}
}
`, website.Spec.Hostname, website.Spec.Upstream)
}

// reloadNginx reloads the Nginx configuration.
func (c *WebsiteController) reloadNginx() error {
	// Reload the Nginx configuration
	cmd := exec.Command("nginx", "-s", "reload")
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to reload Nginx configuration")
	}

	return nil
}
