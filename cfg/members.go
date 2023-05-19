package cfg

type AuthConfig struct {
	AuthMountPoint     string `yaml:"auth_mount_point", default:"/auth"`
	AuthCookieName     string `yaml:"auth_cookie_name", default:"auth"`
	AuthCookiePath     string `yaml:"auth_cookie_path", default:"/"`
	AuthCookieSecure   bool   `yaml:"auth_cookie_secure", default:"false"`
	AuthCookieHTTPOnly bool   `yaml:"auth_cookie_http_only", default:"true"`
	AuthCookieDomain   string `yaml:"auth_cookie_domain", default:""`
	AuthCookieExpire   int    `yaml:"auth_cookie_expire", default:"3600"`
}
