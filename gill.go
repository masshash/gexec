package gill

import (
	"context"
	"os/exec"

	gillExec "github.com/masshash/gill/exec"
)

func Command(name string, arg ...string) *gillExec.Cmd {
	cmd := exec.Command(name, arg...)
	return gillExec.NewCmd(cmd)
}

func CommandContext(ctx context.Context, name string, arg ...string) *gillExec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	gcmd := gillExec.NewCmd(cmd)
	gcmd.SetCancelHandler(func() error {
		return gcmd.ProcessGroup.Kill()
	})
	return gcmd
}
