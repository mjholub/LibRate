package version

import (
	"encoding/json"
	"os"

	h "codeberg.org/mjh/LibRate/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// Get retrieves the version of the frontend
func Get(c *fiber.Ctx) error {
	f, err := os.Open("fe/package.json")
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "error reading package.json")
	}
	defer f.Close()

	var pkg struct {
		Version string `json:"version"`
	}

	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "error parsing package.json")
	}

	return h.Res(c, fiber.StatusOK, pkg.Version)
}
