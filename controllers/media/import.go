package media

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models/media"

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

	client, err := mc.authenticateSpotify(c.Context())
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "failed to authenticate with Spotify", err)
	}

	spotifyAlbumData, err := client.GetAlbum(c.Context(), spotify.ID(spotifyAlbumID))
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "failed to get album from Spotify", err)
	}

	// NOTE: spotify doesn't differentiate groups and single artists, so we have to rely on our own data
	var artists []media.AlbumArtist
	// ify there's is only one artist returned, we'll include that in the response. If not, we'll send an info to the client,
	// that would result in spawning a selection dialog to choose the correct artist
	remoteArtistNames, isUnambiguousResult := listRemoteArtists(artists, spotifyAlbumData.Artists)

	artists, err = mc.lookupSpotifyArtists(c, spotifyAlbumData.Artists)
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "failed to get artist from database", err)
	}

	genres := make([]media.Genre, len(spotifyAlbumData.Genres))
	for i := range spotifyAlbumData.Genres {
		genre, err := mc.storage.GetGenre(c.Context(), "music", "en", spotifyAlbumData.Genres[i])
		if err != nil {
			return handleInternalError(mc.storage.Log, c, "failed to get genre from database", err)
		}
		genres = append(genres, *genre)
	}

	tracks := make([]media.Track, len(spotifyAlbumData.Tracks.Tracks))

	sTracks := spotifyAlbumData.Tracks.Tracks

	for i := range sTracks {
		duration := time.Now().Add(sTracks[i].TimeDuration())
		track := media.Track{
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

	album := media.Album{
		Name:        spotifyAlbumData.Name,
		ReleaseDate: spotifyAlbumData.ReleaseDateTime(),
		Genres:      genres,
		Duration:    sql.NullTime{Time: time.Now().Add(albumDuration), Valid: true},
		Tracks:      tracks,
	}

	if len(remoteArtistNames) > 0 {
		return c.JSON(fiber.Map{
			"remote_artists": remoteArtistNames,
			"album":          album,
		})
	}

	if !isUnambiguousResult {
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

// listRemoteArtists enumerates artists present in the import source
// but not locally
// TODO: modify remoteArtists to accept map[string]string
func listRemoteArtists(localArtists []media.AlbumArtist, remoteArtists []spotify.SimpleArtist) (remoteArtistNames []string, unambiguous bool) {
	if len(localArtists) == 1 {
		return nil, true
	}

	if len(localArtists) == 0 || len(remoteArtists) > len(localArtists) {
		switch len(localArtists) {
		case 0:
			for i := range remoteArtists {
				remoteArtistNames = append(remoteArtistNames, remoteArtists[i].Name)
			}
		default:
			for j := range localArtists {
				for k := range remoteArtists {
					if localArtists[j].Name != remoteArtists[k].Name {
						remoteArtistNames = append(remoteArtistNames, remoteArtists[k].Name)
					}
				}
			}
		}
	}

	return remoteArtistNames, false
}

func (mc *Controller) authenticateSpotify(ctx context.Context) (*spotify.Client, error) {
	spotifyConf := &clientcredentials.Config{
		ClientID:     mc.conf.External.SpotifyClientID,
		ClientSecret: mc.conf.External.SpotifyClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := spotifyConf.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	return spotify.New(httpClient), nil
}

// lookupSpotifyArtists checks if the artists imported from Spotify are already in the database
func (mc *Controller) lookupSpotifyArtists(c *fiber.Ctx, artists []spotify.SimpleArtist) (dbArtists []media.AlbumArtist, err error) {
	for i := range artists {

		// PERF: reduce the number of fields retrieved from the database to only the ones needed
		individual, group, err := mc.storage.Ps.GetArtistsByName(c.Context(), artists[i].Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get artist from database: %w", err)
		}

		for j := range individual {
			fullName := fmt.Sprintf("%s \"%+v\" %s", individual[i].FirstName, individual[i].NickNames, individual[i].LastName)
			individualArtistEntry := media.AlbumArtist{
				ID:         individual[j].ID,
				Name:       fullName,
				ArtistType: "individual",
			}
			dbArtists = append(dbArtists, individualArtistEntry)
		}

		for k := range group {
			groupArtistEntry := media.AlbumArtist{
				ID:         group[k].ID,
				Name:       group[k].Name,
				ArtistType: "group",
			}
			dbArtists = append(dbArtists, groupArtistEntry)
		}
	}

	return dbArtists, nil
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
