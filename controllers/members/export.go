package members

import (
	"archive/zip"
	"bytes"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// @Summary Export all of the member's data
// @Description Exports the data of a member, including profile information as well as other related data such as reviews
// @Tags accounts,members,metadata
// @Accept json
// @Produce json text/csv
// @Param Authorization header string true "JWT access token"
// @Param format path string true "Export format" Enums(json, csv)
// @Router /members/export/{format} [get]
func (mc *Controller) Export(c *fiber.Ctx) error {
	memberName := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["member_name"].(string)
	if memberName == "" {
		return h.BadRequest(mc.log, c, "missing name in JWT token", "export request initialized without a token from "+c.IP(), nil)
	}
	format := c.Params("format")
	availableFormats := []string{"json", "csv"}
	if !lo.Contains(availableFormats, format) {
		return h.BadRequest(
			mc.log,
			c,
			"invalid format",
			fmt.Sprintf("Member %s (IP: %s) tried to initialize export with invalid format '%s'", memberName, c.IP(), format), nil)
	}
	mc.log.Info().Msgf("%s initialized a data export request using %s format", memberName, format)

	baseData, extraData, err := mc.storage.Export(c.Context(), memberName, format)
	if err != nil {
		return h.InternalError(mc.log, c, fmt.Sprintf("failed to export data for %s using %s", memberName, format), err)
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	const ExportError = "error while creating the exported data file"
	baseDataFile, err := w.Create("basicinfo." + format)
	if err != nil {
		return h.InternalError(mc.log, c, ExportError, err)
	}
	_, err = baseDataFile.Write(baseData)
	if err != nil {
		return h.InternalError(mc.log, c, ExportError, err)
	}
	additionalDataFile, err := w.Create("extra" + format)
	if err != nil {
		return h.InternalError(mc.log, c, ExportError, err)
	}
	_, err = additionalDataFile.Write(extraData)
	if err != nil {
		return h.InternalError(mc.log, c, ExportError, err)
	}
	if err = w.Close(); err != nil {
		return h.InternalError(mc.log, c, ExportError, fmt.Errorf("critical error: failed to close writer: %w", err))
	}

	date := c.Request().Header.Peek("Date")
	if date == nil {
		// equivalent to Date().toISOString() in JS
		date = []byte(time.Now().UTC().Format(time.RFC3339))
	}
	_, err = time.Parse(time.RFC3339, string(date))

	if err != nil {
		// request header manipulation here usually indicates a bad actor, so we'll log their info
		serverMessage := fmt.Sprintf("bad Date header was sent by %s, headers: %v, body: %s", c.IP(), c.GetReqHeaders(), string(c.Body()))
		return h.BadRequest(mc.log, c, "Bad export request", serverMessage, err)
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.zip", memberName, string(date)))

	c.Set("Content-Type", "archive/zip")
	return c.Send(buf.Bytes())
}
