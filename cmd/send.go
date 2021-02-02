package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ftl/tci/client"
)

var sendCmd = &cobra.Command{
	Use:   "send <text>",
	Short: "Send the given text as CW",
	Run:   runWithClient(send),
}

func init() {
	rootCmd.AddCommand(sendCmd)
}

func send(ctx context.Context, c *client.Client, _ *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("no text to send, use tci send <text>")
	}
	c.SendCWMacro(*rootFlags.trx, strings.Join(args, " "))
}
