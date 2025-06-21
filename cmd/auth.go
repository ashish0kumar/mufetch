package cmd

import (
	"fmt"
	"os"

	"mufetch/pkg/config"

	"github.com/spf13/cobra"
)

// authCmd represents the authentication command for Spotify API
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Spotify API",
	Long: `Set up your Spotify API credentials.

You need to:
1. Go to https://developer.spotify.com/dashboard
2. Create a new app
3. Copy your Client ID and Client Secret`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Spotify API Authentication Setup")
		fmt.Println()
		fmt.Println("To get your Spotify API credentials:")
		fmt.Println("1. Go to: https://developer.spotify.com/dashboard")
		fmt.Println("2. Log in with your Spotify account")
		fmt.Println("3. Click 'Create an App'")
		fmt.Println("4. Fill in app name and description")
		fmt.Println("5. Copy your Client ID and Client Secret")
		fmt.Println()

		var clientID, clientSecret string

		fmt.Print("Enter your Spotify Client ID: ")
		fmt.Scanln(&clientID)

		fmt.Print("Enter your Spotify Client Secret: ")
		fmt.Scanln(&clientSecret)

		// Validate credentials
		if clientID == "" || clientSecret == "" {
			fmt.Println("Error: Both Client ID and Client Secret are required!")
			os.Exit(1)
		}
		if len(clientID) < 10 || len(clientSecret) < 10 {
			fmt.Println("Warning: Credentials seem too short. Please verify they are correct.")
		}

		// Set credentials in config
		if err := config.SetCredentials(clientID, clientSecret); err != nil {
			fmt.Printf("Failed to save credentials: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Credentials saved successfully!")
		fmt.Println("You can now use 'mufetch search <query>' to search for music.")
	},
}

// init adds the auth command to the root command
func init() {
	rootCmd.AddCommand(authCmd)
}
