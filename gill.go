package gill

import (
	"context"
	"os/exec"

	gillExec "github.com/masshash/gill/exec"
)

func Command(name string, arg ...string) *gillExec.CmdWrapper {
	cmd := exec.Command(name, arg...)
	return gillExec.NewCmdWrapper(cmd)
}

func CommandContext(ctx context.Context, name string, arg ...string) *gillExec.CmdWrapper {
	cmd := exec.CommandContext(ctx, name, arg...)
	cw := gillExec.NewCmdWrapper(cmd)
	cw.SetCancelHandler(func() error {
		return cw.ProcessGroup.Kill()
	})
	return cw
}
