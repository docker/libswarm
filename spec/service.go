package spec

import (
	"fmt"
	"io"
	"strings"
	"time"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/docker/swarm-v2/api"
	"github.com/pmezard/go-difflib/difflib"
)

// ContainerConfig is a human representation of the ContainerSpec
type ContainerConfig struct {
	Image string `yaml:"image,omitempty"`

	// Command to run the the container. The first element is a path to the
	// executable and the following elements are treated as arguments.
	//
	// If command is empty, execution will fall back to the image's entrypoint.
	Command []string `yaml:"command,omitempty"`

	// Args specifies arguments provided to the image's entrypoint.
	// Ignored if command is specified.
	Args []string `yaml:"args,omitempty"`

	// Env specifies the environment variables for the container in NAME=VALUE
	// format. These must be compliant with  [IEEE Std
	// 1003.1-2001](http://pubs.opengroup.org/onlinepubs/009695399/basedefs/xbd_chap08.html).
	Env []string `yaml:"env,omitempty"`

	// Networks specifies all the networks that this service is attached to.
	Networks  []string              `yaml:"networks,omitempty"`
	Resources *ResourceRequirements `yaml:"resources,omitempty"`

	// Mounts describe how volumes should be mounted in the container
	Mounts Mounts `yaml:"mounts,omitempty"`
}

// PortConfig is a human representation of the PortConfiguration
type PortConfig struct {
	Name     string `yaml:"name,omitempty"`
	Protocol string `yaml:"protocol,omitempty"`
	Port     uint32 `yaml:"port,omitempty"`
	NodePort uint32 `yaml:"node_port,omitempty"`
}

// ServiceConfig is a human representation of the Service
type ServiceConfig struct {
	ContainerConfig

	Name      string `yaml:"name,omitempty"`
	Instances int64  `yaml:"instances,omitempty"`
	Mode      string `yaml:"mode,omitempty"`

	Restart      string `yaml:"restart,omitempty"`
	RestartDelay int64  `yaml:"restartdelay,omitempty"`

	Ports []PortConfig `yaml:"ports,omitempty"`
}

// Validate checks the validity of the ServiceConfig.
func (s *ServiceConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is mandatory")
	}
	if s.Image == "" {
		return fmt.Errorf("image is mandatory in %s", s.Name)
	}
	switch s.Mode {
	case "", "running", "fill", "batch":
	default:
		return fmt.Errorf("unrecognized mode %s", s.Mode)
	}

	switch s.Restart {
	case "", "no", "on-failure":
	case "always":
	default:
		return fmt.Errorf("unrecognized restart policy %s", s.Restart)
	}
	if err := s.Resources.Validate(); err != nil {
		return err
	}
	if err := s.Mounts.Validate(); err != nil {
		return err
	}

	return nil
}

// Reset resets the service config to its defaults.
func (s *ServiceConfig) Reset() {
	*s = ServiceConfig{
		Instances: 1,
	}
}

// Read reads a ServiceConfig from an io.Reader.
func (s *ServiceConfig) Read(r io.Reader) error {
	s.Reset()

	if err := yaml.NewDecoder(r).Decode(s); err != nil {
		return err
	}

	return s.Validate()
}

// Write writes a ServiceConfig to an io.Reader.
func (s *ServiceConfig) Write(w io.Writer) error {
	return yaml.NewEncoder(w).Encode(s)
}

