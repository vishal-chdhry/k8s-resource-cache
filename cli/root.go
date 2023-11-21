package cli

import (
	"github.com/spf13/cobra"
	"github.com/vishal-chdhry/k8s-resource-cache/pkg/resourcecache"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
)

var (
	rootCmd       *cobra.Command
	namespace     string
	allNamespaces bool
	mapper        meta.RESTMapper
	client        *dynamic.DynamicClient
	resourceCache resourcecache.ResourceCache
)

func init() {
	kubecfgFlags := genericclioptions.NewConfigFlags(false)

	rootCmd = &cobra.Command{
		Use:   "kubectl-resourcecache",
		Short: "demo of caching kubernetes resources with dynamic informers",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			config, err := kubecfgFlags.ToRESTConfig()
			if err != nil {
				return err
			}

			mapper, err = kubecfgFlags.ToRESTMapper()
			if err != nil {
				return err
			}

			client, err = dynamic.NewForConfig(config)
			if err != nil {
				return err
			}

			resourceCache, err = resourcecache.NewResourceCache(client)
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(resourceCmd)
	rootCmd.AddCommand(externalCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
