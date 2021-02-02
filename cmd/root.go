package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ftl/tci/client"
	"github.com/spf13/cobra"
)

var rootFlags = struct {
	hostAddress *string
	trx         *int
	reconnect   *bool
}{}

var rootCmd = &cobra.Command{
	Use:   "tci",
	Short: "A simple client for the TCI protocol.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootFlags.hostAddress = rootCmd.PersistentFlags().StringP("host", "", "localhost:40001", "connect to this TCI host")
	rootFlags.trx = rootCmd.PersistentFlags().IntP("trx", "t", 0, "use this TRX")
	rootFlags.reconnect = rootCmd.PersistentFlags().BoolP("reconnect", "r", false, "try to reconnect if the TCI connection failed")
}

func runWithClient(f func(context.Context, *client.Client, *cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		host, err := parseHostArg(*rootFlags.hostAddress)
		if err != nil {
			log.Fatalf("invalid host address: %v", err)
		}
		if host.Port == 0 {
			host.Port = client.DefaultPort
			log.Print("using the default port")
		}

		ctx, cancel := context.WithCancel(context.Background())
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go handleCancelation(signals, cancel)

		var c *client.Client
		if *rootFlags.reconnect {
			c = client.KeepOpen(host, 30*time.Second)
		} else {
			c, err = client.Open(host)
		}
		if err != nil {
			log.Fatalf("cannot conntect to %s: %v", host.String(), err)
		}
		defer c.Disconnect()
		if !*rootFlags.reconnect {
			c.WhenDisconnected(cancel)
		}

		f(ctx, c, cmd, args)
	}
}

func handleCancelation(signals <-chan os.Signal, cancel context.CancelFunc) {
	count := 0
	for {
		select {
		case <-signals:
			count++
			if count == 1 {
				cancel()
			} else {
				log.Fatal("hard shutdown")
			}
		}
	}
}

func parseHostArg(arg string) (*net.TCPAddr, error) {
	host, port := splitHostPort(arg)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = strconv.Itoa(client.DefaultPort)
	}

	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
}

func splitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
