package http_test

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mhttp "github.com/mathieumaf/go-monero/pkg/http"
)

func TestParseChallenge(t *testing.T) {
	t.Parallel()

	const monerod018 = `Digest qop="auth",algorithm=MD5,realm="monero-rpc",` +
		`nonce="90phNlK30V9QcgtzvkYBBg==",stale=false`

	tests := []struct {
		name       string
		input      string
		wantQop    string
		wantAlgo   string
		wantRealm  string
		wantNonce  string
		wantStale  string
		wantOpaque string
		wantDomain string
		wantErr    string
	}{
		{
			name:      "monerod 0.18 challenge",
			input:     monerod018,
			wantQop:   "auth",
			wantAlgo:  "MD5",
			wantRealm: "monero-rpc",
			wantNonce: "90phNlK30V9QcgtzvkYBBg==",
			wantStale: "false",
		},
		{
			name: "legacy fixture",
			input: `Digest qop="auth",algorithm=MD5,realm="monero-rpc",` +
				`nonce="IdDHjxbfpLYP/KzjaxaOqA==",stale=false`,
			wantQop:   "auth",
			wantAlgo:  "MD5",
			wantRealm: "monero-rpc",
			wantNonce: "IdDHjxbfpLYP/KzjaxaOqA==",
			wantStale: "false",
		},
		{
			name: "any order, opaque, unknown parameters",
			input: `Digest charset=UTF-8, nonce="n+/=", stale=true, ` +
				`opaque="opq", userhash=false, realm="monero-rpc", ` +
				`algorithm=MD5, qop="auth", domain="/json_rpc"`,
			wantQop:    "auth",
			wantAlgo:   "MD5",
			wantRealm:  "monero-rpc",
			wantNonce:  "n+/=",
			wantStale:  "true",
			wantOpaque: "opq",
			wantDomain: "/json_rpc",
		},
		{
			name: "qop list selects auth",
			input: `Digest realm="monero-rpc", qop="auth-int, auth", ` +
				`nonce="n", stale=false`,
			wantQop:   "auth",
			wantAlgo:  "MD5",
			wantRealm: "monero-rpc",
			wantNonce: "n",
			wantStale: "false",
		},
		{
			name: "qop list without spaces and quoted algorithm",
			input: `Digest realm="monero-rpc", qop="auth,auth-int", ` +
				`nonce="n", algorithm="md5"`,
			wantQop:   "auth",
			wantAlgo:  "MD5",
			wantRealm: "monero-rpc",
			wantNonce: "n",
		},
		{
			name:      "unquoted qop and algorithm omitted",
			input:     `Digest realm="monero-rpc", qop=auth, nonce="n"`,
			wantQop:   "auth",
			wantAlgo:  "MD5",
			wantRealm: "monero-rpc",
			wantNonce: "n",
		},
		{
			name: "spaces around separators and escaped opaque",
			input: `Digest realm = "monero-rpc" , qop = "auth" , ` +
				`nonce = "n" , opaque = "a\"b" , algorithm = MD5`,
			wantQop:    "auth",
			wantAlgo:   "MD5",
			wantRealm:  "monero-rpc",
			wantNonce:  "n",
			wantOpaque: `a"b`,
		},
		{
			name: "case insensitive names and stale",
			input: `digest Qop="AUTH", Algorithm=Md5, Realm="monero-rpc", ` +
				`Nonce="n", Stale=TRUE`,
			wantQop:   "auth",
			wantAlgo:  "MD5",
			wantRealm: "monero-rpc",
			wantNonce: "n",
			wantStale: "TRUE",
		},
		{
			name:    "trailing comma",
			input:   `Digest realm="monero-rpc", qop="auth", nonce="n",`,
			wantQop: "auth", wantAlgo: "MD5",
			wantRealm: "monero-rpc", wantNonce: "n",
		},
		{
			name:    "missing digest scheme",
			input:   `Basic realm="monero-rpc"`,
			wantErr: "Digest",
		},
		{
			name:    "missing realm",
			input:   `Digest qop="auth", nonce="n", algorithm=MD5`,
			wantErr: "missing realm",
		},
		{
			name:    "missing nonce",
			input:   `Digest qop="auth", realm="monero-rpc", algorithm=MD5`,
			wantErr: "missing nonce",
		},
		{
			name:    "missing qop",
			input:   `Digest realm="monero-rpc", nonce="n", algorithm=MD5`,
			wantErr: "missing qop",
		},
		{
			name:    "qop without auth",
			input:   `Digest realm="monero-rpc", qop="auth-int", nonce="n"`,
			wantErr: "does not offer auth",
		},
		{
			name:    "unsupported algorithm",
			input:   `Digest realm="monero-rpc", qop="auth", nonce="n", algorithm=SHA-256`,
			wantErr: "unsupported algorithm 'SHA-256'",
		},
		{
			name:    "md5-sess is not implemented",
			input:   `Digest realm="monero-rpc", qop="auth", nonce="n", algorithm=MD5-sess`,
			wantErr: "unsupported algorithm 'MD5-sess'",
		},
		{
			name:    "unterminated quote",
			input:   `Digest realm="monero-rpc", qop="auth", nonce="n`,
			wantErr: "unterminated quoted string",
		},
		{
			name:    "no parameters",
			input:   `Digest`,
			wantErr: "no parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			challenge, err := mhttp.ParseChallenge(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantQop, challenge.Qop)
			assert.Equal(t, tt.wantAlgo, challenge.Algorithm)
			assert.Equal(t, tt.wantRealm, challenge.Realm)
			assert.Equal(t, tt.wantNonce, challenge.Nonce)
			assert.Equal(t, tt.wantStale, challenge.Stale)
			assert.Equal(t, tt.wantOpaque, challenge.Opaque)
			assert.Equal(t, tt.wantDomain, challenge.Domain)
		})
	}
}

