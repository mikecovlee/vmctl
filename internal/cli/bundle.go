package cli

import (
	"github.com/spf13/cobra"
	"vmctl/internal/vm"
)

func (o *options) addBundleCommands(root *cobra.Command) {
	root.AddCommand(&cobra.Command{Use: "export NAME ARCHIVE", Short: "Export a stopped instance, including credentials", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		m, err := o.manager()
		if err != nil {
			return err
		}
		return m.Export(cmd.Context(), args[0], args[1])
	}})
	var configuration string
	command := &cobra.Command{Use: "import ARCHIVE NAME", Short: "Import persistent guest state as a stopped instance", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		m, err := o.manager()
		if err != nil {
			return err
		}
		record, err := m.Import(cmd.Context(), vm.ImportOptions{Archive: args[0], Name: args[1], Config: configuration})
		if err != nil {
			return err
		}
		return emit(cmd, record)
	}}
	command.Flags().StringVar(&configuration, "config", "", "Configuration supplying local media paths, networking and console ports")
	root.AddCommand(command)
}
