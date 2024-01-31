package auth

import (
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/member"
	"codeberg.org/mjh/LibRate/tests"
)

func prepare() (*Service, *fiber.App, *member.Member) {
	mockService := Service{
		conf: &cfg.Config{
			JWTSecret: "test-secret",
		},
		sess: session.New(),
	}

	memberData := &member.Member{
		MemberName: "John Doe",
		Webfinger:  "john@example.com",
	}

	logger := zerolog.Nop()
	app := tests.NewAppWithLogger(&logger)

	app.Get("/test", func(c *fiber.Ctx) error {
		s, err := mockService.sess.Get(c)
		if err != nil {
			return err
		}
		return c.JSON(s)
	})

	return &mockService, app, memberData
}

func BenchmarkCreateToken(b *testing.B) {
	mockService, app, memberData := prepare()

	timeout := time.Minute

	// make a request to create a session

	resData, err := makeRequest(app)
	if err != nil {
		b.Fatalf("Failed to make request: %s", err.Error())
	}

	// unmarshal the response into sess
	sess := new(session.Session)

	err = json.Unmarshal(resData, sess)

	if err != nil {
		b.Fatalf("Failed to create session: %s", err.Error())
	}

	for i := 0; i < b.N; i++ {
		mockService.createToken(memberData, &timeout, sess)
	}
}

func makeRequest(app *fiber.App) ([]byte, error) {
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, err
	}
	resData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return resData, nil
}

func BenchmarkSetSessionKeys(b *testing.B) {
	_, app, memberData := prepare()

	res, err := makeRequest(app)
	if err != nil {
		b.Fatalf("Failed to make request: %s", err.Error())
	}
	sess := new(session.Session)
	err = json.Unmarshal(res, sess)
	if err != nil {
		b.Fatalf("Failed to create session: %s", err.Error())
	}
	ip := "42.13.21.37"
	ua := "Mozilla/5.0 (Windows NT 10.0; rv:121.0) Gecko/20100101 Firefox/121.0"

	for i := 0; i < b.N; i++ {
		setSessionKeysSequential(ip, ua, sess, memberData)
	}
}
