package target

func ValidateCategory(s string) bool {
	switch s {
	case "users", "artists", "media", "groups", "tags", "posts", "reviews", "genres", "union":
		return true
	default:
		return false
	}
}
