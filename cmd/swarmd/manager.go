package main

import (
	"github.com/docker/swarm-v2/manager"
	"github.com/docker/swarm-v2/manager/dispatcher"
	"github.com/spf13/cobra"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Run the swarm manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("listen-addr")
		if err != nil {
			return err
		}

		joinRaft, err := cmd.Flags().GetString("join-cluster")
		if err != nil {
			return err
		}

		stateDir, err := cmd.Flags().GetString("state-dir")
		if err != nil {
			return err
		}

		m, err := manager.New(&manager.Config{
			ListenProto:      "tcp",
			ListenAddr:       addr,
			JoinRaft:         joinRaft,
			StateDir:         stateDir,
			DispatcherConfig: dispatcher.DefaultConfig(),
		})
		if err != nil {
			return err
		}
		return m.Run()
	},
}

func init() {
	managerCmd.Flags().String("listen-addr", "0.0.0.0:4242", "Listen address")
	managerCmd.Flags().String("join-cluster", "", "Join cluster with a node at this address")
	managerCmd.Flags().String("state-dir", "/var/lib/docker/cluster", "State directory")
}
