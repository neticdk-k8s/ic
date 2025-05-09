## ic

Inventory CLI

### Synopsis

ic is a tool to manage das Inventar

```
ic [command] [flags]
```

### Options

```
      --log-format string                            Log format (plain|json) (default "plain")
      --log-level string                             Log level (debug|info|warn|error) (default "info")
  -o, --output string                                Output format (default "plain")
  -f, --force                                        Force actions
      --no-input                                     Assume non-interactive mode
      --no-color                                     Do not print color
  -d, --debug                                        Debug mode
      --no-headers                                   Do not print headers
  -s, --api-server string                            URL for the inventory server. (default "https://api.k8s.netic.dk")
      --oidc-auth-bind-addr string                   [authcode-browser] Bind address and port for local server used for OIDC redirect (default "localhost:18000")
      --oidc-client-id string                        OIDC client ID (default "inventory-cli")
      --oidc-grant-type string                       OIDC authorization grant type. One of (authcode-browser|authcode-keyboard) (default "authcode-browser")
      --oidc-issuer-url string                       Issuer URL for the OIDC Provider (default "https://keycloak.netic.dk/auth/realms/mcs")
      --oidc-redirect-uri-authcode-keyboard string   [authcode-keyboard] Redirect URI when using authcode keyboard (default "urn:ietf:wg:oauth:2.0:oob")
      --oidc-redirect-url-hostname string            [authcode-browser] Hostname of the redirect URL (default "localhost")
      --oidc-token-cache-dir string                  Directory used to store cached tokens (default "/Users/kn/Library/Caches/ic/oidc-login")
  -h, --help                                         help for ic
```

### SEE ALSO

* [ic api-token](ic_api-token.md)	 - Get access token for the API
* [ic completion](ic_completion.md)	 - Generate the autocompletion script for the specified shell
* [ic create](ic_create.md)	 - Create a resource
* [ic delete](ic_delete.md)	 - Delete a resource
* [ic filters](ic_filters.md)	 - About filters
* [ic get](ic_get.md)	 - Add one or many resources
* [ic login](ic_login.md)	 - Login to Inventory Server
* [ic logout](ic_logout.md)	 - Log out of Inventory Server
* [ic update](ic_update.md)	 - Update a resource

###### Auto generated by spf13/cobra on 17-Mar-2025
