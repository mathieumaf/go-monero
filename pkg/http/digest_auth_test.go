package http_test

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	user          string
	pass          string
	realm         string
	nonce         string
	headers       []string
	qop           string
	algorithm     string
	opaque        string
	omitAlgorithm bool
	sess          bool
}

// digestServer emulates the digest challenge monerod returns for --rpc-login.
type digestServer struct {
	cfg    digestServerConfig
	server *httptest.Server

	mu        sync.Mutex
	nonce     string
	nonceGen  int
	hitsN     int
	ok        []digestAuthObservation
	bodies    []string
	why       string
	staleOnce bool
	reject    bool
	bare      bool
	sawStale  bool
	markStale bool
}

func newDigestServer(t *testing.T, cfg digestServerConfig) *digestServer {
	t.Helper()

	ds := &digestServer{cfg: cfg, nonce: cfg.nonce}
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

func (s *digestServer) requestBodies() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.bodies))
	copy(out, s.bodies)
	return out
}

func (s *digestServer) armStale() {
	s.mu.Lock()
	s.staleOnce = true
	s.mu.Unlock()
}

func (s *digestServer) setReject(reject bool) {
	s.mu.Lock()
	s.reject = reject
	s.mu.Unlock()
}

func (s *digestServer) setBare(bare bool) {
	s.mu.Lock()
	s.bare = bare
	s.mu.Unlock()
}

