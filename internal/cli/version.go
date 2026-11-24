package cli

import (
	"runtime"

	"github.com/spf13/cobra"
)

var version = "dev"
var commit = "unknown"
var buildDate = "unknown"

func versionCommand() *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Show build identity and supported formats", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		return emit(cmd, struct {
			Version       string `json:"version"`
			Commit        string `json:"commit"`
			BuildDate     string `json:"buildDate"`
			GoVersion     string `json:"goVersion"`
			Platform      string `json:"platform"`
			ConfigVersion string `json:"configVersion"`
			BundleVersion int    `json:"bundleVersion"`
		}{version, commit, buildDate, runtime.Version(), runtime.GOOS + "/" + runtime.GOARCH, "v1alpha1", 1})
	}}
}
