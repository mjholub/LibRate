package member

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"codeberg.org/mjh/LibRate/db"
)

func (s *PgMemberStorage) Export(ctx context.Context, memberName, format string) (baseInfo []byte, otherData []byte, err error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		tx, err := s.newClient.BeginTx(ctx, pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to begin transaction: %v", err)
		}
		// nolint:errcheck // we don't care about the error here
		defer tx.Rollback(ctx)

		name := db.Sanitize([]string{memberName})[0]

		var id uuid.UUID
		var webfinger string

		// PERF: sub-optimal round trip of selecting the member twice, once to get the UUID and webfinger to use in unions and once to get the rest of the data
		err = s.newClient.QueryRow(ctx, `SELECT uuid, webfinger FROM public.members WHERE nick = $1`, name).Scan(&id, &webfinger)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get member ID: %v", err)
		}
		memberData, err := s.Read(ctx, webfinger)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get member data: %v", err)
		}
		uuidFuncs := []func(context.Context, pgx.Tx, uuid.UUID) (map[string]interface{}, error){
			s.exportBanInfo,
			s.exportPrefs,
			s.exportTrackRatings,
			s.exportBaseRatings,
			s.exportImages,
		}

		webfingerFuncs := []func(context.Context, pgx.Tx, string) (map[string]interface{}, error){
			s.exportBlocks,
			s.exportFollowRelationshipsOut,
			s.exportFollowRelationshipsIn,
			s.exportMediaContributions,
			s.exportArtistContributions,
		}

		// channel to collect the JSON output from the export functions
		dataChan := make(chan map[string]interface{}, len(uuidFuncs)+len(webfingerFuncs))
		finalData := make(chan []byte, 1)
		errChan := make(chan error, len(uuidFuncs)+len(webfingerFuncs))

		var wg sync.WaitGroup

		go func(
			context.Context,
			<-chan map[string]interface{},
			<-chan error,
			chan []byte,
			string,
		) {
			defer close(finalData)
			err = consumeData(ctx, dataChan, errChan, finalData, format)
		}(ctx, dataChan, errChan, finalData, format)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to process exported data: %v", err)
		}

		for _, f := range uuidFuncs {
			wg.Add(1)
			go func(ctx context.Context, tx pgx.Tx, id uuid.UUID) {
				defer wg.Done()
				data, err := f(ctx, tx, id)
				if err != nil {
					errChan <- err
				}
				dataChan <- data
			}(ctx, tx, id)
		}

		for _, f := range webfingerFuncs {
			wg.Add(1)
			go func(ctx context.Context, tx pgx.Tx, webfinger string) {
				defer wg.Done()
				data, err := f(ctx, tx, webfinger)
				if err != nil {
					errChan <- err
				}
				dataChan <- data
			}(ctx, tx, webfinger)
		}

		wg.Wait()
		close(dataChan)
		close(errChan)
		output := <-finalData
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process exported data: %v", err)
		}
		baseInfo, err := json.Marshal(memberData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal JSON: %v", err)
		}
		return baseInfo, output, nil
	}
}

