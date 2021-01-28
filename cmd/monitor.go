package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/ftl/tci/client"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor host",
	Short: "Connect to the given host and log the incoming messages to stdout.",
	Run:   runWithClient(monitor),
}

func init() {
	rootCmd.AddCommand(monitorCmd)
}

func monitor(ctx context.Context, c *client.Client, _ *cobra.Command, _ []string) {
	c.Notify(new(messageLogger))
	<-ctx.Done()
}

type messageLogger struct{}

func (l *messageLogger) Message(msg client.Message) {
	log.Print(msg)
}
