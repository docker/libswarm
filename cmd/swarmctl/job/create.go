package job

import (
	"errors"
	"fmt"

	"github.com/docker/swarm-v2/cmd/swarmctl/common"
	"github.com/docker/swarm-v2/pb/docker/cluster/api"
	specspb "github.com/docker/swarm-v2/pb/docker/cluster/specs"
	typespb "github.com/docker/swarm-v2/pb/docker/cluster/types"
	"github.com/spf13/cobra"
)

var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a job",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			var spec *specspb.JobSpec

			if flags.Changed("file") {
				service, err := readServiceConfig(flags)
				if err != nil {
					return err
				}
				spec = service.ToProto()
			} else { // TODO(vieux): support or error on both file.
				if !flags.Changed("name") || !flags.Changed("image") {
					return errors.New("--name and --image are mandatory")
				}
				name, err := flags.GetString("name")
				if err != nil {
					return err
				}
				image, err := flags.GetString("image")
				if err != nil {
					return err
				}
				instances, err := flags.GetInt64("instances")
				if err != nil {
					return err
				}

				containerArgs, err := flags.GetStringSlice("args")
				if err != nil {
					return err
				}

				env, err := flags.GetStringSlice("env")
				if err != nil {
					return err
				}

				spec = &specspb.JobSpec{
					Meta: specspb.Meta{
						Name: name,
					},
					Template: &specspb.TaskSpec{
						Runtime: &specspb.TaskSpec_Container{
							Container: &typespb.Container{
								Image: &typespb.Image{
									Reference: image,
								},
								Command: args,
								Args:    containerArgs,
								Env:     env,
							},
						},
					},
					Orchestration: &specspb.JobSpec_Service{
						Service: &specspb.JobSpec_ServiceJob{
							Instances: instances,
						},
					},
				}
			}

			c, err := common.Dial(cmd)
			if err != nil {
				return err
			}
			r, err := c.CreateJob(common.Context(cmd), &api.CreateJobRequest{Spec: spec})
			if err != nil {
				return err
			}
			fmt.Println(r.Job.ID)
			return nil
		},
	}
)

func init() {
	createCmd.Flags().String("name", "", "Job name")
	createCmd.Flags().String("image", "", "Image")
	createCmd.Flags().StringSlice("args", nil, "Args")
	createCmd.Flags().StringSlice("env", nil, "Env")
	createCmd.Flags().StringP("file", "f", "", "Spec to use")
	// TODO(aluzzardi): This should be called `service-instances` so that every
	// orchestrator can have its own flag namespace.
	createCmd.Flags().Int64("instances", 1, "Number of instances for the service Job")
}
