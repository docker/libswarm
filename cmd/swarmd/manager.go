package main

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/docker/swarm-v2/ca"
	"github.com/docker/swarm-v2/manager"
	"github.com/docker/swarm-v2/picker"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Run the swarm manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		addr, err := cmd.Flags().GetString("listen-addr")
		if err != nil {
			return err
		}
		addrHost, _, err := net.SplitHostPort(addr)
		if err == nil {
			ip := net.ParseIP(addrHost)
			if ip != nil && (ip.IsUnspecified() || ip.IsLoopback()) {
				fmt.Println("Warning: Specifying a valid address with --listen-addr may be necessary for other managers to reach this one.")
			}
		}

		managerAddr, err := cmd.Flags().GetString("join-cluster")
		if err != nil {
			return err
		}

		forceNewCluster, err := cmd.Flags().GetBool("force-new-cluster")
		if err != nil {
			return err
		}

		stateDir, err := cmd.Flags().GetString("state-dir")
		if err != nil {
			return err
		}

		certDir := filepath.Join(stateDir, "certificates")

		token, err := cmd.Flags().GetString("token")
		if err != nil {
			return err
		}

		// Create a context for our GRPC call
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var p *picker.Picker
		if managerAddr != "" {
			managers := picker.NewRemotes(managerAddr)
			p = picker.NewPicker(managerAddr, managers)
		}

		// If we don't have a valid picker, means that we're bootstraping a new cluster
		if p == nil {
			if err := ca.BootstrapCluster(certDir); err != nil {
				return err
			}
		}

		// We either just boostraped our cluster from scratch, or have a valid picker and
		// are thus joining an existing cluster
		securityConfig, err := ca.LoadOrCreateSecurityConfig(ctx, certDir, token, p)
		if err != nil {
			return err
		}

		m, err := manager.New(&manager.Config{
			ListenProto:     "tcp",
			SecurityConfig:  securityConfig,
			ListenAddr:      addr,
			ForceNewCluster: forceNewCluster,
			JoinRaft:        managerAddr,
			StateDir:        stateDir,
		})
		if err != nil {
			return err
		}
		return m.Run(ctx)
	},
}

func init() {
	managerCmd.Flags().String("listen-addr", "0.0.0.0:4242", "Listen address")
	managerCmd.Flags().String("join-cluster", "", "Join cluster with a node at this address")
	managerCmd.Flags().Bool("force-new-cluster", false, "Force the creation of a new cluster from data directory")
}
