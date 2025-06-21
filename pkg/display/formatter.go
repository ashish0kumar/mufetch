package display

import (
	"fmt"
	"image"
	"net/http"
	"strings"
	"time"

	"mufetch/pkg/spotify"

	"github.com/disintegration/imaging"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// ImageRenderer handles terminal image rendering using Unicode blocks
type ImageRenderer struct {
	width  int
	height int
}

// NewImageRenderer creates an image renderer with specified size
func NewImageRenderer(size int) *ImageRenderer {
	return &ImageRenderer{
		width:  size,
		height: size,
	}
}

// RenderImageLines converts image URL to terminal-displayable lines
func (r *ImageRenderer) RenderImageLines(imageURL string) []string {
	if imageURL == "" {
		return r.getPlaceholderLines()
	}

	img, err := r.downloadImage(imageURL)
	if err != nil {
		return r.getPlaceholderLines()
	}

	return r.getBlockArtLines(img)
}

// downloadImage fetches and decodes image from URL
func (r *ImageRenderer) downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// getBlockArtLines converts image to colored terminal blocks
func (r *ImageRenderer) getBlockArtLines(img image.Image) []string {
	resized := imaging.Resize(img, r.width, r.height, imaging.Lanczos)
	bounds := resized.Bounds()

	var lines []string

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var line strings.Builder
		line.WriteString(" ") // Left padding

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixelColor := resized.At(x, y)
			r8, g8, b8, _ := pixelColor.RGBA()

			// Convert to 8-bit RGB values
			r := uint8(r8 >> 8)
			g := uint8(g8 >> 8)
			b := uint8(b8 >> 8)

			// Create colored block using true color ANSI codes
			line.WriteString(fmt.Sprintf("\033[48;2;%d;%d;%dm  \033[0m", r, g, b))
		}
		lines = append(lines, line.String())
	}

	return lines
}

// getPlaceholderLines creates a placeholder when no image is available
func (r *ImageRenderer) getPlaceholderLines() []string {
	lines := []string{
		fmt.Sprintf(" %s┌───────────────────────────────┐%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s│                               │%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s│                               │%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s│           NO IMAGE            │%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s│           AVAILABLE           │%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s│                               │%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s│                               │%s", ColorWhite, ColorReset),
		fmt.Sprintf(" %s└───────────────────────────────┘%s", ColorWhite, ColorReset),
	}

	// Pad to match image height
	for len(lines) < 20 {
		lines = append(lines, strings.Repeat(" ", 31))
	}

	return lines
}

// createClickableLink creates terminal hyperlink using ANSI escape codes
func createClickableLink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// DisplayTrack renders track information with album art
func DisplayTrack(track spotify.Track, client *spotify.Client, imageSize int) {
	renderer := NewImageRenderer(imageSize)

	var imageLines []string
	if len(track.Album.Images) > 0 {
		imageLines = renderer.RenderImageLines(track.Album.Images[0].URL)
	} else {
		imageLines = renderer.getPlaceholderLines()
	}

	// Create clickable artist links
	artistNames := make([]string, len(track.Artists))
	for i, artist := range track.Artists {
		artistNames[i] = createClickableLink(artist.ExternalURL.Spotify, artist.Name)
	}

	duration := time.Duration(track.Duration) * time.Millisecond

	// Get genres from album or fallback to artist genres
	genres := track.Album.Genres
	if len(genres) == 0 && len(track.Artists) > 0 && client != nil {
		if artist, err := client.GetArtist(track.Artists[0].ID); err == nil {
			genres = artist.Genres
		}
	}

	// Create clickable album name
	albumName := createClickableLink(track.Album.ExternalURL.Spotify, track.Album.Name)

	infoLines := []string{
		formatInfoLine("Name", track.Name, ColorGreen),
		formatInfoLine("Artist", strings.Join(artistNames, ", "), ColorYellow),
		formatInfoLine("Album", albumName, ColorBlue),
		formatInfoLine("Duration", formatDuration(duration), ColorWhite),
		formatInfoLine("Track", fmt.Sprintf("%d", track.TrackNumber), ColorCyan),
		formatInfoLine("Explicit", formatBool(track.Explicit), ColorRed),
		formatInfoLine("Released", formatOrdinalDate(track.Album.ReleaseDate), ColorCyan),
		formatInfoLine("Popularity", fmt.Sprintf("%d%%", track.Popularity), ColorPurple),
	}

	if len(genres) > 0 {
		genreString := strings.Join(genres, ", ")
		if len(genreString) > 50 {
			genreString = genreString[:47] + "..."
		}
		infoLines = append(infoLines, formatInfoLine("Genres", genreString, ColorRed))
	}

	// Prepare clickable links for bottom placement
	var links []string
	if len(track.Album.Images) > 0 {
		links = append(links, fmt.Sprintf("%s%s%s", ColorBlue, createClickableLink(track.Album.Images[0].URL, "Album Cover"), ColorReset))
	}
	links = append(links, fmt.Sprintf("%s%s%s", ColorGreen, createClickableLink(track.ExternalURL.Spotify, "Spotify"), ColorReset))

	displaySideBySideWithLinks(imageLines, infoLines, links)
}