func consumeData(ctx context.Context, dataChan <-chan map[string]interface{}, errChan <-chan error, output chan []byte, format string) error {
	combinedData := make(map[string]interface{})
	for {
		select {
		case data := <-dataChan:
			for k, v := range data {
				combinedData[k] = v
			}
		case err := <-errChan:
			return err
		case <-ctx.Done():
			switch format {
			case "json":
				result, err := json.Marshal(combinedData)
				if err != nil {
					close(output)
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				close(output)
				output <- result
				return nil
			case "csv":

				var buf bytes.Buffer
				w := csv.NewWriter(&buf)

				for k, v := range combinedData {
					w.Write([]string{k, v.(string)})
				}

				w.Flush()

				if err := w.Error(); err != nil {
					return fmt.Errorf("failed to write CSV: %v", err)
				}

				close(output)
				output <- buf.Bytes()
				return nil
			}
		}
	}
}

// PERF: benchmark whether it'd be optimal to scan, marshal to JSON,
// then unmarshal again to finally combine the output in the desired format
func (s *PgMemberStorage) exportBanInfo(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `SELECT reason, ends, can_appeal, mask, occurrence, started
	FROM public.bans WHERE member_uuid = $1`, memberID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportPrefs(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `SELECT * FROM public.member_prefs WHERE member_id = $1`, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportTrackRatings(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx,
		`SELECT 
    t."name" AS track_name, 
    a."name" AS album_name, 
    p.first_name || ' ' || p.last_name AS artist_name, 
    r.stars AS rating
FROM 
    reviews.track_ratings r
JOIN 
    media.tracks t ON r.track = t.media_id
JOIN 
    media.albums a ON t.album = a.media_id
JOIN 
    media.album_artists aa ON a.media_id = aa.album
JOIN 
    people.person p ON aa.artist = p.id
WHERE 
    r.user_id = $1 -- replace $1 with the user_id
ORDER BY 
    a."name", t.track_number;
`)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for track ratings export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportBaseRatings(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `
	SELECT m."title", m."kind" AS media_title, media_kind, 
		r.stars, r.body, r.topic, r.attribution FROM reviews.ratings r
	JOIN media.media m ON r.media_id = m.media_id
	WHERE r.user_id = $1`, memberID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for base ratings export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

// TODO: add cast review imports when cast voting is implemented

func (s *PgMemberStorage) exportBlocks(ctx context.Context, tx pgx.Tx, webfinger string) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `SELECT * FROM public.blocks WHERE blocker_webfinger = $1`, webfinger)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for block export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

// export followed users and sent follow requests
func (s *PgMemberStorage) exportFollowRelationshipsOut(ctx context.Context, tx pgx.Tx, webfinger string) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `SELECT f.followee, f.created, f.notifications, f.reblogs,
	frout.created, frout.reblogs, frout.notify, frout.target_webfinger,
	FROM public.followers f
	LEFT JOIN public.follow_requests frout ON f.follower = frout.requester_webfinger
	WHERE f.follower = $1`, webfinger)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for follow relationships export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportFollowRelationshipsIn(ctx context.Context, tx pgx.Tx, webfinger string) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `SELECT f.follower, f.created, frin.created, frin.requester_webfinger
FROM public.followers f
WHERE f.followee = $1
LEFT JOIN public.follow_requests frin ON f.followee = frin.target_webfinger`, webfinger)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for follow relationships export: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportImages(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `SELECT * FROM cdn.images WHERE uploader = $1`, memberID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for image export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportMediaContributions(ctx context.Context, tx pgx.Tx, webfinger string) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `
SELECT 
    m."title", 
    m."kind" AS media_title, 
  FROM 
	contributors.media cm,
JOIN 
    media.media m ON cm.media_id = m.id
WHERE 
    cm.contributor = $1
	`, webfinger)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for media contributions export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&output); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}

func (s *PgMemberStorage) exportArtistContributions(ctx context.Context, tx pgx.Tx, webfinger string) (output map[string]interface{}, err error) {
	rows, err := tx.Query(ctx, `
SELECT 
    CONCAT(p.first_name, ' ', p.last_name) AS name, 
    p.roles AS roles,
    g.name AS group_name,
    g.kind AS group_kind,
		s.name AS studio_name,
		s.kind AS studio_kind
FROM 
    contributors.person cp
JOIN 
    people.person p ON cp.person_id = p.id
JOIN 
    contributors."group" cg ON cp.contributor = cg.contributor
JOIN 
    people."group" g ON cg.group_id = g.id
JOIN
		contributors.studio cs ON cp.contributor = cs.contributor
JOIN
		people.studio s ON cs.studio_id = s.id
WHERE 
    cp.contributor = $1
	`, webfinger)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for artist contributions export: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
	}

	return output, nil
}