func TestRoundTripChallengeVariations(t *testing.T) {
	t.Parallel()

	const (
		user  = "rpc"
		pass  = "secret"
		realm = "monero-rpc"
		nonce = "90phNlK30V9QcgtzvkYBBg=="
	)

	monerod := `Digest qop="auth",algorithm=MD5,realm="monero-rpc",` +
		`nonce="` + nonce + `",stale=false`
	sess := `Digest qop="auth",algorithm=MD5-sess,realm="monero-rpc",` +
		`nonce="` + nonce + `",stale=false`

	tests := []struct {
		name       string
		headers    []string
		wantOpaque string
		wantQop    string
		wantAlgo   string
	}{
		{
			name:       "monerod 0.18 advertises md5 then md5-sess",
			headers:    []string{monerod, sess},
			wantQop:    "auth",
			wantAlgo:   "MD5",
			wantOpaque: "",
		},
		{
			name:       "md5-sess before md5 is skipped",
			headers:    []string{sess, monerod},
			wantQop:    "auth",
			wantAlgo:   "MD5",
			wantOpaque: "",
		},
		{
			name: "basic challenge is ignored when digest is present",
			headers: []string{
				`Basic realm="proxy"`,
				monerod,
			},
			wantQop:  "auth",
			wantAlgo: "MD5",
		},
		{
			name: "qop list, opaque, lowercase algorithm, unknown param",
			headers: []string{
				`Digest charset=UTF-8, realm="monero-rpc", ` +
					`qop="auth, auth-int", nonce="` + nonce + `", ` +
					`opaque="rpc-opaque", algorithm=md5, stale=false`,
			},
			wantOpaque: "rpc-opaque",
			wantQop:    "auth",
			wantAlgo:   "MD5",
		},
		{
			name: "algorithm omitted defaults to md5 and opaque absent",
			headers: []string{
				`Digest realm="monero-rpc", qop="auth", nonce="` + nonce + `", stale=false`,
			},
			wantQop:  "auth",
			wantAlgo: "MD5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newDigestServer(t, digestServerConfig{
				user:    user,
				pass:    pass,
				realm:   realm,
				nonce:   nonce,
				headers: tt.headers,
			})
			defer srv.Close()

			body := []byte(`{"jsonrpc":"2.0","id":"0","method":"get_version"}`)
			req, err := http.NewRequestWithContext(
				t.Context(), http.MethodPost, srv.URL()+"/json_rpc",
				strings.NewReader(string(body)),
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rt := mhttp.NewDigestAuthTransport(user, pass, srv.Client().Transport)
			resp, err := rt.RoundTrip(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			payload, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode, string(payload))
			assert.Equal(t, `{"ok":true}`, string(payload))

			authed := srv.authorized()
			require.Len(t, authed, 1)
			assert.Equal(t, tt.wantOpaque, authed[0].opaque)
			assert.Equal(t, tt.wantQop, authed[0].qop)
			assert.Equal(t, tt.wantAlgo, authed[0].algorithm)
			assert.Equal(t, "00000001", authed[0].nc)
			assert.Equal(t, string(body), authed[0].body)
			assert.Equal(t, 2, srv.hits(), "handshake is challenge then authorization")
		})
	}
}

func TestRoundTripUnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("WWW-Authenticate",
			`Digest qop="auth",algorithm=SHA-256,realm="monero-rpc",nonce="n",stale=false`)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "unauthorized")
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodGet, srv.URL+"/json_rpc", nil,
	)
	require.NoError(t, err)

	rt := mhttp.NewDigestAuthTransport("rpc", "secret", srv.Client().Transport)
	resp, err := rt.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unsupported algorithm 'SHA-256'")
}

type digestAuthObservation struct {
	opaque    string
	qop       string
	algorithm string
	nc        string
	nonce     string
	body      string
	uri       string
	cnonce    string
	response  string
}

type digestServerConfig struct {
	user    string
	pass    string
	realm   string
	nonce   string
	headers []string
}

// digestServer emulates the digest challenge monerod returns for --rpc-login.
type digestServer struct {
	cfg    digestServerConfig
	server *httptest.Server

	mu    sync.Mutex
	hitsN int
	ok    []digestAuthObservation
	why   string
}

func newDigestServer(t *testing.T, cfg digestServerConfig) *digestServer {
	t.Helper()

	ds := &digestServer{cfg: cfg}
	ds.server = httptest.NewServer(ds)
	return ds
}

func (s *digestServer) Close() { s.server.Close() }

func (s *digestServer) URL() string { return s.server.URL }

func (s *digestServer) Client() *http.Client { return s.server.Client() }

func (s *digestServer) hits() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hitsN
}

func (s *digestServer) authorized() []digestAuthObservation {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]digestAuthObservation, len(s.ok))
	copy(out, s.ok)
	return out
}

func (s *digestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.hitsN++

	auth := r.Header.Get("Authorization")
	if auth == "" {
		s.writeChallenge(w)
		return
	}

	obs, ok := s.verify(r, auth, string(body))
	if !ok {
		s.writeChallenge(w)
		return
	}

	s.ok = append(s.ok, obs)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"ok":true}`)
}

func (s *digestServer) writeChallenge(w http.ResponseWriter) {
	for _, header := range s.cfg.headers {
		w.Header().Add("WWW-Authenticate", header)
	}
	w.WriteHeader(http.StatusUnauthorized)
	msg := "unauthorized"
	if s.why != "" {
		msg = s.why
	}
	_, _ = io.WriteString(w, msg)
}

func (s *digestServer) verify(
	r *http.Request, auth, body string,
) (digestAuthObservation, bool) {
	params, err := mhttp.ParseDigestParams(auth)
	if err != nil {
		s.why = err.Error()
		return digestAuthObservation{}, false
	}

	obs := digestAuthObservation{
		opaque:    params["opaque"],
		qop:       params["qop"],
		algorithm: params["algorithm"],
		nc:        params["nc"],
		nonce:     params["nonce"],
		body:      body,
		uri:       params["uri"],
		cnonce:    params["cnonce"],
		response:  params["response"],
	}

	if obs.nonce != s.cfg.nonce || obs.uri != r.URL.RequestURI() {
		s.why = "nonce or uri mismatch"
		return obs, false
	}
	if obs.qop != "auth" {
		s.why = "qop " + obs.qop
		return obs, false
	}

	ha1 := md5Hex(s.cfg.user + ":" + s.cfg.realm + ":" + s.cfg.pass)
	ha2 := md5Hex(r.Method + ":" + obs.uri)
	want := md5Hex(strings.Join(
		[]string{ha1, obs.nonce, obs.nc, obs.cnonce, obs.qop, ha2}, ":",
	))
	if !strings.EqualFold(obs.response, want) {
		s.why = "digest mismatch"
		return obs, false
	}

	return obs, true
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s)) //nolint:gosec
	return hex.EncodeToString(sum[:])
}