// DisplayAlbum renders album information with cover art
func DisplayAlbum(album spotify.Album, client *spotify.Client, imageSize int) {
	renderer := NewImageRenderer(imageSize)

	var imageLines []string
	if len(album.Images) > 0 {
		imageLines = renderer.RenderImageLines(album.Images[0].URL)
	} else {
		imageLines = renderer.getPlaceholderLines()
	}

	// Create clickable artist links
	artistNames := make([]string, len(album.Artists))
	for i, artist := range album.Artists {
		artistNames[i] = createClickableLink(artist.ExternalURL.Spotify, artist.Name)
	}

	// Calculate total duration from all tracks
	totalDuration := 0
	for _, track := range album.Tracks.Items {
		totalDuration += track.Duration
	}

	// Get genres from album or fallback to artist genres
	genres := album.Genres
	if len(genres) == 0 && len(album.Artists) > 0 && client != nil {
		if artist, err := client.GetArtist(album.Artists[0].ID); err == nil {
			genres = artist.Genres
		}
	}

	infoLines := []string{
		formatInfoLine("Name", album.Name, ColorGreen),
		formatInfoLine("Artist", strings.Join(artistNames, ", "), ColorYellow),
		formatInfoLine("Type", album.AlbumType, ColorBlue),
		formatInfoLine("Released", formatOrdinalDate(album.ReleaseDate), ColorCyan),
		formatInfoLine("Tracks", fmt.Sprintf("%d", album.TotalTracks), ColorPurple),
		formatInfoLine("Duration", formatDuration(time.Duration(totalDuration)*time.Millisecond), ColorWhite),
		formatInfoLine("Popularity", fmt.Sprintf("%d%%", album.Popularity), ColorPurple),
	}

	if len(genres) > 0 {
		genreString := strings.Join(genres, ", ")
		if len(genreString) > 50 {
			genreString = genreString[:47] + "..."
		}
		infoLines = append(infoLines, formatInfoLine("Genres", genreString, ColorRed))
	}

	if len(album.Label) > 0 {
		infoLines = append(infoLines, formatInfoLine("Label", formatString(album.Label), ColorWhite))
	}

	// Add top tracks with clickable links
	if len(album.Tracks.Items) > 0 {
		infoLines = append(infoLines, "")
		infoLines = append(infoLines, fmt.Sprintf("%sTop Tracks%s", ColorBold, ColorReset))

		for i, track := range album.Tracks.Items {
			if i >= 5 {
				break
			}
			trackLink := createClickableLink(track.ExternalURL.Spotify, track.Name)
			infoLines = append(infoLines, fmt.Sprintf("%s%s%s", ColorGreen, trackLink, ColorReset))
		}
	}

	// Prepare clickable links for bottom placement
	var links []string
	if len(album.Images) > 0 {
		links = append(links, fmt.Sprintf("%s%s%s", ColorBlue, createClickableLink(album.Images[0].URL, "Album Cover"), ColorReset))
	}
	links = append(links, fmt.Sprintf("%s%s%s", ColorGreen, createClickableLink(album.ExternalURL.Spotify, "Spotify"), ColorReset))

	displaySideBySideWithLinks(imageLines, infoLines, links)
}

// DisplayArtist renders artist information with profile image
func DisplayArtist(artist spotify.Artist, client *spotify.Client, imageSize int) {
	renderer := NewImageRenderer(imageSize)

	var imageLines []string
	if len(artist.Images) > 0 {
		imageLines = renderer.RenderImageLines(artist.Images[0].URL)
	} else {
		imageLines = renderer.getPlaceholderLines()
	}

	// Fetch additional artist data from API
	var topTracks *spotify.TopTracksResponse
	var albums *spotify.ArtistAlbumsResponse
	var singles *spotify.ArtistAlbumsResponse

	if client != nil {
		topTracks, _ = client.GetArtistTopTracks(artist.ID)
		albums, _ = client.GetArtistAlbums(artist.ID, "album")
		singles, _ = client.GetArtistAlbums(artist.ID, "single")
	}

	infoLines := []string{
		formatInfoLine("Name", artist.Name, ColorGreen),
		formatInfoLine("Followers", formatNumber(artist.Followers.Total), ColorYellow),
		formatInfoLine("Popularity", fmt.Sprintf("%d%%", artist.Popularity), ColorPurple),
	}

	if len(artist.Genres) > 0 {
		genreString := strings.Join(artist.Genres, ", ")
		if len(genreString) > 50 {
			genreString = genreString[:47] + "..."
		}
		infoLines = append(infoLines, formatInfoLine("Genres", genreString, ColorRed))
	}

	if albums != nil {
		infoLines = append(infoLines, formatInfoLine("Albums", fmt.Sprintf("%d", albums.Total), ColorGreen))
	}

	if singles != nil {
		infoLines = append(infoLines, formatInfoLine("Singles", fmt.Sprintf("%d", singles.Total), ColorYellow))
	}

	// Add top tracks with clickable links
	if topTracks != nil && len(topTracks.Tracks) > 0 {
		infoLines = append(infoLines, "")
		infoLines = append(infoLines, fmt.Sprintf("%sTop Tracks%s", ColorBold, ColorReset))

		for i, track := range topTracks.Tracks {
			if i >= 5 {
				break
			}
			trackLink := createClickableLink(track.ExternalURL.Spotify, track.Name)
			infoLines = append(infoLines, fmt.Sprintf("%s%s%s", ColorGreen, trackLink, ColorReset))
		}
	}

	// Prepare clickable links for bottom placement
	var links []string
	links = append(links, fmt.Sprintf("%s%s%s", ColorGreen, createClickableLink(artist.ExternalURL.Spotify, "Spotify"), ColorReset))
	if len(artist.Images) > 0 {
		links = append(links, fmt.Sprintf("%s%s%s", ColorBlue, createClickableLink(artist.Images[0].URL, "Artist Photo"), ColorReset))
	}

	displaySideBySideWithLinks(imageLines, infoLines, links)
}

