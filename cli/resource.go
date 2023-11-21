package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type promptContent struct {
	errorMsg string
	label    string
}

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "get resources from cache or source",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("too many arguments to 'get'")
		}
		return getResource()
	},
}

func getInputPromptUi(pc promptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     pc.label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func getResource() error {
	groupPrompt := promptContent{
		errorMsg: "Please provide a group.",
		label:    "Group:",
	}
	group := getInputPromptUi(groupPrompt)

	versionPrompt := promptContent{
		errorMsg: "Please provide a version.",
		label:    "Version:",
	}
	version := getInputPromptUi(versionPrompt)

	kindPrompt := promptContent{
		errorMsg: "Please provide a kind.",
		label:    "Kind:",
	}
	kind := getInputPromptUi(kindPrompt)

	namespacePrompt := promptContent{
		errorMsg: "Please provide a namespace.",
		label:    "Namespace:",
	}
	namespace := getInputPromptUi(namespacePrompt)

	resource := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
	}

	for {
		start := time.Now()
		lister, err := resourceCache.GetLister(resource, namespace)
		if err != nil {
			return err
		}

		ret, err := lister.List(labels.Everything())
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, len(ret), "resources found of type:", resource.String())
		fmt.Fprintln(os.Stdout, "Time taken: ", fmt.Sprint(time.Since(start).Microseconds())+"μs")

		time.Sleep(time.Second * 1)
	}
}
