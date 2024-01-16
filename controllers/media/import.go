package media

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"

	"github.com/zmb3/spotify/v2"
)

type ImportSource struct {
	Name string `json:"name" validate:"required"`
	URI  string `json:"uri" validate:"required"`
}

func (mc *Controller) GetImportSources(c *fiber.Ctx) error {
	return c.JSON(mc.conf.External.ImportSources)
}

// ImportWeb handles the import of media from 3rd party sources
// Importing from the local filesystem is handled by the client since
// 1. Uploading music files to the server would unnecessarily overload the server
// 2. Although processing ID3 tags may be faster on the server, one has to consider the
// round-trip time of the request, which would be much slower than processing the tags on the client
func (mc *Controller) ImportWeb(c *fiber.Ctx) error {
	var source ImportSource
	if err := c.BodyParser(&source); err != nil || !lo.Contains(mc.conf.External.ImportSources, source.Name) {
		return handleBadRequest(mc.storage.Log, c, "Invalid request body")
	}
	switch source.Name {
	case "spotify":
		return mc.importSpotify(c, source.URI)
	case "discogs":
		return mc.importDiscogs(c, source)
	case "lastfm":
		return mc.importLastFM(c, source)
	case "listenbrainz":
		return mc.importListenBrainz(c, source)
	case "bandcamp":
		return mc.importBandcamp(c, source)
	case "mediawiki":
		return mc.importMW(c, source)
	case "rym":
		return mc.importRYM(c, source)
	case "pitchfork":
		return mc.importPF(c, source)
	default:
		return handleBadRequest(mc.storage.Log, c, "Invalid source name")
	}
}

func (mc *Controller) importSpotify(c *fiber.Ctx, uri string) error {
	spotifyAlbumID := db.Sanitize([]string{strings.Split(uri, "/")[4]})[0]
	if len(spotifyAlbumID) != 22 {
		return handleBadRequest(mc.storage.Log, c, "Invalid Spotify album ID "+spotifyAlbumID)
	}

	spotifyConf := &clientcredentials.Config{
		ClientID:     mc.conf.External.SpotifyClientID,
		ClientSecret: mc.conf.External.SpotifyClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := spotifyConf.Token(c.Context())
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "failed to get Spotify API token", err)
	}

	httpClient := spotifyauth.New().Client(c.Context(), token)
	client := spotify.New(httpClient)

	spotifyAlbumData, err := client.GetAlbum(c.Context(), spotify.ID(spotifyAlbumID))
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "failed to get album from Spotify", err)
	}

	// NOTE: spotify doesn't differentiate groups and single artists, so we have to rely on our own data
	var artists models.AlbumArtist
	// ify there's is only one artist returned, we'll include that in the response. If not, we'll send an info to the client,
	// that would result in spawning a selection dialog to choose the correct artist
	var isUnambiguousResult bool
	for i := range spotifyAlbumData.Artists {
		individual, group, err := mc.storage.Ps.GetArtistsByName(c.Context(), spotifyAlbumData.Artists[i].Name)
		if err != nil {
			return handleInternalError(mc.storage.Log, c, "failed to get artist from database", err)
		}
		artists.PersonArtists = append(artists.PersonArtists, individual...)
		artists.GroupArtists = append(artists.GroupArtists, group...)
	}
	if len(artists.PersonArtists)+len(artists.GroupArtists) == 1 {
		isUnambiguousResult = true
	}

	genres := make([]models.Genre, len(spotifyAlbumData.Genres))
	for i := range spotifyAlbumData.Genres {
		genre, err := mc.storage.GetGenre(c.Context(), "music", "en", spotifyAlbumData.Genres[i])
		if err != nil {
			return handleInternalError(mc.storage.Log, c, "failed to get genre from database", err)
		}
		genres = append(genres, *genre)
	}

	tracks := make([]models.Track, len(spotifyAlbumData.Tracks.Tracks))

	sTracks := spotifyAlbumData.Tracks.Tracks

	for i := range sTracks {
		duration := time.Now().Add(sTracks[i].TimeDuration())
		track := models.Track{
			Name:     sTracks[i].Name,
			Duration: duration,
			Lyrics:   "",
			Number:   int16(sTracks[i].TrackNumber),
		}
		tracks = append(tracks, track)
	}

	// sum the duration of all tracks
	albumDuration := time.Duration(0)

	for i := range tracks {
		albumDuration += tracks[i].Duration.Sub(time.Time{})
	}

	album := models.Album{
		Name:        spotifyAlbumData.Name,
		ReleaseDate: spotifyAlbumData.ReleaseDateTime(),
		Genres:      genres,
		Duration:    sql.NullTime{Time: time.Now().Add(albumDuration), Valid: true},
		Tracks:      tracks,
	}
	if isUnambiguousResult {
		album.AlbumArtists = artists
		// TODO: add returning the image as a blob together with the album
		return c.JSON(album)
	} else {
		return c.JSON(fiber.Map{
			"album":   album,
			"artists": artists,
		})
	}
}

func (mc *Controller) importDiscogs(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *Controller) importLastFM(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *Controller) importListenBrainz(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *Controller) importBandcamp(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *Controller) importMW(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *Controller) importRYM(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (mc *Controller) importPF(c *fiber.Ctx, source ImportSource) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
