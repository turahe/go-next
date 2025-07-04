package cmd

import (
	"os"
	"wordpress-go-next/backend/internal"

	"github.com/spf13/cobra"
)

var host string
var port string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the RESTful API server",
	Run: func(cmd *cobra.Command, args []string) {
		if host == "" {
			host = "0.0.0.0"
		}
		if port == "" {
			port = "8080"
		}
		// Set environment variables so internal.RunServer can use them
		_ = os.Setenv("HOST", host)
		_ = os.Setenv("PORT", port)
		internal.RunServer(host, port)
	},
}

func init() {
	serverCmd.Flags().StringVar(&host, "host", "", "Host to bind the server to (default: 0.0.0.0)")
	serverCmd.Flags().StringVar(&port, "port", "", "Port to run the server on (default: 8080)")
	rootCmd.AddCommand(serverCmd)
}