func (s *digestServer) sawStaleNonce() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sawStale
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
	s.bodies = append(s.bodies, string(body))

	if s.bare {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "unauthorized")
		return
	}

	if s.mustChallenge(r) {
		s.writeChallenge(w, s.markStale)
		s.markStale = false
		return
	}

	obs, ok := s.verify(r, r.Header.Get("Authorization"), string(body))
	if !ok {
		s.writeChallenge(w, false)
		return
	}

	s.ok = append(s.ok, obs)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"ok":true}`)
}

// mustChallenge reports whether this request should be answered with 401.
// The caller holds s.mu.
func (s *digestServer) mustChallenge(r *http.Request) bool {
	if s.reject || r.Header.Get("Authorization") == "" {
		return true
	}
	if !s.staleOnce {
		return false
	}

	s.staleOnce = false
	s.rotateNonce()
	s.markStale = true
	s.sawStale = true
	return true
}

func (s *digestServer) rotateNonce() {
	s.nonceGen++
	switch s.nonceGen {
	case 1:
		s.nonce = "P4lbGCQYPa9IkjAMvlgMHw=="
	case 2:
		s.nonce = "HvSmkDOsva9+baX7mqizfA=="
	default:
		s.nonce = fmt.Sprintf("nonce-%d", s.nonceGen)
	}
}

func (s *digestServer) writeChallenge(w http.ResponseWriter, stale bool) {
	headers := s.cfg.headers
	if len(headers) == 0 {
		headers = s.buildHeaders(stale)
	}
	for _, header := range headers {
		w.Header().Add("WWW-Authenticate", header)
	}
	w.WriteHeader(http.StatusUnauthorized)
	msg := "unauthorized"
	if s.why != "" {
		msg = s.why
	}
	_, _ = io.WriteString(w, msg)
}

func (s *digestServer) buildHeaders(stale bool) []string {
	staleVal := "false"
	if stale {
		staleVal = "true"
	}
	qop := s.cfg.qop
	if qop == "" {
		qop = "auth"
	}

	line := fmt.Sprintf(
		`Digest qop="%s",realm="%s",nonce="%s",stale=%s`,
		qop, s.cfg.realm, s.nonce, staleVal,
	)
	if !s.cfg.omitAlgorithm {
		algo := s.cfg.algorithm
		if algo == "" {
			algo = "MD5"
		}
		line += ",algorithm=" + algo
	}
	if s.cfg.opaque != "" {
		line += `,opaque="` + s.cfg.opaque + `"`
	}

	headers := []string{line}
	if s.cfg.sess {
		headers = append(headers, fmt.Sprintf(
			`Digest qop="auth",algorithm=MD5-sess,realm="%s",nonce="%s",stale=%s`,
			s.cfg.realm, s.nonce, staleVal,
		))
	}
	return headers
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

	if obs.nonce != s.nonce || obs.uri != r.URL.RequestURI() {
		s.why = fmt.Sprintf("nonce or uri mismatch: got nonce %q want %q uri %q",
			obs.nonce, s.nonce, obs.uri)
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

func TestDigestAuthReusesNonce(t *testing.T) {
	t.Parallel()

	const (
		user  = "rpc"
		pass  = "secret"
		realm = "monero-rpc"
		nonce = "90phNlK30V9QcgtzvkYBBg=="
	)

	tests := []struct {
		name string
		cfg  digestServerConfig
	}{
		{
			name: "monerod md5 and md5-sess",
			cfg: digestServerConfig{
				user: user, pass: pass, realm: realm, nonce: nonce,
				qop: "auth", algorithm: "MD5", sess: true,
			},
		},
		{
			name: "qop list opaque and lowercase algorithm",
			cfg: digestServerConfig{
				user: user, pass: pass, realm: realm, nonce: nonce,
				qop: "auth, auth-int", algorithm: "md5", opaque: "rpc-opaque",
			},
		},
		{
			name: "algorithm omitted",
			cfg: digestServerConfig{
				user: user, pass: pass, realm: realm, nonce: nonce,
				qop: "auth", omitAlgorithm: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newDigestServer(t, tt.cfg)
			t.Cleanup(srv.Close)
			rt := mhttp.NewDigestAuthTransport(user, pass, srv.Client().Transport)
			payload := `{"jsonrpc":"2.0","id":"0","method":"get_version"}`

			postDigest(t, rt, srv.URL()+"/json_rpc", payload)
			assert.Equal(t, 2, srv.hits())

			postDigest(t, rt, srv.URL()+"/json_rpc", payload)
			assert.Equal(t, 3, srv.hits())

			postDigest(t, rt, srv.URL()+"/json_rpc", payload)
			assert.Equal(t, 4, srv.hits())

			authed := srv.authorized()
			require.Len(t, authed, 3)
			assert.Equal(t, []string{"00000001", "00000002", "00000003"}, []string{
				authed[0].nc, authed[1].nc, authed[2].nc,
			})
			for _, obs := range authed {
				assert.Equal(t, nonce, obs.nonce)
				assert.Equal(t, "auth", obs.qop)
				assert.Equal(t, "MD5", obs.algorithm)
				assert.Equal(t, tt.cfg.opaque, obs.opaque)
				assert.Equal(t, payload, obs.body)
			}
			for _, body := range srv.requestBodies() {
				assert.Equal(t, payload, body)
			}
		})
	}
}

func TestDigestAuthStaleNonceRetriesOnce(t *testing.T) {
	t.Parallel()

	const (
		user    = "rpc"
		pass    = "secret"
		realm   = "monero-rpc"
		nonce   = "90phNlK30V9QcgtzvkYBBg=="
		renewed = "P4lbGCQYPa9IkjAMvlgMHw=="
	)

	srv := newDigestServer(t, digestServerConfig{
		user: user, pass: pass, realm: realm, nonce: nonce,
		qop: "auth", algorithm: "MD5", opaque: "opq", sess: true,
	})
	t.Cleanup(srv.Close)
	rt := mhttp.NewDigestAuthTransport(user, pass, srv.Client().Transport)
	url := srv.URL() + "/json_rpc"
	payload := `{"jsonrpc":"2.0","id":"0","method":"get_info"}`

	postDigest(t, rt, url, payload)
	srv.armStale()

	before := srv.hits()
	postDigest(t, rt, url, payload)
	assert.Equal(t, 2, srv.hits()-before, "stale nonce is one rejection plus one retry")
	assert.True(t, srv.sawStaleNonce())

	authed := srv.authorized()
	require.Len(t, authed, 2)
	assert.Equal(t, nonce, authed[0].nonce)
	assert.Equal(t, "00000001", authed[0].nc)
	assert.Equal(t, renewed, authed[1].nonce)
	assert.Equal(t, "00000001", authed[1].nc)
	assert.Equal(t, "opq", authed[1].opaque)

	before = srv.hits()
	postDigest(t, rt, url, payload)
	assert.Equal(t, 1, srv.hits()-before)

	authed = srv.authorized()
	require.Len(t, authed, 3)
	assert.Equal(t, renewed, authed[2].nonce)
	assert.Equal(t, "00000002", authed[2].nc)
}

func TestDigestAuth401RetriesOnce(t *testing.T) {
	t.Parallel()

	srv := newDigestServer(t, digestServerConfig{
		user: "rpc", pass: "secret", realm: "monero-rpc",
		nonce: "90phNlK30V9QcgtzvkYBBg==",
		qop:   "auth", algorithm: "MD5",
	})
	t.Cleanup(srv.Close)
	rt := mhttp.NewDigestAuthTransport("rpc", "secret", srv.Client().Transport)
	url := srv.URL() + "/json_rpc"

	postDigest(t, rt, url, `{"method":"get_version"}`)
	srv.setReject(true)

	for range 2 {
		before := srv.hits()
		resp := postDigestStatus(t, rt, url, `{"method":"get_version"}`)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 2, srv.hits()-before)
	}
}

func TestDigestAuthBare401DropsCache(t *testing.T) {
	t.Parallel()

	srv := newDigestServer(t, digestServerConfig{
		user: "rpc", pass: "secret", realm: "monero-rpc",
		nonce: "90phNlK30V9QcgtzvkYBBg==",
		qop:   "auth", algorithm: "MD5",
	})
	t.Cleanup(srv.Close)
	rt := mhttp.NewDigestAuthTransport("rpc", "secret", srv.Client().Transport)
	url := srv.URL() + "/json_rpc"

	postDigest(t, rt, url, `{"method":"get_version"}`)
	srv.setBare(true)

	before := srv.hits()
	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, url, strings.NewReader(`{"method":"get_version"}`),
	)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WWW-Authenticate")
	assert.Equal(t, 1, srv.hits()-before, "no challenge means no second try")

	srv.setBare(false)
	before = srv.hits()
	postDigest(t, rt, url, `{"method":"get_version"}`)
	assert.Equal(t, 2, srv.hits()-before, "the nonce was dropped, so the next call handshakes")
}

func TestDigestAuthConcurrentNonceReuse(t *testing.T) {
	t.Parallel()

	const n = 32
	srv := newDigestServer(t, digestServerConfig{
		user: "rpc", pass: "secret", realm: "monero-rpc",
		nonce: "90phNlK30V9QcgtzvkYBBg==",
		qop:   "auth", algorithm: "MD5", sess: true,
	})
	t.Cleanup(srv.Close)
	rt := mhttp.NewDigestAuthTransport("rpc", "secret", srv.Client().Transport)
	url := srv.URL() + "/json_rpc"

	postDigest(t, rt, url, `{"n":"warmup"}`)

	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()

			payload := fmt.Sprintf(`{"n":%d}`, i)
			req, err := http.NewRequestWithContext(
				context.Background(), http.MethodPost, url, strings.NewReader(payload),
			)
			if err != nil {
				errCh <- err
				return
			}
			resp, err := rt.RoundTrip(req)
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()
			if _, err := io.Copy(io.Discard, resp.Body); err != nil {
				errCh <- err
				return
			}
			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("status %d for %s", resp.StatusCode, payload)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	assert.Equal(t, 2+n, srv.hits())
	authed := srv.authorized()
	require.Len(t, authed, n+1)

	seenNC := map[uint64]struct{}{}
	seenBody := map[string]int{}
	for _, obs := range authed {
		nc, err := strconv.ParseUint(obs.nc, 16, 32)
		require.NoError(t, err)
		_, dup := seenNC[nc]
		assert.False(t, dup, "nc %s reused", obs.nc)
		seenNC[nc] = struct{}{}
		seenBody[obs.body]++
		assert.Equal(t, "90phNlK30V9QcgtzvkYBBg==", obs.nonce)
	}
	require.Len(t, seenNC, n+1)
	for i := 1; i <= n+1; i++ {
		_, ok := seenNC[uint64(i)]
		assert.True(t, ok, "missing nc %d", i)
	}
	assert.Equal(t, 1, seenBody[`{"n":"warmup"}`])
	for i := range n {
		assert.Equal(t, 1, seenBody[fmt.Sprintf(`{"n":%d}`, i)])
	}
}

func TestDigestAuthReplaysBody(t *testing.T) {
	t.Parallel()

	const payload = `{"jsonrpc":"2.0","id":"0","method":"get_version"}`
	srv := newDigestServer(t, digestServerConfig{
		user: "rpc", pass: "secret", realm: "monero-rpc",
		nonce: "90phNlK30V9QcgtzvkYBBg==",
		qop:   "auth", algorithm: "MD5",
	})
	t.Cleanup(srv.Close)
	rt := mhttp.NewDigestAuthTransport("rpc", "secret", srv.Client().Transport)
	url := srv.URL() + "/json_rpc"

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, url, strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Body = io.NopCloser(failReader{})
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(got))

	req, err = http.NewRequestWithContext(t.Context(), http.MethodPost, url, nil)
	require.NoError(t, err)
	req.Body = io.NopCloser(bytes.NewReader([]byte(payload)))
	req.ContentLength = int64(len(payload))
	req.GetBody = nil
	resp, err = rt.RoundTrip(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	got, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(got))

	assert.Equal(t, 3, srv.hits(), "second call reuses the nonce")
	for _, body := range srv.requestBodies() {
		assert.Equal(t, payload, body)
	}
}

func TestDigestAuthNonceIsPerHost(t *testing.T) {
	t.Parallel()

	const (
		user  = "rpc"
		pass  = "secret"
		realm = "monero-rpc"
	)
	srvA := newDigestServer(t, digestServerConfig{
		user: user, pass: pass, realm: realm, nonce: "nonce-a",
		qop: "auth", algorithm: "MD5",
	})
	srvB := newDigestServer(t, digestServerConfig{
		user: user, pass: pass, realm: realm, nonce: "nonce-b",
		qop: "auth", algorithm: "MD5",
	})
	t.Cleanup(srvA.Close)
	t.Cleanup(srvB.Close)

	rt := mhttp.NewDigestAuthTransport(user, pass, &http.Transport{})
	postDigest(t, rt, srvA.URL()+"/json_rpc", `{"host":"a"}`)
	postDigest(t, rt, srvB.URL()+"/json_rpc", `{"host":"b"}`)
	postDigest(t, rt, srvA.URL()+"/json_rpc", `{"host":"a"}`)
	postDigest(t, rt, srvB.URL()+"/json_rpc", `{"host":"b"}`)

	assert.Equal(t, 3, srvA.hits())
	assert.Equal(t, 3, srvB.hits())
	assert.Equal(t, "nonce-a", srvA.authorized()[1].nonce)
	assert.Equal(t, "00000002", srvA.authorized()[1].nc)
	assert.Equal(t, "nonce-b", srvB.authorized()[1].nonce)
	assert.Equal(t, "00000002", srvB.authorized()[1].nc)
}

func TestDigestTransportLeavesOpenResponsesAlone(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get("Authorization") != "" {
			http.Error(w, "unexpected authorization", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "open")
	}))
	t.Cleanup(srv.Close)

	rt := mhttp.NewDigestAuthTransport("rpc", "secret", srv.Client().Transport)
	for range 2 {
		resp := postDigestStatus(t, rt, srv.URL+"/json_rpc", `{"method":"get_version"}`)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	assert.Equal(t, int32(2), hits.Load())
}

func postDigest(t *testing.T, rt http.RoundTripper, url, payload string) {
	t.Helper()

	resp := postDigestStatus(t, rt, url, payload)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
}

func postDigestStatus(
	t *testing.T, rt http.RoundTripper, url, payload string,
) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, url, strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	return resp
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) {
	return 0, errors.New("body must be replayed with GetBody")
}
