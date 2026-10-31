// This is a derived work based on `code.google.com/p/mlab-ns2/gae/ns/digest`
// (original work of Bipasa Chattopadhyay bipasa@cs.unc.edu Eric Gavaletz
// gavaletz@gmail.com Seon-Wook Park seon.wook@swook.net, from the fork
// maintained by Bob Ziuchkovski @bobziuchkovski
// (https://github.com/rkl-/digest).
package http

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// md5Algorithm is the only digest algorithm this client implements.
// RFC 7616 defaults a missing algorithm to MD5.
const md5Algorithm = "MD5"

// DigestAuthTransport is an implementation of http.RoundTripper that takes
// care of http digest authentication.
type DigestAuthTransport struct {
	Username string
	Password string
	rt       http.RoundTripper
}

// NewDigestAuthTransport creates a new digest transport using the
// http.DefaultTransport.
func NewDigestAuthTransport(
	username, password string, rt http.RoundTripper,
) *DigestAuthTransport {
	return &DigestAuthTransport{
		Username: username,
		Password: password,
		rt:       rt,
	}
}

func (t *DigestAuthTransport) newCredentials(
	req *http.Request, c *challenge,
) *credentials {
	return &credentials{
		Algorithm:  c.Algorithm,
		DigestURI:  req.URL.RequestURI(),
		MessageQop: c.Qop, // "auth" must be a single value
		Nonce:      c.Nonce,
		NonceCount: 0,
		Opaque:     c.Opaque,
		Realm:      c.Realm,
		Username:   t.Username,

		method:   req.Method,
		password: t.Password,
	}
}

// RoundTrip makes a request expecting a 401 response that will require digest
// authentication. It creates the credentials it needs and makes a follow-up
// request.
func (t *DigestAuthTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	finalRequest := req.Clone(req.Context())

	if req.Body != nil {
		bodyContents, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read all body: %w", err)
		}

		reqBody01 := io.NopCloser(bytes.NewBuffer(bodyContents))
		reqBody02 := io.NopCloser(bytes.NewBuffer(bodyContents))

		req.Body = reqBody01
		finalRequest.Body = reqBody02
	}

	// make a request to get the 401 that contains the challenge.
	//
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("round trip err: %w", err)
	}
	if resp.StatusCode != 401 { // cool, reached what we needed
		return resp, nil
	}

	// we must ensure that the initial response has been totally drained
	// otherwise the http client won't reuse the connection.
	//
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return nil, fmt.Errorf("copy body to null dev: %w", err)
	}
	resp.Body.Close()

	chal, err := parseWWWAuthenticate(resp.Header)
	if err != nil {
		return nil, fmt.Errorf("parse challenge: %w", err)
	}

	cr := t.newCredentials(finalRequest, chal)
	auth, err := cr.authorize()
	if err != nil {
		return nil, fmt.Errorf("authorize: %w", err)
	}

	finalRequest.Header.Set("Authorization", auth)
	return t.rt.RoundTrip(finalRequest)
}

type challenge struct {
	Realm     string
	Domain    string
	Nonce     string
	Opaque    string
	Stale     string
	Algorithm string
	Qop       string
}

