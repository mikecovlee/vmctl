package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
	"vmctl/internal/vm"
)

type options struct {
	directory  string
	image      string
	jsonOutput bool
}

func New() *cobra.Command {
	o := &options{}
	root := &cobra.Command{Use: "vmctl", Short: "Manage isolated Linux and Windows guests with Docker and KVM", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringVar(&o.directory, "state-dir", "", "Instance state directory")
	root.PersistentFlags().StringVar(&o.image, "image", "vmctl-runtime:local", "Runtime image")
	root.PersistentFlags().BoolVar(&o.jsonOutput, "json", false, "Emit JSON")
	root.AddCommand(&cobra.Command{Use: "validate CONFIG", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		spec, err := config.Read(args[0])
		if err != nil {
			return err
		}
		return emit(cmd, spec)
	}})
	root.AddCommand(&cobra.Command{Use: "create CONFIG", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		spec, err := config.Read(args[0])
		if err != nil {
			return err
		}
		manager, err := o.manager()
		if err != nil {
			return err
		}
		record, err := manager.Create(cmd.Context(), spec)
		if err != nil {
			return err
		}
		return emit(cmd, record)
	}})
	root.AddCommand(&cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := o.manager()
		if err != nil {
			return err
		}
		records, err := manager.List(cmd.Context())
		if err != nil {
			return err
		}
		return emit(cmd, records)
	}})
	for _, action := range []string{"start", "install", "finish-install", "status", "verify"} {
		var wait bool
		command := &cobra.Command{Use: action + " NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := o.manager()
			if err != nil {
				return err
			}
			switch cmd.Name() {
			case "start":
				err = manager.Start(cmd.Context(), args[0], false)
			case "install":
				err = manager.Start(cmd.Context(), args[0], true)
				if err == nil && wait {
					err = manager.WaitInstall(cmd.Context(), args[0])
				}
			case "finish-install":
				err = manager.FinishInstall(cmd.Context(), args[0])
			case "verify":
				err = manager.Verify(cmd.Context(), args[0])
			case "status":
				result, statusErr := manager.Status(cmd.Context(), args[0])
				if statusErr != nil {
					return statusErr
				}
				return emit(cmd, result)
			}
			if err != nil {
				return err
			}
			return emit(cmd, struct {
				Name   string `json:"name"`
				Action string `json:"action"`
			}{args[0], cmd.Name()})
		}}
		if action == "install" {
			command.Flags().BoolVar(&wait, "wait", false, "Wait up to 45 minutes for Kickstart installation shutdown")
		}
		root.AddCommand(command)
	}
	var force, yes bool
	stop := &cobra.Command{Use: "stop NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := o.manager()
		if err != nil {
			return err
		}
		return manager.Stop(cmd.Context(), args[0], force)
	}}
	stop.Flags().BoolVar(&force, "force", false, "Power off without guest shutdown")
	remove := &cobra.Command{Use: "delete NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := o.manager()
		if err != nil {
			return err
		}
		return manager.Delete(cmd.Context(), args[0], yes)
	}}
	remove.Flags().BoolVar(&yes, "yes", false, "Confirm permanent instance deletion")
	root.AddCommand(stop, remove)
	o.addRuntimeCommands(root)
	return root
}

func (o options) manager() (vm.Manager, error) {
	directory := o.directory
	if directory == "" {
		base := os.Getenv("XDG_STATE_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return vm.Manager{}, err
			}
			base = filepath.Join(home, ".local", "state")
		}
		directory = filepath.Join(base, "vmctl")
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return vm.Manager{}, fmt.Errorf("state directory: %w", err)
	}
	return vm.Manager{Store: state.Store{Root: absolute}, Docker: docker.Client{}, Image: o.image}, nil
}

func emit[T any](cmd *cobra.Command, value T) error {
	compact, err := cmd.Root().PersistentFlags().GetBool("json")
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	if !compact {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(value)
}
