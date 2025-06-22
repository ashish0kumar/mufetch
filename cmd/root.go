package cmd

import (
	"fmt"
	"os"

	"github.com/ashish0kumar/mufetch/pkg/config"
	"github.com/ashish0kumar/mufetch/pkg/display"
	"github.com/ashish0kumar/mufetch/pkg/spotify"
	"github.com/spf13/cobra"
)

// version of the application
var version = "dev"

// variables to hold command line args and configuration
var (
	searchType string
	imageSize  int
	cfg        *config.Config
	client     *spotify.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mufetch",
	Short: "neofetch-like CLI for music",
	Long: `mufetch displays beautiful music information with cover art in your terminal.
Search for tracks, albums, or artists.`,
	Version: version,
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for music",
	Long:  `Search for tracks, albums, or artists and display their metadata`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		if !config.HasCredentials() {
			fmt.Println("No Spotify credentials found!")
			fmt.Println("Run 'mufetch auth' to set up your API credentials.")
			os.Exit(1)
		}

		// Load config
		var err error
		cfg, err = config.GetConfig()
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}

		// Initialize Spotify client with credentials
		client = spotify.NewClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)

		fmt.Print("\033[?25l")
		defer fmt.Print("\033[?25h")

		fmt.Printf("\n")

		// Validate image size
		if imageSize < 15 {
			imageSize = 15
		}
		if imageSize > 35 {
			imageSize = 35
		}

		// Perform search
		if searchType == "auto" {
			searchAuto(query)
		} else {
			searchSpecific(query, searchType)
		}

		// Move cursor up and clear the line
		fmt.Print("\033[F\033[K\n")
	},
}

// searchAuto performs an automatic search based on the query
func searchAuto(query string) {
	// Try track first
	if result, err := client.Search(query, "track"); err == nil && len(result.Tracks.Items) > 0 {
		display.DisplayTrack(result.Tracks.Items[0], client, imageSize)
		return
	}

	// Try album
	if result, err := client.Search(query, "album"); err == nil && len(result.Albums.Items) > 0 {
		if album, err := client.GetAlbum(result.Albums.Items[0].ID); err == nil {
			display.DisplayAlbum(*album, client, imageSize)
			return
		}
	}

	// Try artist
	if result, err := client.Search(query, "artist"); err == nil && len(result.Artists.Items) > 0 {
		if artist, err := client.GetArtist(result.Artists.Items[0].ID); err == nil {
			display.DisplayArtist(*artist, client, imageSize)
			return
		}
	}

	fmt.Printf("No results found for: %s\n", query)
}

// searchSpecific performs a search for a specific type (track, album, artist)
func searchSpecific(query, sType string) {

	result, err := client.Search(query, sType)
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		os.Exit(1)
	}

	switch sType {
	case "track":
		if len(result.Tracks.Items) > 0 {
			display.DisplayTrack(result.Tracks.Items[0], client, imageSize)
		} else {
			fmt.Printf("No tracks found for: %s\n", query)
		}
	case "album":
		if len(result.Albums.Items) > 0 {
			if album, err := client.GetAlbum(result.Albums.Items[0].ID); err == nil {
				display.DisplayAlbum(*album, client, imageSize)
			} else {
				fmt.Printf("Failed to get album details: %v\n", err)
			}
		} else {
			fmt.Printf("No albums found for: %s\n", query)
		}
	case "artist":
		if len(result.Artists.Items) > 0 {
			if artist, err := client.GetArtist(result.Artists.Items[0].ID); err == nil {
				display.DisplayArtist(*artist, client, imageSize)
			} else {
				fmt.Printf("Failed to get artist details: %v\n", err)
			}
		} else {
			fmt.Printf("No artists found for: %s\n", query)
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init initializes the root command and adds subcommands
func init() {
	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Flags for search command
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "auto", "Search type: track, album, artist, or auto")
	searchCmd.Flags().IntVarP(&imageSize, "size", "s", 20, "Image size (20-50)")

	rootCmd.AddCommand(searchCmd)
}
