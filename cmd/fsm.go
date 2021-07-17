package cmd

import (
	"fmt"
	"strings"

	"github.com/coomp/ccs/cmd/app"
	"github.com/spf13/cobra"
)

func NewFSMCmd() *cobra.Command {

	var fsmCmd = &cobra.Command{
		Use:   "fsm",
		Short: "Show fsm working flow",
		Long:  `Show fsm working flow.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Run FSM app with args: " + strings.Join(args, " ") + " ...")
			app := &app.FSMApp{}
			app.Run()
		},
	}
	return fsmCmd
}
