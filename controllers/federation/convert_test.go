package federation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"codeberg.org/mjh/LibRate/models/member"
	"codeberg.org/mjh/LibRate/tests"

	"github.com/go-ap/activitypub"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestMemberToActor(t *testing.T) {
	logger := zerolog.Nop()
	app := tests.NewAppWithLogger(&logger)

	port, err := tests.TryFindFreePort()
	require.Nil(t, err)

	app.Post("/member-to-actor", MemberToActorTestHandler)
	go func() {
		err = app.Listen(fmt.Sprintf(":%d", port))
		require.Nil(t, err)
	}()

	// pass the test data as a multipart form to avoid dealing with all required fields
	// nullable types etc.
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	writer.WriteField("member_name", "test")
	writer.WriteField("display_name", "Test")
	writer.WriteField("public_key_pem", "test")

	require.NotEmpty(t, buf)
	require.NoErrorf(t, writer.Close(), "error closing multipart writer: %s", err)

	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:%d/member-to-actor", port), buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)

	require.NoErrorf(t, err, "error testing request: %s", err)
	if resp.StatusCode != 200 {
		// get the Message property from the response
		messageRaw, err := io.ReadAll(resp.Body)
		assert.NoErrorf(t, err, "error reading response body: %s", err)
		require.Equalf(t, string(messageRaw), "error", "error message")
		_, err = resp.Body.Read(messageRaw)
		assert.NoErrorf(t, err, "error reading response body: %s", err)
		assert.NotEmpty(t, messageRaw)
		type responseBody struct {
			Error string `json:"message"`
		}
		r := new(responseBody)
		err = json.Unmarshal(messageRaw, &r)
		require.NoErrorf(t, err, "error unmarshalling response body: %s", err)
		// get the error by purposefully wrong assertion
		assert.Equalf(t, "error", r.Error, "error message")
	}

	assert.Equal(t, 200, resp.StatusCode)
	//	var actor *activitypub.Actor
	body, err := io.ReadAll(resp.Body)
	require.NoErrorf(t, err, "error reading response body: %s", err)
	require.NotEmpty(t, body)
	require.NotZero(t, len(body))
	jsonVal, err := fastjson.ParseBytes(body)
	require.NoErrorf(t, err, "error parsing response body: %s", err)
	require.NotNil(t, jsonVal)
	fmt.Printf("jsonVal: %+v\n", jsonVal)

	// FIXME: either upstream issue or my misunderstanding of the library
	//	err = activitypub.JSONLoadActor(obj., actor)
	//	require.NoErrorf(t, err, "error decoding actor: %s", err)
	addr := fmt.Sprintf("http://localhost:%d/api/members/test", port)
	//	require.NotNil(t, actor)
	assert.Equalf(t, activitypub.IRI(addr).String(), string(jsonVal.GetStringBytes("id")), "actor id")
	assert.Equalf(t, activitypub.IRI(addr+"/inbox").String(), string(jsonVal.GetStringBytes("inbox")), "actor inbox")
	assert.Equalf(t, activitypub.IRI(addr+"/outbox").String(), string(jsonVal.GetStringBytes("outbox")), "actor outbox")
	assert.Equalf(t, activitypub.NaturalLanguageValues{
		activitypub.DefaultLangRef("Test"),
	}.String(), string(jsonVal.GetStringBytes("preferredUsername")), "actor preferred username")
	//assert.Equalf(t, &activitypub.Endpoints{
	//		SharedInbox: activitypub.IRI(fmt.Sprintf("http://localhost:%d/api/inbox", port)),
	//}, string(jsonVal.GetStringBytes("endpoints")), "actor endpoints")
	// 	assert.Equalf(t, activitypub.PublicKey{
	//	ID:           activitypub.IRI(addr + "#main-key"),
	//	Owner:        activitypub.IRI(addr),
	//	PublicKeyPem: "test",
	//}.ID.String(), string(jsonVal.GetStringBytes("publicKey")), "actor public key id")
}

func MemberToActorTestHandler(c *fiber.Ctx) error {
	c.Accepts("multipart/form-data")
	var m member.Member
	m.MemberName = c.FormValue("member_name")
	m.DisplayName.String = c.FormValue("display_name")
	m.DisplayName.Valid = true
	m.PublicKeyPem = c.FormValue("public_key_pem")
	log := zerolog.Nop()
	ch := ConversionHandler{
		log: &log,
	}

	memberAsActor, err := ch.MemberToActor(c, &m)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(200).Send(memberAsActor)
}
