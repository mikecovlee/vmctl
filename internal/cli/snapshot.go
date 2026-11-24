package cli

import (
	"time"

	"github.com/spf13/cobra"
	"vmctl/internal/vm"
)

func (o *options) addSnapshotCommands(root *cobra.Command) {
	for _, action := range []string{"checkpoint", "restore-checkpoint"} {
		action := action
		var yes bool
		var timeout time.Duration
		command := &cobra.Command{Use: action + " NAME LABEL", Short: "Save or restore BIOS guest memory and device state (experimental)", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			m, err := o.manager()
			if err != nil {
				return err
			}
			if !yes {
				return vm.ErrConfirmation
			}
			return m.CheckpointWithTimeout(cmd.Context(), args[0], args[1], action == "restore-checkpoint", timeout)
		}}
		command.Flags().BoolVar(&yes, "yes", false, "Confirm checkpoint write or rollback; an existing checkpoint label may be replaced")
		command.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Maximum time to wait for QEMU save or restore")
		root.AddCommand(command)
	}
	root.AddCommand(&cobra.Command{Use: "snapshots NAME", Short: "List saved snapshots", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		m, err := o.manager()
		if err != nil {
			return err
		}
		labels, err := m.Snapshots(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return emit(cmd, labels)
	}})
	var yes bool
	restore := &cobra.Command{Use: "restore NAME LABEL", Short: "Restore stopped guest persistent state", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		m, err := o.manager()
		if err != nil {
			return err
		}
		return m.Restore(cmd.Context(), args[0], args[1], yes)
	}}
	restore.Flags().BoolVar(&yes, "yes", false, "Confirm replacing current guest state")
	root.AddCommand(restore)
	root.AddCommand(&cobra.Command{Use: "snapshot NAME LABEL", Short: "Save stopped guest disk, NVRAM and TPM state", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		m, err := o.manager()
		if err != nil {
			return err
		}
		return m.Snapshot(cmd.Context(), args[0], args[1])
	}})
}