// parseWWWAuthenticate selects the first Digest challenge this client can
// satisfy. monerod advertises MD5 and MD5-sess as separate header values;
// unsupported schemes and algorithms are skipped when another challenge
// works.
func parseWWWAuthenticate(header http.Header) (*challenge, error) {
	values := header.Values("WWW-Authenticate")
	if len(values) == 0 {
		return nil, fmt.Errorf("missing WWW-Authenticate header")
	}

	var firstErr error
	for _, value := range values {
		chal, err := parseChallenge(value)
		if err == nil {
			return chal, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return nil, firstErr
}

// parseChallenge reads one WWW-Authenticate Digest challenge.
//
// Parameters may appear in any order and in any number. Unknown parameters
// are ignored. qop may be a list; auth is selected when it is offered.
// algorithm is optional, defaults to MD5, and is compared case-insensitively.
func parseChallenge(input string) (*challenge, error) {
	params, err := parseDigestParams(input)
	if err != nil {
		return nil, fmt.Errorf("parse challenge fields: %w", err)
	}

	c := &challenge{Algorithm: md5Algorithm}
	for key, value := range params {
		applyChallengeParam(c, key, value)
	}

	if err := validateChallenge(c); err != nil {
		return nil, err
	}

	return c, nil
}

func applyChallengeParam(c *challenge, key, value string) {
	switch key {
	case "realm":
		c.Realm = value
	case "domain":
		c.Domain = value
	case "nonce":
		c.Nonce = value
	case "opaque":
		c.Opaque = value
	case "stale":
		c.Stale = value
	case "algorithm":
		c.Algorithm = value
	case "qop":
		c.Qop = value
	}
}

func validateChallenge(c *challenge) error {
	if c.Realm == "" {
		return fmt.Errorf("challenge missing realm")
	}
	if c.Nonce == "" {
		return fmt.Errorf("challenge missing nonce")
	}

	algo, err := normalizeAlgorithm(c.Algorithm)
	if err != nil {
		return err
	}
	c.Algorithm = algo

	qop, err := selectQop(c.Qop)
	if err != nil {
		return err
	}
	c.Qop = qop

	return nil
}

func normalizeAlgorithm(raw string) (string, error) {
	if raw == "" || strings.EqualFold(raw, md5Algorithm) {
		return md5Algorithm, nil
	}

	return "", fmt.Errorf("unsupported algorithm '%s'", raw)
}

// selectQop picks auth from a qop list. auth-int is not implemented.
func selectQop(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("challenge missing qop")
	}

	for _, part := range strings.Split(raw, ",") {
		if strings.EqualFold(strings.TrimSpace(part), "auth") {
			return "auth", nil
		}
	}

	return "", fmt.Errorf("qop '%s' does not offer auth", raw)
}

func parseDigestParams(input string) (map[string]string, error) {
	rest, err := stripDigestPrefix(input)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for rest != "" {
		var key, value string
		key, value, rest, err = readDigestParam(rest)
		if err != nil {
			return nil, err
		}
		params[strings.ToLower(strings.TrimSpace(key))] = value
		rest = strings.TrimLeft(rest, " \t\r\n,")
	}

	if len(params) == 0 {
		return nil, fmt.Errorf("challenge has no parameters")
	}

	return params, nil
}

func stripDigestPrefix(input string) (string, error) {
	const prefix = "digest"

	s := strings.TrimSpace(input)
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", fmt.Errorf("bad challenge: input doesn't start with 'Digest'")
	}

	rest := s[len(prefix):]
	if rest != "" && !isDigestDelim(rest[0]) {
		return "", fmt.Errorf("bad challenge: input doesn't start with 'Digest'")
	}

	return strings.TrimLeft(rest, " \t\r\n,"), nil
}

func isDigestDelim(b byte) bool {
	switch b {
	case ' ', '\t', '\r', '\n', ',':
		return true
	default:
		return false
	}
}

func readDigestParam(s string) (key, value, rest string, err error) {
	key, rest, err = readDigestKey(s)
	if err != nil {
		return "", "", "", err
	}

	if strings.HasPrefix(rest, `"`) {
		value, rest, err = readQuotedDigestValue(rest)
		return key, value, rest, err
	}

	value, rest = readTokenDigestValue(rest)
	return key, value, rest, nil
}

func readDigestKey(s string) (string, string, error) {
	key, rest, ok := strings.Cut(s, "=")
	if !ok {
		return "", "", fmt.Errorf("challenge parameter missing '=': '%s'",
			truncate(s))
	}

	key = strings.TrimSpace(key)
	if key == "" || strings.ContainsAny(key, " \t\",") {
		return "", "", fmt.Errorf("invalid challenge parameter '%s'", key)
	}

	rest = strings.TrimLeft(rest, " \t")
	if rest == "" {
		return "", "", fmt.Errorf("challenge parameter '%s' missing value", key)
	}

	return key, rest, nil
}

