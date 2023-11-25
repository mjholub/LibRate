// This file contains the implementation of the MemberStorage interface for Neo4j.
package member

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Save stores a new member in the database.
func (s Neo4jMemberStorage) Save(ctx context.Context, member *Member) error {
	params := map[string]interface{}{
		"uuid":          member.UUID,
		"passhash":      member.PassHash,
		"nick":          member.MemberName,
		"email":         member.Email,
		"reg_timestamp": member.RegTimestamp.Unix(),
		"active":        true,
		// TODO: convert pq.StringArray to neo4j compatible type
		"roles": nil,
	}

	s.log.Debug().Msgf("params: %v", params)

	session := s.client.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := neo4j.ExecuteWrite[int64](ctx, session, func(tx neo4j.ManagedTransaction) (int64, error) {
		result, err := tx.Run(ctx,
			`CREATE (m:Member) SET m = $params`, map[string]interface{}{"params": params})
		if err != nil {
			return 0, err
		}

		_, err = result.Consume(ctx)
		if err != nil {
			return 0, err
		}

		check, err := tx.Run(ctx, "MATCH (m:Member) RETURN count(m) AS count", nil)
		if err != nil {
			return 0, err
		}
		record, err := check.Single(ctx)
		if err != nil {
			return 0, err
		}
		count, ok := record.Get("count")
		if !ok {
			return 0, fmt.Errorf("failed to get count")
		}
		if count, ok := count.(int64); ok {
			return count, nil
		}
		return 0, fmt.Errorf("failed to convert count to int64")
	})
	if err != nil {
		return fmt.Errorf("failed to save member: %v", err)
	}
	if result == 0 {
		return errors.New("failed to save member: no records created")
	}
	return nil
}

func (s Neo4jMemberStorage) Update(ctx context.Context, member *Member) error {
	session := s.client.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := neo4j.ExecuteWrite[int64](ctx, session, func(tx neo4j.ManagedTransaction) (int64, error) {
		result, err := tx.Run(ctx,
			`MATCH (m:Member {uuid: $uuid}) SET m = $params`, map[string]interface{}{"uuid": member.UUID, "params": member})
		if err != nil {
			return 0, err
		}
		_, err = result.Consume(ctx)
		if err != nil {
			return 0, err
		}
		return 1, nil
	})
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}
	if result == 0 {
		return errors.New("failed to update member: no records updated")
	}
	return nil
}

func (s Neo4jMemberStorage) Delete(ctx context.Context, member *Member) error {
	session := s.client.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := neo4j.ExecuteWrite[int64](ctx, session, func(tx neo4j.ManagedTransaction) (int64, error) {
		result, err := tx.Run(ctx,
			`MATCH (m:Member {uuid: $uuid}) DELETE m`, map[string]interface{}{"uuid": member.UUID})
		if err != nil {
			return 0, err
		}
		_, err = result.Consume(ctx)
		if err != nil {
			return 0, err
		}
		return 1, nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete member: %v", err)
	}
	if result == 0 {
		return errors.New("failed to delete member: no records deleted")
	}
	return nil
}

func (s Neo4jMemberStorage) Read(ctx context.Context, keyName, key string) (*Member, error) {
	session := s.client.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := neo4j.ExecuteRead[*Member](ctx, session, func(tx neo4j.ManagedTransaction) (*Member, error) {
		result, err := tx.Run(ctx,
			"MATCH (m:Member) WHERE m."+keyName+" = $key RETURN m", map[string]interface{}{"key": key})
		if err != nil {
			return nil, fmt.Errorf("failed to read member: %v", err)
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read member: %v", err)
		}
		member, ok := record.Get("m")
		if !ok {
			return nil, fmt.Errorf("failed to read member: %v", err)
		}
		memberNode, ok := member.(neo4j.Node)
		if !ok {
			return nil, fmt.Errorf("expected result to be a map, but was %T", member)
		}
		err = mapstructure.Decode(memberNode.Props, &member)
		if err != nil {
			return nil, fmt.Errorf("failed to decode member: %v", err)
		}
		return member.(*Member), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read member: %v", err)
	}
	return result, nil
}

func (s Neo4jMemberStorage) CreateSession(ctx context.Context, m *Member) (t string, err error) {
	return "", errors.New("not implemented")
}

func (s Neo4jMemberStorage) GetID(ctx context.Context, key string) (uint32, error) {
	return 0, errors.New("not implemented")
}

func (s Neo4jMemberStorage) GetPassHash(email, login string) (string, error) {
	return "", errors.New("not implemented")
}

func (s Neo4jMemberStorage) RequestFollow(ctx context.Context, fr *FollowRequest) error {
	return errors.New("not implemented")
}

func (s Neo4jMemberStorage) Close() error {
	return s.client.Close(context.Background())
}

func (s Neo4jMemberStorage) GetSessionTimeout(
	ctx context.Context, memberID int, deviceID uuid.UUID,
) (timeout int, err error) {
	return 0, fmt.Errorf("GetSessionTimeout not implemented yet")
}

func (s Neo4jMemberStorage) LookupDevice(ctx context.Context, deviceID uuid.UUID) error {
	return fmt.Errorf("LookupDevice not implemented yet")
}

func (s Neo4jMemberStorage) Check(ctx context.Context, email, nick string) (bool, error) {
	return false, fmt.Errorf("Check not implemented yet for neo4j")
}
