package commands

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// PingCommand is the "ping" CLI command
	PingCommand = &cobra.Command{
		Use:   "ping",
		Short: "Check API connection",
		Long:  `Tests the connection to the API`,
		Run:   ping,
	}
)

func init() {
	RootCommand.AddCommand(PingCommand)
}

// ping checks the API connections
func ping(cmd *cobra.Command, args []string) {
	uri := "https://api.giantswarm.io/v1/ping"
	start := time.Now()
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println(color.RedString("API cannot be reached"))
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)
	if resp.StatusCode == 200 {
		fmt.Println(color.GreenString("API connection is fine"))
		fmt.Printf("Ping took %v\n", elapsed)
	} else {
		fmt.Println(color.RedString("API returned unexpected response status %v", resp.StatusCode))
		os.Exit(2)
	}
}
