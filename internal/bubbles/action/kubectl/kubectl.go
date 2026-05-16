package kubectl

import (
	tea "github.com/charmbracelet/bubbletea"
)

type shell interface {
	Exec(c string, args ...string) tea.Cmd
	Pager(c string, args ...string) tea.Cmd
}

type Cmd struct {
	kubectx string
	shell   shell
}

func New(kubectx string, s shell) *Cmd {
	return &Cmd{kubectx: kubectx, shell: s}
}

func (k *Cmd) Edit(ns, resource string) tea.Cmd {
	args := []string{"edit", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	if k.kubectx != "" {
		args = append(args, "--context", k.kubectx)
	}
	return k.shell.Exec("kubectl", args...)
}

func (k *Cmd) Pause(ns, resource string) tea.Cmd {
	args := []string{"annotate"}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	if k.kubectx != "" {
		args = append(args, "--context", k.kubectx)
	}
	args = append(args, resource, "--overwrite", "crossplane.io/paused=true")
	return k.shell.Exec("kubectl", args...)
}

func (k *Cmd) Unpause(ns, resource string) tea.Cmd {
	args := []string{"annotate"}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	if k.kubectx != "" {
		args = append(args, "--context", k.kubectx)
	}
	args = append(args, resource, "--overwrite", "crossplane.io/paused-")
	return k.shell.Exec("kubectl", args...)
}

func (k *Cmd) Describe(ns, resource string) tea.Cmd {
	args := []string{"describe", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	if k.kubectx != "" {
		args = append(args, "--context", k.kubectx)
	}
	return k.shell.Pager("kubectl", args...)
}

func (k *Cmd) Get(ns, resource string) tea.Cmd {
	args := []string{"get", resource, "-o", "yaml"}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	if k.kubectx != "" {
		args = append(args, "--context", k.kubectx)
	}
	return k.shell.Pager("kubectl", args...)
}

func (k *Cmd) Delete(ns, resource string) tea.Cmd {
	args := []string{"delete", resource}
	if ns != "" {
		args = append(args, "-n", ns, "--wait", "false")
	}
	if k.kubectx != "" {
		args = append(args, "--context", k.kubectx)
	}
	return k.shell.Exec("kubectl", args...)
}
