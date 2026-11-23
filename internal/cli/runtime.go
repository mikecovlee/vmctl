package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"vmctl/internal/docker"
	"vmctl/internal/vm"
)

func (o *options) addRuntimeCommands(root *cobra.Command) {
	var reveal bool
	console := &cobra.Command{Use: "console NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := o.manager()
		if err != nil {
			return err
		}
		instance, err := manager.Load(args[0])
		if err != nil {
			return err
		}
		result := struct {
			Address  string `json:"address"`
			Password string `json:"password,omitempty"`
		}{Address: fmt.Sprintf("%s:%d", instance.Spec.Console.VNCBind, instance.Spec.Console.VNCPort)}
		if reveal {
			data, err := os.ReadFile(filepath.Join(instance.Directory, "credentials.json"))
			if err != nil {
				return err
			}
			var credentials vm.Credentials
			if err := json.Unmarshal(data, &credentials); err != nil {
				return err
			}
			result.Password = credentials.VNCPassword
		}
		return emit(cmd, result)
	}}
	console.Flags().BoolVar(&reveal, "show-secret", false, "Reveal this instance's VNC password")
	logs := &cobra.Command{Use: "logs NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := o.manager()
		if err != nil {
			return err
		}
		instance, err := manager.Load(args[0])
		if err != nil {
			return err
		}
		container, err := manager.Docker.InspectInstance(cmd.Context(), instance.Directory)
		if err != nil {
			return err
		}
		data, err := manager.Docker.Run(cmd.Context(), "logs", "--tail", "200", container.ID)
		if err != nil {
			return err
		}
		_, err = cmd.OutOrStdout().Write(data)
		return err
	}}
	image := &cobra.Command{Use: "image"}
	image.AddCommand(&cobra.Command{Use: "build [RUNTIME_DIRECTORY]", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		directory := "runtime"
		if len(args) == 1 {
			directory = args[0]
		}
		client := docker.Client{Timeout: 15 * time.Minute}
		data, err := client.Run(cmd.Context(), "build", "--tag", o.image, directory)
		_, writeErr := cmd.OutOrStdout().Write(data)
		return errors.Join(err, writeErr)
	}})
	doctor := &cobra.Command{Use: "doctor", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		client := docker.Client{}
		if err := client.Doctor(cmd.Context(), o.image); err != nil {
			return err
		}
		return emit(cmd, struct {
			KVM    bool `json:"kvm"`
			Docker bool `json:"docker"`
		}{true, true})
	}}
	root.AddCommand(console, logs, image, doctor)
}
