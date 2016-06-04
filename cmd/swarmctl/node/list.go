package node

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/cmd/swarmctl/common"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "ls",
		Short: "List nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			quiet, err := flags.GetBool("quiet")
			if err != nil {
				return err
			}

			c, err := common.Dial(cmd)
			if err != nil {
				return err
			}
			r, err := c.ListNodes(common.Context(cmd), &api.ListNodesRequest{})
			if err != nil {
				return err
			}

			var output func(n *api.Node)

			if !quiet {
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				defer func() {
					// Ignore flushing errors - there's nothing we can do.
					_ = w.Flush()
				}()
				common.PrintHeader(w, "ID", "Name", "Status", "Availability/Membership", "Manager status", "Leader")
				output = func(n *api.Node) {
					spec := &n.Spec
					name := spec.Annotations.Name
					membership := spec.Availability.String()
					if n.Certificate.Status.State != api.IssuanceStateIssued && n.Certificate.Status.State != api.IssuanceStateRenew {
						membership = spec.Membership.String()
					}
					if name == "" && n.Description != nil {
						name = n.Description.Hostname
					}
					leader := ""
					if n.Manager != nil && n.Manager.Raft.Status.Leader {
						leader = "*"
					}
					reachability := ""
					if n.Manager != nil {
						reachability = n.Manager.Raft.Status.Reachability.String()
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
						n.ID,
						name,
						n.Status.State.String(),
						membership,
						reachability,
						leader,
					)
				}
			} else {
				output = func(n *api.Node) { fmt.Println(n.ID) }
			}

			for _, n := range r.Nodes {
				output(n)
			}
			return nil
		},
	}
)

func init() {
	listCmd.Flags().BoolP("quiet", "q", false, "Only display IDs")
}
