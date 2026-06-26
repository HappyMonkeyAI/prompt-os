package agent

import (
	"fmt"
)

type InstallerBuilder struct{}

func NewInstallerBuilder() *InstallerBuilder {
	return &InstallerBuilder{}
}

func (b *InstallerBuilder) BuildPlan(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("installer name must not be empty")
	}
	return fmt.Sprintf("builder(%s): plan ready", name), nil
}
