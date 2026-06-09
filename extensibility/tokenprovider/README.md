### Token Provider

This sample shows how to inject an `auth.TokenProvider` so the server attaches
a bearer token to every outbound cross-cluster (replication) RPC. The receiver
side validates the token via the existing Authorizer/ClaimMapper pipeline —
this sample is only about the sender.

### Implementation shape

The Temporal Service calls `TokenProvider.GetToken` on every outbound
cross-cluster RPC and does not cache the result. Caching, expiry-aware
refresh, and deduplication of concurrent fetches are the provider's
responsibility.

The included sample structures itself like a network-backed provider:

- `fileTokenSource` implements [`oauth2.TokenSource`](https://pkg.go.dev/golang.org/x/oauth2#TokenSource).
  It reads the JWT from disk and reports the token's `exp` claim as the
  `oauth2.Token.Expiry`.
- `fileTokenProvider` wraps that source in [`oauth2.ReuseTokenSource`](https://pkg.go.dev/golang.org/x/oauth2#ReuseTokenSource),
  which hands back the cached token until it's close to expiry and only then
  calls back into the underlying source.

**This is deliberately overengineered for a file source** — local reads are
cheap and a per-RPC re-read would be fine. The point is that swapping the
underlying source for a real OAuth2 token source (for example
[`clientcredentials.Config.TokenSource`](https://pkg.go.dev/golang.org/x/oauth2/clientcredentials#Config.TokenSource))
requires no other changes; the cache, the `TokenProvider` adapter, and the
server wiring all stay the same. Caching is load-bearing once an IdP round
trip sits behind the source.

### Steps to run this sample

1. Start up the dependencies by running the `make start-dependencies` command from within the main Temporal repository as described in the [contribution guide](https://github.com/temporalio/temporal/blob/master/CONTRIBUTING.md#run-temporal-server-locally).

2. Create the database schema by running `make install-schema-cass-es`.

3. Write a bearer token to a file and point the server at it:
   ```
   echo -n "my-bearer-token" > /tmp/token
   export TOKEN_FILE=/tmp/token
   ```

4. Start Temporal by running `go run tokenprovider/server/main.go`.

   The token attaches to outbound replication dials only; single-cluster traffic
   on `localhost:7233` is unaffected. To see the wiring exercised, register a
   second cluster as a peer and create a global namespace.

### Configuration knobs

The TokenProvider integration adds one YAML key under
`global.authorization.remoteClusterAuth` (shipped by Temporal Server, not
this sample):

| Field | Purpose |
|---|---|
| `require` | Fail boot if no TokenProvider is configured; fail outbound RPCs that get an empty token. |

`global.tls.remoteClusters` (or `WithTLSConfigProvider`) must also be
configured — the credential layer requires transport security and refuses to
attach a bearer token to a plaintext dial. The test config under
`server/testdata/config.yaml` has a stub entry that demonstrates the minimum
required to satisfy boot validation.

### Token rotation behaviour

Because tokens are cached by expiry, rotation works as follows:

- **JWT tokens** rotate naturally: each new file contents has its own `exp`
  claim, and the cache picks up the new value on the next refresh.
- **Opaque tokens** carry no expiry information, so the cache holds them
  indefinitely. If you need to swap an opaque token without restarting the
  process, replace the file *and* extend `fileTokenSource` to inspect the
  file's modification time, or restart.