// displaySideBySideWithLinks renders image and info side-by-side with links at bottom
func displaySideBySideWithLinks(imageLines, infoLines, links []string) {
	maxLines := len(imageLines)
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}

	// Pad info lines to match image height minus 2 for link placement
	targetLines := maxLines - 2
	for len(infoLines) < targetLines {
		infoLines = append(infoLines, "")
	}

	// Display main content side by side
	for i := 0; i < targetLines; i++ {
		if i < len(imageLines) {
			fmt.Printf("%s   %s\n", imageLines[i], infoLines[i])
		} else {
			fmt.Printf("%s   %s\n", strings.Repeat(" ", 46), infoLines[i])
		}
	}

	// Display links 2 spaces up from bottom
	if len(links) > 0 {
		var imageLine string
		if targetLines < len(imageLines) {
			imageLine = imageLines[targetLines]
		} else {
			imageLine = strings.Repeat(" ", 46)
		}

		// Separate Spotify and image links for consistent ordering
		var spotifyLink, imageLink string
		for _, link := range links {
			if strings.Contains(strings.ToLower(link), "spotify") {
				spotifyLink = link
			} else {
				imageLink = link
			}
		}

		// Fallback if any link is missing
		if spotifyLink == "" && len(links) > 0 {
			spotifyLink = links[0]
		}
		if imageLink == "" && len(links) > 1 {
			imageLink = links[1]
		}

		// Format link line with proper spacing
		var linkLine string
		if spotifyLink != "" && imageLink != "" {
			linkLine = fmt.Sprintf("%s   %s", spotifyLink, imageLink)
		} else if spotifyLink != "" {
			linkLine = spotifyLink
		} else {
			linkLine = imageLink
		}

		fmt.Printf("%s   %s\n", imageLine, linkLine)

		// Display remaining image lines after links
		for i := targetLines + 1; i < len(imageLines); i++ {
			fmt.Printf("%s   \n", imageLines[i])
		}
	} else {
		// If no links, display remaining image lines normally
		for i := targetLines; i < len(imageLines); i++ {
			fmt.Printf("%s   \n", imageLines[i])
		}
	}
}

// formatInfoLine creates consistently formatted label-value pairs
func formatInfoLine(label, value, color string) string {
	const minPadding = 2
	const maxLabelWidth = 12

	labelWidth := len(label)
	padding := maxLabelWidth - labelWidth + minPadding
	if padding < minPadding {
		padding = minPadding
	}

	return fmt.Sprintf("%s%s%s%s%s%s%s",
		ColorBold, label, ColorReset,
		strings.Repeat(" ", padding),
		color, value, ColorReset)
}

// formatOrdinalDate converts date string to ordinal format (1st Jan 2020)
func formatOrdinalDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}

	// Parse YYYY-MM-DD format from Spotify
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// Fallback to year-only format
		if t, err = time.Parse("2006", dateStr); err != nil {
			return dateStr
		}
		return t.Format("2006")
	}

	day := t.Day()
	month := t.Format("Jan")
	year := t.Year()

	// Add ordinal suffix (1st, 2nd, 3rd, 4th, etc.)
	var suffix string
	if day >= 11 && day <= 13 {
		suffix = "th"
	} else {
		switch day % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		default:
			suffix = "th"
		}
	}

	return fmt.Sprintf("%d%s %s %d", day, suffix, month, year)
}

// formatDuration converts milliseconds to MM:SS format
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// formatNumber converts large numbers to readable format (1.2M, 15.3K)
func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// formatBool converts boolean to Yes/No string
func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// formatString handles empty strings with N/A fallback
func formatString(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
