package cli

import (
	"github.com/spf13/cobra"
	"vmctl/internal/config"
	"vmctl/internal/vm"
)

func (o *options) addNetworkCommands(root *cobra.Command) {
	network := &cobra.Command{Use: "network", Short: "Manage stopped guest networking (TCP forwarding only)"}
	show := &cobra.Command{Use: "show NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		m, e := o.manager()
		if e != nil {
			return e
		}
		n, e := m.Network(cmd.Context(), args[0])
		if e != nil {
			return e
		}
		return emit(cmd, n)
	}}
	var mode string
	set := &cobra.Command{Use: "set NAME --mode isolated|nat", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		m, e := o.manager()
		if e != nil {
			return e
		}
		return m.ChangeNetwork(cmd.Context(), args[0], vm.NetworkChange{Mode: mode})
	}}
	set.Flags().StringVar(&mode, "mode", "", "Network mode (egress is a legacy alias for nat)")
	if err := set.MarkFlagRequired("mode"); err != nil {
		panic(err)
	}
	forward := &cobra.Command{Use: "forward", Short: "Manage incoming TCP forwards"}
	forward.AddCommand(&cobra.Command{Use: "list NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		m, e := o.manager()
		if e != nil {
			return e
		}
		n, e := m.Network(cmd.Context(), args[0])
		if e != nil {
			return e
		}
		return emit(cmd, n.Forwards)
	}})
	for _, action := range []string{"add", "remove"} {
		action := action
		var bind string
		var host, guest int
		command := &cobra.Command{Use: action + " NAME", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			m, e := o.manager()
			if e != nil {
				return e
			}
			f := config.Forward{Bind: bind, Host: host, Guest: guest}
			change := vm.NetworkChange{}
			if action == "add" {
				change.Add = &f
			} else {
				change.Remove = &f
			}
			return m.ChangeNetwork(cmd.Context(), args[0], change)
		}}
		command.Flags().StringVar(&bind, "bind", "127.0.0.1", "Host listen IP")
		command.Flags().IntVar(&host, "host", 0, "Host TCP port")
		if err := command.MarkFlagRequired("host"); err != nil {
			panic(err)
		}
		if action == "add" {
			command.Flags().IntVar(&guest, "guest", 0, "Guest TCP port")
			if err := command.MarkFlagRequired("guest"); err != nil {
				panic(err)
			}
		}
		forward.AddCommand(command)
	}
	network.AddCommand(show, set, forward)
	root.AddCommand(network)
}