// ToProto converts a ServiceConfig to a ServiceSpec.
func (s *ServiceConfig) ToProto() *api.ServiceSpec {
	spec := &api.ServiceSpec{
		Annotations: api.Annotations{
			Name:   s.Name,
			Labels: make(map[string]string),
		},
		Template: &api.TaskSpec{
			Runtime: &api.TaskSpec_Container{
				Container: &api.Container{
					Resources: s.Resources.ToProto(),
					Mounts:    s.Mounts.ToProto(),
					Image: &api.Image{
						Reference: s.Image,
					},
					Env:     s.Env,
					Command: s.Command,
					Args:    s.Args,
				},
			},
		},
		Instances: s.Instances,
		Restart: &api.RestartPolicy{
			Delay: time.Duration(s.RestartDelay) * time.Second,
		},
	}

	if len(s.Ports) != 0 {
		endpoint := &api.Endpoint{}
		for _, portConfig := range s.Ports {
			endpoint.Ports = append(endpoint.Ports, &api.Endpoint_PortConfiguration{
				Name:     portConfig.Name,
				Protocol: api.Endpoint_Protocol(api.Endpoint_Protocol_value[strings.ToUpper(portConfig.Protocol)]),
				Port:     portConfig.Port,
				NodePort: portConfig.NodePort,
			})
		}

		spec.Endpoint = endpoint
	}

	if len(s.Networks) != 0 {
		networks := make([]*api.Container_NetworkAttachment, 0, len(s.Networks))
		for _, nwkName := range s.Networks {
			networks = append(networks, &api.Container_NetworkAttachment{
				Reference: &api.Container_NetworkAttachment_Name{
					Name: nwkName,
				},
			})
		}

		spec.Template.GetContainer().Networks = networks

	}

	switch s.Mode {
	case "", "running":
		spec.Mode = api.ServiceModeRunning
	case "batch":
		spec.Mode = api.ServiceModeBatch
	case "fill":
		spec.Mode = api.ServiceModeFill
	}

	switch s.Restart {
	case "no":
		spec.Restart.Condition = api.RestartNever
	case "on-failure":
		spec.Restart.Condition = api.RestartOnFailure
	case "", "always":
		spec.Restart.Condition = api.RestartAlways
	}

	return spec
}

// FromProto converts a ServiceSpec to a ServiceConfig.
func (s *ServiceConfig) FromProto(serviceSpec *api.ServiceSpec) {
	*s = ServiceConfig{
		Name:      serviceSpec.Annotations.Name,
		Instances: serviceSpec.Instances,
		ContainerConfig: ContainerConfig{
			Image:   serviceSpec.Template.GetContainer().Image.Reference,
			Env:     serviceSpec.Template.GetContainer().Env,
			Args:    serviceSpec.Template.GetContainer().Args,
			Command: serviceSpec.Template.GetContainer().Command,
		},
	}
	if serviceSpec.Template.GetContainer().Resources != nil {
		s.Resources = &ResourceRequirements{}
		s.Resources.FromProto(serviceSpec.Template.GetContainer().Resources)
	}
	if serviceSpec.Template.GetContainer().Mounts != nil {
		apiMounts := serviceSpec.Template.GetContainer().Mounts
		s.Mounts = make(Mounts, len(apiMounts))
		s.Mounts.FromProto(apiMounts)
	}

	switch serviceSpec.Mode {
	case api.ServiceModeRunning:
		s.Mode = "running"
	case api.ServiceModeFill:
		s.Mode = "fill"
	case api.ServiceModeBatch:
		s.Mode = "batch"
	}

	if serviceSpec.Restart != nil {
		switch serviceSpec.Restart.Condition {
		case api.RestartNever:
			s.Restart = "no"
		case api.RestartOnFailure:
			s.Restart = "on-failure"
		case api.RestartAlways:
			s.Restart = "always"
		}
		s.RestartDelay = int64(serviceSpec.Restart.Delay / time.Second)
	}
}

// Diff returns a diff between two ServiceConfigs.
func (s *ServiceConfig) Diff(context int, fromFile, toFile string, other *ServiceConfig) (string, error) {
	from, err := yaml.Marshal(other)
	if err != nil {
		return "", err
	}

	to, err := yaml.Marshal(s)
	if err != nil {
		return "", err
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(from)),
		FromFile: fromFile,
		B:        difflib.SplitLines(string(to)),
		ToFile:   toFile,
		Context:  context,
	}

	return difflib.GetUnifiedDiffString(diff)
}
