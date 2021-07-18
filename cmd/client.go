package cmd

import (
	"log"

	"github.com/coomp/ccs/cmd/app"
	"github.com/spf13/cobra"
)

func NewClientCmd() *cobra.Command {

	var clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Send CCS request",
		Long:  `Send CCS request.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			method := args[0]
			// TODO
			client := app.NewMessageClient("localhost:2388")
			switch method {
			case "echo":
				if len(args) > 1 {
					client.Echo(args[1])
				} else {
					log.Println("Required argument(s) not present")
				}
			case "msg":
				client.MessageRequest()
			}
		},
	}

	return clientCmd
}
