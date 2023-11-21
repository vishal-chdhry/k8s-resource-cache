package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "get external api response from cache or source",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("too many arguments to 'get'")
		}
		return getExternal()
	},
}

func getExternal() error {
	urlPrompt := promptContent{
		errorMsg: "Please provide a url.",
		label:    "URL:",
	}
	url := getInputPromptUi(urlPrompt)

	caBundlePrompt := promptContent{
		errorMsg: "Please provide a ca bundle file.",
		label:    "CABundle:",
	}
	caBundleFile := getInputPromptUi(caBundlePrompt)

	refreshIntervalPrompt := promptContent{
		errorMsg: "Please provide a namespace.",
		label:    "RefreshInterval:",
	}
	refreshIntervalStr := getInputPromptUi(refreshIntervalPrompt)

	refreshInterval, err := strconv.ParseInt(refreshIntervalStr, 10, 32)
	if err != nil {
		panic(err)
	}

	caBundle, err := os.ReadFile(caBundleFile)
	if err != nil {
		panic(err)
	}

	for {
		start := time.Now()
		data, err := resourceCache.GetExternalData(url, string(caBundle), int(refreshInterval))
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, "resources found from url:", url, "body:", data)
		fmt.Fprintln(os.Stdout, "Time taken: ", fmt.Sprint(time.Since(start).Microseconds())+"μs")

		time.Sleep(time.Second * 1)
	}
}
