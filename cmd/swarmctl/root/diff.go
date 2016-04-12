package root

import (
	"fmt"

	"github.com/docker/swarm-v2/cmd/swarmctl/common"
	"github.com/docker/swarm-v2/pb/docker/cluster/api"
	specspb "github.com/docker/swarm-v2/pb/docker/cluster/specs"
	"github.com/docker/swarm-v2/spec"
	"github.com/spf13/cobra"
)

var (
	diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "Diff an app",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := common.Dial(cmd)
			if err != nil {
				return err
			}

			flags := cmd.Flags()

			context, err := flags.GetInt("context")
			if err != nil {
				return err
			}

			r, err := c.ListJobs(common.Context(cmd), &api.ListJobsRequest{})
			if err != nil {
				return err
			}

			localSpec, err := readSpec(flags)
			if err != nil {
				return err
			}

			jobspecs := []*specspb.JobSpec{}

			for _, j := range r.Jobs {
				if j.Spec.Meta.Labels["namespace"] == localSpec.Namespace {
					jobspecs = append(jobspecs, j.Spec)
				}
			}
			remoteSpec := &spec.Spec{
				Version:   localSpec.Version,
				Namespace: localSpec.Namespace,
				Services:  make(map[string]*spec.ServiceConfig),
			}
			remoteSpec.FromJobSpecs(jobspecs)

			diff, err := localSpec.Diff(context, "remote", "local", remoteSpec)
			if err != nil {
				return err
			}
			fmt.Print(diff)
			return nil
		},
	}
)

func init() {
	diffCmd.Flags().StringP("file", "f", "docker.yml", "Spec file to diff")
	diffCmd.Flags().IntP("context", "c", 3, "lines of copied context (default 3)")
}
