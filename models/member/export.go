package member

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"codeberg.org/mjh/LibRate/db"
)

func (s *PgMemberStorage) Export(ctx context.Context, memberName, format string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// first, query the database for all the tables referencing the member
		/* INFO: ban history
		* public.bans WHERE member_uuid = $1
		* INFO: preferences
		* public.member_prefs WHERE  member_id = $1
		* INFO: reviews
		* reviews.cast_ratings WHERE user_id = $1
		* reviews.track_ratings WHERE user_id = $1
		* reviews.ratings WHERE user_id = $1
		* INFO: users that the member has blocked
		* public.member_blocks WHERE requester_webfinger = member.webfinger
		* INFO: in that order: followees, followers, received follow requests, sent follow requests
		* public.followers(followee, created, notifications, reblogs) WHERE follower = member.webfinger
		* public.followers(follower) WHERE followee = member.webfinger
		* public.follow_requests(requester_webfinger, created) WHERE target_webfinger = member.webfinger
		* public.follow_requests(target_webfinger, created, reblogs, notify) WHERE requester_webfinger = member.webfinger
		* INFO: uploaded images
		* * FROM cdn.images WHERE uploader = member.nick OR uploader = member.webfinger
		 */
		// TODO: when we add tracking of contributions, also scan the media schema for all
		// rows where added_by = member.webfinger

		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %v", err)
		}
		// nolint:errcheck // we don't care about the error here
		defer tx.Rollback(ctx)

		name := db.Sanitize([]string{memberName})[0]

		var uuid uuid.UUID
		var webfinger string

		// PERF: sub-optimal round trip of selecting the member twice, once to get the UUID and webfinger to use in unions and once to get the rest of the data
		err = s.newClient.QueryRow(ctx, `SELECT uuid, webfinger FROM public.members WHERE nick = $1`, name).Scan(&uuid, &webfinger)
		// exclude uuid and the admin/mod name from ban history
		// return track ratings in a human-readable format that includes the corresponding album and artist name
		_, err = tx.Prepare(ctx, "member_union", `
			SELECT * FROM public.members WHERE nick = $1 
			UNION ALL
			SELECT reason, ends, can_appeal, mask, occurrence, started FROM public.bans WHERE member_uuid = $2
			UNION ALL
			SELECT * FROM public.member_prefs WHERE member_id = $2
			UNION ALL
			SELECT * FROM reviews.cast_ratings WHERE user_id = $2 
			UNION ALL
SELECT 
    a."name" AS album_name, 
    t."name" AS track_name, 
    p."name" AS artist_name, 
    r.stars AS rating
FROM 
    reviews.track_ratings r
JOIN 
    media.album_tracks at ON r.track = at.track
JOIN 
    media.albums a ON at.album = a.media_id
JOIN 
    media.tracks t ON r.track = t.media_id
JOIN 
    media.album_artists aa ON a.media_id = aa.album
JOIN 
    people.person p ON aa.artist = p.id
WHERE 
    r.user_id = $2 
ORDER BY 
    a."name", at.track_number;
			UNION ALL
			SELECT * FROM reviews.ratings WHERE user_id = $2 AS ratings
			JOIN 
				media.media m ON ratings.media_id = m.media_id
			UNION ALL
			SELECT * FROM public.member_blocks WHERE requester_webfinger = $3
			UNION ALL
			SELECT followee, created, notifications, reblogs FROM public.followers WHERE follower = $3
			AS followed_members
			UNION ALL
			SELECT follower FROM public.followers WHERE followee = $3 AS followers
			UNION ALL
			SELECT requester_webfinger, created FROM public.follow_requests WHERE target_webfinger = $3 AS received_follow_requests 
			UNION ALL
			SELECT target_webfinger, created, reblogs, notify FROM public.follow_requests WHERE requester_webfinger = $3 AS sent_follow_requests
			UNION ALL
			SELECT * FROM cdn.images WHERE uploader = $1 OR uploader = $3 AS uploaded_images
			`)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %v", err)
		}

		rows, err := tx.Query(ctx, "member_union", name, uuid, webfinger)
		if err != nil {
			return nil, fmt.Errorf("failed to execute statement: %v", err)
		}
		defer rows.Close()
		var output map[string]interface{}
		for rows.Next() {
			err = rows.Scan(&output)
			if err != nil {
				return nil, fmt.Errorf("failed to scan row: %v", err)
			}
		}

		switch format {
		case "json":
			return json.Marshal(output)
		case "csv":
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)

			for k, v := range output {
				w.Write([]string{k, v.(string)})
			}

			w.Flush()

			return buf.Bytes(), nil
		default:
			return nil, fmt.Errorf("invalid format: %s", format)
		}

	}
}