func readQuotedDigestValue(rest string) (string, string, error) {
	value, next, err := unquoteDigest(rest)
	if err != nil {
		return "", "", err
	}

	next = strings.TrimLeft(next, " \t")
	if next != "" && next[0] != ',' {
		return "", "", fmt.Errorf("quoted challenge value has trailing data")
	}

	return value, next, nil
}

func readTokenDigestValue(rest string) (string, string) {
	end := strings.IndexByte(rest, ',')
	if end < 0 {
		return strings.TrimSpace(rest), ""
	}

	return strings.TrimSpace(rest[:end]), rest[end:]
}

func unquoteDigest(s string) (string, string, error) {
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 >= len(s) {
				return "", "", fmt.Errorf("unterminated escape in challenge")
			}
			i++
			b.WriteByte(s[i])
		case '"':
			return b.String(), s[i+1:], nil
		default:
			b.WriteByte(s[i])
		}
	}

	return "", "", fmt.Errorf("unterminated quoted string in challenge")
}

func truncate(s string) string {
	const maxLen = 40
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "..."
}

type credentials struct {
	Algorithm  string
	Cnonce     string
	DigestURI  string
	MessageQop string
	Nonce      string
	NonceCount int
	Opaque     string
	Realm      string
	Username   string

	method   string
	password string
}

func (c *credentials) ha1() string {
	return h(fmt.Sprintf("%s:%s:%s", c.Username, c.Realm, c.password))
}

func (c *credentials) ha2() string {
	return h(fmt.Sprintf("%s:%s", c.method, c.DigestURI))
}

func (c *credentials) resp() (string, error) {
	c.NonceCount++

	if c.MessageQop != "auth" {
		return "", fmt.Errorf("unexpected messageqop '%s'",
			c.MessageQop)
	}

	b := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", fmt.Errorf("read full: %w", err)
	}

	c.Cnonce = fmt.Sprintf("%x", b)[:16]

	data := fmt.Sprintf("%s:%08x:%s:%s:%s",
		c.Nonce, c.NonceCount, c.Cnonce, c.MessageQop, c.ha2())
	return kd(c.ha1(), data), nil
}

func (c *credentials) authorize() (string, error) {
	// Note that this is only implemented for MD5 and NOT MD5-sess.
	// MD5-sess is rarely supported and those that do are a big mess.
	if c.Algorithm != md5Algorithm {
		return "", fmt.Errorf("unsupported algorithm '%s'",
			c.Algorithm)
	}

	resp, err := c.resp()
	if err != nil {
		return "", fmt.Errorf("resp: %w", err)
	}

	sl := []string{fmt.Sprintf(`username="%s"`, c.Username)}
	sl = append(sl, fmt.Sprintf(`realm="%s"`, c.Realm))
	sl = append(sl, fmt.Sprintf(`nonce="%s"`, c.Nonce))
	sl = append(sl, fmt.Sprintf(`uri="%s"`, c.DigestURI))
	sl = append(sl, fmt.Sprintf(`response="%s"`, resp))

	if c.Algorithm != "" {
		sl = append(sl, fmt.Sprintf(`algorithm="%s"`, c.Algorithm))
	}

	if c.Opaque != "" {
		sl = append(sl, fmt.Sprintf(`opaque="%s"`, c.Opaque))
	}

	if c.MessageQop != "" {
		sl = append(sl, fmt.Sprintf("qop=%s", c.MessageQop))
		sl = append(sl, fmt.Sprintf("nc=%08x", c.NonceCount))
		sl = append(sl, fmt.Sprintf(`cnonce="%s"`, c.Cnonce))
	}

	return fmt.Sprintf("Digest %s", strings.Join(sl, ", ")), nil
}

func h(data string) string {
	// `gosec` won't be happy ("weak crypto primitive"), but it's what the
	// server uses.
	//
	//nolint:gosec
	hf := md5.New()
	if _, err := io.WriteString(hf, data); err != nil {
		panic(fmt.Errorf("write string: %w", err))
	}
	return fmt.Sprintf("%x", hf.Sum(nil))
}

func kd(secret, data string) string {
	return h(fmt.Sprintf("%s:%s", secret, data))
}
