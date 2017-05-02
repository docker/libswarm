package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/docker/swarmkit/agent/configs"
	"github.com/docker/swarmkit/agent/exec"
	"github.com/docker/swarmkit/agent/secrets"
	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/api/naming"
	"github.com/pkg/errors"
)

// Context defines the strict set of values that can be injected into a
// template expression in SwarmKit data structure.
// NOTE: Be very careful adding any fields to this structure with types
// that have methods defined on them. The template would be able to
// invoke those methods.
type Context struct {
	Service struct {
		ID     string
		Name   string
		Labels map[string]string
	}

	Node struct {
		ID string
	}

	Task struct {
		ID   string
		Name string
		Slot string

		// NOTE(stevvooe): Why no labels here? Tasks don't actually have labels
		// (from a user perspective). The labels are part of the container! If
		// one wants to use labels for templating, use service labels!
	}
}

// NewContextFromTask returns a new template context from the data available in
// task. The provided context can then be used to populate runtime values in a
// ContainerSpec.
func NewContextFromTask(t *api.Task) (ctx Context) {
	ctx.Service.ID = t.ServiceID
	ctx.Service.Name = t.ServiceAnnotations.Name
	ctx.Service.Labels = t.ServiceAnnotations.Labels

	ctx.Node.ID = t.NodeID

	ctx.Task.ID = t.ID
	ctx.Task.Name = naming.Task(t)

	if t.Slot != 0 {
		ctx.Task.Slot = fmt.Sprint(t.Slot)
	} else {
		// fall back to node id for slot when there is no slot
		ctx.Task.Slot = t.NodeID
	}

	return
}

// Expand treats the string s as a template and populates it with values from
// the context.
func (ctx *Context) Expand(s string) (string, error) {
	tmpl, err := newTemplate(s, nil)
	if err != nil {
		return s, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return s, err
	}

	return buf.String(), nil
}

// PayloadContext provides a context for expanding a config or secret payload.
// NOTE: Be very careful adding any fields to this structure with types
// that have methods defined on them. The template would be able to
// invoke those methods.
type PayloadContext struct {
	Context

	t                 *api.Task
	restrictedSecrets exec.SecretGetter
	restrictedConfigs exec.ConfigGetter
}

func (ctx PayloadContext) secretGetter(sourceOrTarget string, opts ...string) (string, error) {
	if ctx.restrictedSecrets == nil {
		return "", errors.New("secrets unavailable")
	}

	container := ctx.t.Spec.GetContainer()
	if container == nil {
		return "", errors.New("task is not a container")
	}

	if len(opts) > 1 {
		return "", errors.New("too many arguments to secret")
	}

	bySource := false

	if len(opts) == 1 {
		switch opts[0] {
		case "bysource=true":
			bySource = true
		case "bytarget=true":
		default:
			return "", errors.Errorf("unrecognized secret argument %q", opts[0])
		}
	}

	if bySource {
		for _, secretRef := range container.Secrets {
			if secretRef.SecretName == sourceOrTarget {
				secret, err := ctx.restrictedSecrets.Get(secretRef.SecretID)
				if err != nil {
					return "", err
				}
				return string(secret.Spec.Data), nil
			}
		}

		return "", errors.Errorf("secret source %s not found", sourceOrTarget)
	}

	for _, secretRef := range container.Secrets {
		file := secretRef.GetFile()
		if file != nil && file.Name == sourceOrTarget {
			secret, err := ctx.restrictedSecrets.Get(secretRef.SecretID)
			if err != nil {
				return "", err
			}
			return string(secret.Spec.Data), nil
		}
	}

	return "", errors.Errorf("secret target %s not found", sourceOrTarget)
}

func (ctx PayloadContext) configGetter(sourceOrTarget string, opts ...string) (string, error) {
	if ctx.restrictedConfigs == nil {
		return "", errors.New("configs unavailable")
	}

	container := ctx.t.Spec.GetContainer()
	if container == nil {
		return "", errors.New("task is not a container")
	}

	if len(opts) > 1 {
		return "", errors.New("too many arguments to secret")
	}

	bySource := false

	if len(opts) == 1 {
		switch opts[0] {
		case "bysource=true":
			bySource = true
		case "bytarget=true":
		default:
			return "", errors.Errorf("unrecognized secret argument %q", opts[0])
		}
	}

	if bySource {
		for _, configRef := range container.Configs {
			if configRef.ConfigName == sourceOrTarget {
				config, err := ctx.restrictedConfigs.Get(configRef.ConfigID)
				if err != nil {
					return "", err
				}
				return string(config.Spec.Data), nil
			}
		}

		return "", errors.Errorf("config source %s not found", sourceOrTarget)
	}

	for _, configRef := range container.Configs {
		file := configRef.GetFile()
		if file != nil && file.Name == sourceOrTarget {
			config, err := ctx.restrictedConfigs.Get(configRef.ConfigID)
			if err != nil {
				return "", err
			}
			return string(config.Spec.Data), nil
		}
	}

	return "", errors.Errorf("config target %s not found", sourceOrTarget)
}

func (ctx PayloadContext) envGetter(variable string) (string, error) {
	container := ctx.t.Spec.GetContainer()
	if container == nil {
		return "", errors.New("task is not a container")
	}

	for _, env := range container.Env {
		parts := strings.SplitN(env, "=", 2)

		if len(parts) > 1 && parts[0] == variable {
			return parts[1], nil
		}
	}
	return "", nil
}

// NewPayloadContextFromTask returns a new template context from the data
// available in the task. This context also provides access to the configs
// and secrets that the task has access to. The provided context can then
// be used to populate runtime values in a templated config or secret.
func NewPayloadContextFromTask(t *api.Task, dependencies exec.DependencyGetter) (ctx PayloadContext) {
	return PayloadContext{
		Context:           NewContextFromTask(t),
		t:                 t,
		restrictedSecrets: secrets.Restrict(dependencies.Secrets(), t),
		restrictedConfigs: configs.Restrict(dependencies.Configs(), t),
	}
}

// Expand treats the string s as a template and populates it with values from
// the context.
func (ctx *PayloadContext) Expand(s string) (string, error) {
	funcMap := template.FuncMap{
		"secret": ctx.secretGetter,
		"config": ctx.configGetter,
		"env":    ctx.envGetter,
	}

	tmpl, err := newTemplate(s, funcMap)
	if err != nil {
		return s, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return s, err
	}

	return buf.String(), nil
}
