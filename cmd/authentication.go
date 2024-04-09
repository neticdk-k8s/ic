package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type authenticationOptions struct {
	OIDCIssuerURL                   string
	OIDCClientID                    string
	OIDCGrantType                   string
	OIDCRedirectURLHostname         string
	OIDCRedirectURIAuthCodeKeyboard string
	OIDCAuthBindAddr                string
	OIDCTokenCacheDir               string
}

func (o *authenticationOptions) addFlags(f *pflag.FlagSet) {
	f.StringVar(&o.OIDCIssuerURL, "oidc-issuer-url", "http://localhost:8080/realms/test", "Issuer URL for the OIDC Provider")
	_ = viper.BindPFlag("oidc-issuer-url", f.Lookup("oidc-issuer-url"))

	f.StringVar(&o.OIDCClientID, "oidc-client-id", "inventory-cli", "OIDC client ID")
	_ = viper.BindPFlag("oidc-client-id", f.Lookup("oidc-client-id"))

	f.StringVar(&o.OIDCGrantType, "oidc-grant-type", "authcode-browser", "OIDC authorization grant type. One of (authcode-browser|authcode-keyboard)")
	_ = viper.BindPFlag("oidc-grant-type", f.Lookup("oidc-grant-type"))

	f.StringVar(&o.OIDCRedirectURLHostname, "oidc-redirect-url-hostname", "localhost", "[authcode-browser] Hostname of the redirect URL")
	_ = viper.BindPFlag("oidc-redirect-url-hostname", f.Lookup("oidc-redirect-url-hostname"))

	f.StringVar(&o.OIDCAuthBindAddr, "oidc-auth-bind-addr", "localhost:18000", "[authcode-browser] Bind address and port for local server used for OIDC redirect")
	_ = viper.BindPFlag("oidc-auth-bind-addr", f.Lookup("oidc-auth-bind-addr"))

	f.StringVar(&o.OIDCRedirectURIAuthCodeKeyboard, "oidc-redirect-uri-authcode-keyboard", oobRedirectURI, "[authcode-keyboard] Redirect URI when using authcode keyboard")
	_ = viper.BindPFlag("oidc-redirect-uri-authcode-keyboard", f.Lookup("oidc-redirect-uri-authcode-keyboard"))

	f.StringVar(&o.OIDCTokenCacheDir, "oidc-token-cache-dir", getDefaultTokenCacheDir(), "Directory used to store cached tokens")
	_ = viper.BindPFlag("oidc-token-cache-dir", f.Lookup("oidc-token-cache-dir"))
}

func getDefaultTokenCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "ic", "oidc-login")
	}
	return filepath.Join(cacheDir, "ic", "oidc-login")
}
