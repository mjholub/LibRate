package media

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"

	"github.com/zmb3/spotify/v2"
)

type ImportSource struct {
	// Could be 'web' for a web source or 'fs' for a local filesystem source
	Kind         string             `json:"kind" validate:"required,oneof=web fs"`
	Name         string             `json:"name" validate:"required"`
	HasAPI       bool               `json:"has_api"`
	URI          string             `json:"uri" validate:"uri"`
	URIValidator func(string) error `json:"-"`
}

// ImportWeb handles the import of media from 3rd party sources
func (mc *Controller) ImportWeb(c *fiber.Ctx) error {
	var source ImportSource
	if err := c.BodyParser(&source); err != nil || !lo.Contains(mc.conf.ImportSources, source.Name) {
		return handleBadRequest(mc.storage.Log, c, "Invalid request body")
	}
	switch source.Name {
	case "spotify":
		return mc.importSpotify(c, source)
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

func (mc *Controller) importSpotify(c *fiber.Ctx, source ImportSource) error {
	spotifyAlbumID := db.Sanitize(strings.Split(c.Params("import_url"), "/")[4])
	if len(spotifyAlbumID) != 22 {
		return handleBadRequest(mc.storage.Log, c, "Invalid Spotify album ID "+spotifyAlbumID)
	}

	authorization := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", mc.conf.External.SpotifyClientID, mc.conf.External.SpotifyClientSecret))))

	req := fiber.Get(fmt.Sprintf("https://api.spotify.com/v1/albums/%s", spotifyAlbumID))
	req.Set("Authorization", authorization)
	req.Set("Content-Type", "application/json")

	status, body, errs := req.Bytes()
	if errs != nil {
		var aggrError string
		for i := range errs {
			aggrError += errs[i].Error() + "\n"
		}
		return handleInternalError(mc.storage.Log, c, "failed to get album from Spotify API", fmt.Errorf(aggrError))
	}

	if status != fiber.StatusOK {
		return handleInternalError(mc.storage.Log, c, "failed to get album from Spotify API", fmt.Errorf("status code %d", status))
	}

	// generic map since we don't want to deal with the maintenance of a spotify dependency
	var spotifyAlbumData spotify.FullAlbum
	if err := json.Unmarshal(body, &spotifyAlbumData); err != nil {
		return handleInternalError(mc.storage.Log, c, "failed to unmarshal Spotify album data", err)
	}

	// NOTE: spotify doesn't differentiate groups and single artists, so we have to rely on our own data
	var artists models.AlbumArtist
	// ify theres is only one artist returned, we'll include that in the response. If not, we'll send an info to the client,
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

	var genres []models.Genre
	for i := range spotifyAlbumData.Genres {
		genre, err := mc.storage.GetGenre(c.Context(), "music", "en", spotifyAlbumData.Genres[i])
		if err != nil {
			return handleInternalError(mc.storage.Log, c, "failed to get genre from database", err)
		}
		genres = append(genres, *genre)
	}

	var tracks []models.Track

	for _, track := range spotifyAlbumData.Tracks.Tracks {
		duration := time.Now().Add(track.TimeDuration())
		track := models.Track{
			Name:     track.Name,
			Duration: duration,
			Lyrics:   "",
			Number:   int16(track.TrackNumber),
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
