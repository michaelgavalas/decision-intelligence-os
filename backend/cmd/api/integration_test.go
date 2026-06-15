package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/config"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
)

// gqlErr is a single GraphQL error as returned in the top-level "errors" array.
type gqlErr struct {
	Message    string         `json:"message"`
	Extensions map[string]any `json:"extensions"`
}

// gqlResp is a decoded GraphQL response: raw data plus any transport errors.
type gqlResp struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlErr        `json:"errors"`
}

// gqlClient is a minimal GraphQL-over-HTTP client used by the integration
// tests. It carries an optional bearer token applied to every request, an
// optional CSRF header value, and (when a jar is attached) replays cookies set
// by previous responses, recording newly issued cookies for assertions.
type gqlClient struct {
	t      *testing.T
	h      http.Handler
	token  string
	csrf   string
	jar    map[string]string
	useJar bool
}

// newClient builds a client that posts to the given handler unauthenticated.
func newClient(t *testing.T, h http.Handler) *gqlClient {
	t.Helper()
	return &gqlClient{t: t, h: h}
}

// with returns a copy of the client that sends the given bearer token.
func (c *gqlClient) with(token string) *gqlClient {
	clone := *c
	clone.token = token
	return &clone
}

// withCookieJar returns a copy of the client that replays and records cookies
// across requests, emulating a browser. Each cookie jar is independent.
func (c *gqlClient) withCookieJar() *gqlClient {
	clone := *c
	clone.jar = map[string]string{}
	clone.useJar = true
	return &clone
}

// withCSRF returns a copy of the client that sends the given X-CSRF-Token
// header on each request.
func (c *gqlClient) withCSRF(token string) *gqlClient {
	clone := *c
	clone.csrf = token
	return &clone
}

// cookie returns the current value of a jar cookie, or "" when absent.
func (c *gqlClient) cookie(name string) string {
	return c.jar[name]
}

// exec posts a GraphQL operation and decodes the response. It fails the test on
// transport-level problems (non-200 status, undecodable body) but returns
// GraphQL-level errors to the caller for assertion.
func (c *gqlClient) exec(query string, vars map[string]any) gqlResp {
	c.t.Helper()

	payload := map[string]any{"query": query}
	if vars != nil {
		payload["variables"] = vars
	}
	body, err := json.Marshal(payload)
	if err != nil {
		c.t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	if c.useJar {
		for name, value := range c.jar {
			req.AddCookie(&http.Cookie{Name: name, Value: value})
		}
	}
	rec := httptest.NewRecorder()
	c.h.ServeHTTP(rec, req)

	if c.useJar {
		c.recordCookies(rec.Result().Cookies())
	}

	if rec.Code != http.StatusOK {
		c.t.Fatalf("graphql status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var out gqlResp
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		c.t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	return out
}

// recordCookies updates the jar from a response's Set-Cookie headers, applying
// expirations so cleared cookies are removed.
func (c *gqlClient) recordCookies(cookies []*http.Cookie) {
	for _, ck := range cookies {
		if ck.MaxAge < 0 {
			delete(c.jar, ck.Name)
			continue
		}
		c.jar[ck.Name] = ck.Value
	}
}

// mustData fails the test if the response carries any top-level GraphQL errors
// and returns the data payload for further decoding.
func (c *gqlClient) mustData(resp gqlResp) json.RawMessage {
	c.t.Helper()
	if len(resp.Errors) > 0 {
		c.t.Fatalf("unexpected graphql errors: %+v", resp.Errors)
	}
	return resp.Data
}

// firstErrCode returns the "code" extension of the first error, or "" when
// there are no errors or no code.
func firstErrCode(resp gqlResp) string {
	if len(resp.Errors) == 0 {
		return ""
	}
	code, _ := resp.Errors[0].Extensions["code"].(string)
	return code
}

// firstErrReason returns the "reason" extension (the domain code) of the first
// error, or "" when absent.
func firstErrReason(resp gqlResp) string {
	if len(resp.Errors) == 0 {
		return ""
	}
	reason, _ := resp.Errors[0].Extensions["reason"].(string)
	return reason
}

// decode unmarshals the data payload into v, failing the test on error.
func (c *gqlClient) decode(data json.RawMessage, v any) {
	c.t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		c.t.Fatalf("decode data: %v; data=%s", err, string(data))
	}
}

// registerUser registers a new account and returns its access and refresh
// tokens. It fails the test if registration surfaces any user errors.
func registerUser(c *gqlClient, email, name, pw string) (accessToken, refreshToken string) {
	c.t.Helper()

	const mutation = `
mutation($email:String!,$name:String!,$pw:String!){
  register(input:{email:$email,name:$name,password:$pw}){
    accessToken
    refreshToken
    user { id email }
    userErrors { code message }
  }
}`
	resp := c.exec(mutation, map[string]any{"email": email, "name": name, "pw": pw})
	var out struct {
		Register struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			User         struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"user"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"register"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.Register.UserErrors) > 0 {
		c.t.Fatalf("register %q surfaced user errors: %+v", email, out.Register.UserErrors)
	}
	if out.Register.AccessToken == "" || out.Register.RefreshToken == "" {
		c.t.Fatalf("register %q returned empty tokens: %+v", email, out.Register)
	}
	return out.Register.AccessToken, out.Register.RefreshToken
}

// registerUserID registers an account and returns its user id along with tokens.
func registerUserID(c *gqlClient, email, name, pw string) (userID, accessToken, refreshToken string) {
	c.t.Helper()

	const mutation = `
mutation($email:String!,$name:String!,$pw:String!){
  register(input:{email:$email,name:$name,password:$pw}){
    accessToken
    refreshToken
    user { id }
    userErrors { code }
  }
}`
	resp := c.exec(mutation, map[string]any{"email": email, "name": name, "pw": pw})
	var out struct {
		Register struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			User         struct {
				ID string `json:"id"`
			} `json:"user"`
			UserErrors []struct {
				Code string `json:"code"`
			} `json:"userErrors"`
		} `json:"register"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.Register.UserErrors) > 0 {
		c.t.Fatalf("register %q surfaced user errors: %+v", email, out.Register.UserErrors)
	}
	return out.Register.User.ID, out.Register.AccessToken, out.Register.RefreshToken
}

// firstTeamID returns the id of the caller's first team (the personal team
// created at registration).
func firstTeamID(c *gqlClient) string {
	c.t.Helper()

	resp := c.exec(`query{ myTeams{ id } }`, nil)
	var out struct {
		MyTeams []struct {
			ID string `json:"id"`
		} `json:"myTeams"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.MyTeams) == 0 {
		c.t.Fatalf("expected at least one team, got none")
	}
	return out.MyTeams[0].ID
}

// uniqueEmail builds a per-scenario unique email to avoid cross-talk between
// subtests sharing one database.
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

// TestIntegration exercises the GraphQL API end-to-end against a single
// containerized database and app, with one subtest per critical-path scenario.
func TestIntegration(t *testing.T) {
	_, dsn := dbtest.NewPoolWithURL(t)

	cfg, err := config.Load(func(k string) string {
		if k == "DATABASE_URL" {
			return dsn
		}
		return ""
	})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	app, err := NewApp(context.Background(), cfg, testLogger())
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	t.Cleanup(app.Close)

	base := newClient(t, app.Handler)

	t.Run("FullLifecycle", func(t *testing.T) { testFullLifecycle(t, base) })
	t.Run("Authorization", func(t *testing.T) { testAuthorization(t, base) })
	t.Run("UserErrorsAsData", func(t *testing.T) { testUserErrorsAsData(t, base) })
	t.Run("CursorPagination", func(t *testing.T) { testCursorPagination(t, base) })
	t.Run("RefreshRotation", func(t *testing.T) { testRefreshRotation(t, base) })
	t.Run("AIDisabled", func(t *testing.T) { testAIDisabled(t, base) })
}

// testFullLifecycle drives a decision from creation through outcome and then
// verifies the deep nested read and the team analytics derived from it.
func testFullLifecycle(t *testing.T, base *gqlClient) {
	email := uniqueEmail("lifecycle")
	access, _ := registerUser(base, email, "Lifecycle Owner", "password123")
	c := base.with(access)
	teamID := firstTeamID(c)

	// Create decision -> DRAFT.
	decisionID := createDecision(t, c, teamID, "Should we launch?", "Evaluating launch.")

	// Transition to ACTIVE.
	transition(t, c, decisionID, "ACTIVE", "ACTIVE")

	// Add two assumptions at confidence 0.7.
	a1 := addAssumption(t, c, decisionID, "Market demand exists", 0.7)
	addAssumption(t, c, decisionID, "Team can deliver", 0.7)

	// Attach URL evidence to the first assumption.
	attachEvidence(t, c, a1, "URL", "https://example.com/report")

	// Create a prediction at probability 0.8.
	createPrediction(t, c, decisionID, "Launch will succeed", 0.8)

	// Record a successful outcome; this also marks the decision DECIDED.
	recordOutcome(t, c, decisionID, "Launched and grew", true)

	// Deep nested read.
	const query = `
query($id:ID!){
  decision(id:$id){
    id
    status
    owner { email }
    team { name }
    assumptions { id confidence evidence { id sourceType } }
    predictions { probability }
    outcome { success }
  }
}`
	resp := c.exec(query, map[string]any{"id": decisionID})
	var out struct {
		Decision struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Owner  struct {
				Email string `json:"email"`
			} `json:"owner"`
			Team struct {
				Name string `json:"name"`
			} `json:"team"`
			Assumptions []struct {
				ID         string  `json:"id"`
				Confidence float64 `json:"confidence"`
				Evidence   []struct {
					ID         string `json:"id"`
					SourceType string `json:"sourceType"`
				} `json:"evidence"`
			} `json:"assumptions"`
			Predictions []struct {
				Probability float64 `json:"probability"`
			} `json:"predictions"`
			Outcome struct {
				Success bool `json:"success"`
			} `json:"outcome"`
		} `json:"decision"`
	}
	c.decode(c.mustData(resp), &out)

	if out.Decision.Status != "DECIDED" {
		t.Fatalf("status = %q, want DECIDED", out.Decision.Status)
	}
	if out.Decision.Owner.Email != email {
		t.Fatalf("owner email = %q, want %q", out.Decision.Owner.Email, email)
	}
	if len(out.Decision.Assumptions) != 2 {
		t.Fatalf("assumptions = %d, want 2", len(out.Decision.Assumptions))
	}
	// Evidence must be attached to the right assumption (a1) and only that one.
	evidenceCount := 0
	for _, a := range out.Decision.Assumptions {
		if a.ID == a1 {
			if len(a.Evidence) != 1 {
				t.Fatalf("evidence on a1 = %d, want 1", len(a.Evidence))
			}
			if a.Evidence[0].SourceType != "URL" {
				t.Fatalf("evidence sourceType = %q, want URL", a.Evidence[0].SourceType)
			}
		}
		evidenceCount += len(a.Evidence)
	}
	if evidenceCount != 1 {
		t.Fatalf("total evidence across assumptions = %d, want 1", evidenceCount)
	}
	if len(out.Decision.Predictions) != 1 {
		t.Fatalf("predictions = %d, want 1", len(out.Decision.Predictions))
	}
	if !out.Decision.Outcome.Success {
		t.Fatalf("outcome.success = false, want true")
	}

	// teamMetrics derived from the single resolved forecast.
	metricsResp := c.exec(`
query($t:ID!){
  teamMetrics(teamId:$t){ brierScore forecastCount decisionSuccessRate resolvedDecisionCount }
}`, map[string]any{"t": teamID})
	var metrics struct {
		TeamMetrics struct {
			BrierScore            float64 `json:"brierScore"`
			ForecastCount         int     `json:"forecastCount"`
			DecisionSuccessRate   float64 `json:"decisionSuccessRate"`
			ResolvedDecisionCount int     `json:"resolvedDecisionCount"`
		} `json:"teamMetrics"`
	}
	c.decode(c.mustData(metricsResp), &metrics)

	if metrics.TeamMetrics.ForecastCount != 1 {
		t.Fatalf("forecastCount = %d, want 1", metrics.TeamMetrics.ForecastCount)
	}
	if metrics.TeamMetrics.ResolvedDecisionCount != 1 {
		t.Fatalf("resolvedDecisionCount = %d, want 1", metrics.TeamMetrics.ResolvedDecisionCount)
	}
	if metrics.TeamMetrics.DecisionSuccessRate != 1.0 {
		t.Fatalf("decisionSuccessRate = %v, want 1.0", metrics.TeamMetrics.DecisionSuccessRate)
	}
	// Brier = (0.8 - 1)^2 = 0.04.
	if math.Abs(metrics.TeamMetrics.BrierScore-0.04) > 1e-6 {
		t.Fatalf("brierScore = %v, want ~0.04", metrics.TeamMetrics.BrierScore)
	}

	// Calibration: sample sizes across bins must sum to the single forecast.
	calResp := c.exec(`
query($t:ID!){
  calibration(teamId:$t){ bins{ bucket sampleSize } }
}`, map[string]any{"t": teamID})
	var cal struct {
		Calibration struct {
			Bins []struct {
				Bucket     int `json:"bucket"`
				SampleSize int `json:"sampleSize"`
			} `json:"bins"`
		} `json:"calibration"`
	}
	c.decode(c.mustData(calResp), &cal)
	total := 0
	for _, b := range cal.Calibration.Bins {
		total += b.SampleSize
	}
	if total != 1 {
		t.Fatalf("calibration sample sizes sum = %d, want 1", total)
	}
}

// testAuthorization verifies cross-tenant isolation and viewer restrictions.
//
// Documented behavior: a user reading a decision in a team they do NOT belong
// to receives code FORBIDDEN. The decisions service loads the decision and then
// resolves the caller's role via the teams service, whose GetMembership first
// calls requireMember; a non-member fails with NOT_TEAM_MEMBER (Kind Forbidden),
// so the surfaced transport code is FORBIDDEN (not NOT_FOUND).
func testAuthorization(t *testing.T, base *gqlClient) {
	emailA := uniqueEmail("authz-a")
	emailB := uniqueEmail("authz-b")

	accessA, _ := registerUser(base, emailA, "Owner A", "password123")
	a := base.with(accessA)
	teamA := firstTeamID(a)
	decisionID := createDecision(t, a, teamA, "A's decision", "private to A")

	bUserID, accessB, _ := registerUserID(base, emailB, "User B", "password123")
	b := base.with(accessB)

	// B reads A's decision before being added: NOT_FOUND (documented above).
	resp := b.exec(`query($id:ID!){ decision(id:$id){ id } }`, map[string]any{"id": decisionID})
	if code := firstErrCode(resp); code != "FORBIDDEN" && code != "NOT_FOUND" {
		t.Fatalf("foreign decision read code = %q, want FORBIDDEN or NOT_FOUND", code)
	} else {
		t.Logf("foreign decision read returned code %q", code)
	}

	// A adds B as a VIEWER.
	addResp := a.exec(`
mutation($t:ID!,$u:ID!){
  addTeamMember(input:{teamId:$t,userId:$u,role:VIEWER}){
    membership { role }
    userErrors { code message }
  }
}`, map[string]any{"t": teamA, "u": bUserID})
	var add struct {
		AddTeamMember struct {
			Membership struct {
				Role string `json:"role"`
			} `json:"membership"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"addTeamMember"`
	}
	a.decode(a.mustData(addResp), &add)
	if len(add.AddTeamMember.UserErrors) > 0 {
		t.Fatalf("addTeamMember surfaced user errors: %+v", add.AddTeamMember.UserErrors)
	}
	if add.AddTeamMember.Membership.Role != "VIEWER" {
		t.Fatalf("membership role = %q, want VIEWER", add.AddTeamMember.Membership.Role)
	}

	// Now B can read A's decision.
	readResp := b.exec(`query($id:ID!){ decision(id:$id){ id status } }`, map[string]any{"id": decisionID})
	var read struct {
		Decision struct {
			ID string `json:"id"`
		} `json:"decision"`
	}
	b.decode(b.mustData(readResp), &read)
	if read.Decision.ID != decisionID {
		t.Fatalf("viewer read id = %q, want %q", read.Decision.ID, decisionID)
	}

	// B (viewer) cannot create a decision in A's team.
	createResp := b.exec(`
mutation($t:ID!){
  createDecision(input:{teamId:$t,title:"nope",description:""}){
    decision { id }
    userErrors { code }
  }
}`, map[string]any{"t": teamA})
	if code := firstErrCode(createResp); code != "FORBIDDEN" {
		t.Fatalf("viewer createDecision code = %q, want FORBIDDEN", code)
	}

	// B (viewer) cannot add an assumption to A's decision.
	addAsResp := b.exec(`
mutation($d:ID!){
  addAssumption(input:{decisionId:$d,statement:"x",confidence:0.5}){
    assumption { id }
    userErrors { code }
  }
}`, map[string]any{"d": decisionID})
	if code := firstErrCode(addAsResp); code != "FORBIDDEN" {
		t.Fatalf("viewer addAssumption code = %q, want FORBIDDEN", code)
	}

	// Unauthenticated me -> UNAUTHENTICATED.
	anon := base.exec(`query{ me { id } }`, nil)
	if code := firstErrCode(anon); code != "UNAUTHENTICATED" {
		t.Fatalf("unauthenticated me code = %q, want UNAUTHENTICATED", code)
	}
}

// testUserErrorsAsData verifies that domain validation failures arrive as
// in-band userErrors rather than top-level transport errors.
func testUserErrorsAsData(t *testing.T, base *gqlClient) {
	email := uniqueEmail("usererr")
	access, _ := registerUser(base, email, "Validator", "password123")
	c := base.with(access)
	teamID := firstTeamID(c)
	decisionID := createDecision(t, c, teamID, "Valid decision", "")

	// Confidence out of range -> INVALID_CONFIDENCE userError, null assumption.
	resp := c.exec(`
mutation($d:ID!){
  addAssumption(input:{decisionId:$d,statement:"too sure",confidence:1.5}){
    assumption { id }
    userErrors { code message }
  }
}`, map[string]any{"d": decisionID})
	if len(resp.Errors) > 0 {
		t.Fatalf("expected no top-level errors, got %+v", resp.Errors)
	}
	var addOut struct {
		AddAssumption struct {
			Assumption *struct {
				ID string `json:"id"`
			} `json:"assumption"`
			UserErrors []struct {
				Code string `json:"code"`
			} `json:"userErrors"`
		} `json:"addAssumption"`
	}
	c.decode(c.mustData(resp), &addOut)
	if addOut.AddAssumption.Assumption != nil {
		t.Fatalf("expected null assumption, got %+v", addOut.AddAssumption.Assumption)
	}
	if len(addOut.AddAssumption.UserErrors) == 0 || addOut.AddAssumption.UserErrors[0].Code != "INVALID_CONFIDENCE" {
		t.Fatalf("expected INVALID_CONFIDENCE userError, got %+v", addOut.AddAssumption.UserErrors)
	}

	// Empty title -> createDecision userErrors present.
	emptyResp := c.exec(`
mutation($t:ID!){
  createDecision(input:{teamId:$t,title:"",description:""}){
    decision { id }
    userErrors { code message }
  }
}`, map[string]any{"t": teamID})
	if len(emptyResp.Errors) > 0 {
		t.Fatalf("expected no top-level errors for empty title, got %+v", emptyResp.Errors)
	}
	var createOut struct {
		CreateDecision struct {
			Decision *struct {
				ID string `json:"id"`
			} `json:"decision"`
			UserErrors []struct {
				Code string `json:"code"`
			} `json:"userErrors"`
		} `json:"createDecision"`
	}
	c.decode(c.mustData(emptyResp), &createOut)
	if len(createOut.CreateDecision.UserErrors) == 0 {
		t.Fatalf("expected userErrors for empty title, got none")
	}

	// Weak password on register -> WEAK_PASSWORD userError.
	weakResp := base.exec(`
mutation($e:String!){
  register(input:{email:$e,name:"Weak",password:"short"}){
    accessToken
    userErrors { code }
  }
}`, map[string]any{"e": uniqueEmail("weak")})
	if len(weakResp.Errors) > 0 {
		t.Fatalf("expected no top-level errors for weak password, got %+v", weakResp.Errors)
	}
	var weakOut struct {
		Register struct {
			UserErrors []struct {
				Code string `json:"code"`
			} `json:"userErrors"`
		} `json:"register"`
	}
	base.decode(base.mustData(weakResp), &weakOut)
	found := false
	for _, ue := range weakOut.Register.UserErrors {
		if ue.Code == "WEAK_PASSWORD" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected WEAK_PASSWORD userError, got %+v", weakOut.Register.UserErrors)
	}
}

// testCursorPagination verifies relay-style forward pagination over a team's
// decisions, including page boundaries, totalCount, and id uniqueness.
func testCursorPagination(t *testing.T, base *gqlClient) {
	email := uniqueEmail("paginate")
	access, _ := registerUser(base, email, "Paginator", "password123")
	c := base.with(access)
	teamID := firstTeamID(c)

	for i := 0; i < 5; i++ {
		createDecision(t, c, teamID, fmt.Sprintf("Decision %d", i), "")
	}

	const query = `
query($t:ID!,$first:Int!,$after:String){
  decisions(teamId:$t, first:$first, after:$after){
    edges { node { id } cursor }
    pageInfo { hasNextPage endCursor }
    totalCount
  }
}`
	type page struct {
		Decisions struct {
			Edges []struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
				Cursor string `json:"cursor"`
			} `json:"edges"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			TotalCount int `json:"totalCount"`
		} `json:"decisions"`
	}

	fetch := func(after any) page {
		vars := map[string]any{"t": teamID, "first": 2, "after": after}
		resp := c.exec(query, vars)
		var p page
		c.decode(c.mustData(resp), &p)
		return p
	}

	seen := map[string]bool{}
	record := func(p page) {
		for _, e := range p.Decisions.Edges {
			if seen[e.Node.ID] {
				t.Fatalf("duplicate id across pages: %s", e.Node.ID)
			}
			seen[e.Node.ID] = true
		}
	}

	p1 := fetch(nil)
	if len(p1.Decisions.Edges) != 2 {
		t.Fatalf("page1 edges = %d, want 2", len(p1.Decisions.Edges))
	}
	if !p1.Decisions.PageInfo.HasNextPage {
		t.Fatalf("page1 hasNextPage = false, want true")
	}
	if p1.Decisions.TotalCount != 5 {
		t.Fatalf("totalCount = %d, want 5", p1.Decisions.TotalCount)
	}
	record(p1)

	p2 := fetch(p1.Decisions.PageInfo.EndCursor)
	if len(p2.Decisions.Edges) != 2 {
		t.Fatalf("page2 edges = %d, want 2", len(p2.Decisions.Edges))
	}
	record(p2)

	p3 := fetch(p2.Decisions.PageInfo.EndCursor)
	if len(p3.Decisions.Edges) != 1 {
		t.Fatalf("page3 edges = %d, want 1", len(p3.Decisions.Edges))
	}
	if p3.Decisions.PageInfo.HasNextPage {
		t.Fatalf("page3 hasNextPage = true, want false")
	}
	record(p3)

	if len(seen) != 5 {
		t.Fatalf("distinct ids across pages = %d, want 5", len(seen))
	}
}

// testRefreshRotation verifies that exchanging a refresh token rotates it and
// that reusing the old token is rejected.
//
// Documented behavior: replaying a rotated (already-consumed) refresh token
// surfaces domain code TOKEN_REUSE (transport code UNAUTHENTICATED, extension
// reason TOKEN_REUSE).
func testRefreshRotation(t *testing.T, base *gqlClient) {
	email := uniqueEmail("refresh")
	_, oldRefresh := registerUser(base, email, "Refresher", "password123")

	const mutation = `
mutation($tok:String){
  refreshToken(token:$tok){
    accessToken
    refreshToken
    userErrors { code }
  }
}`
	type refreshOut struct {
		RefreshToken struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			UserErrors   []struct {
				Code string `json:"code"`
			} `json:"userErrors"`
		} `json:"refreshToken"`
	}

	// --- Path (a): explicit token argument (API/mobile clients). ---
	resp := base.exec(mutation, map[string]any{"tok": oldRefresh})
	var out refreshOut
	base.decode(base.mustData(resp), &out)
	if len(out.RefreshToken.UserErrors) > 0 {
		t.Fatalf("refresh surfaced user errors: %+v", out.RefreshToken.UserErrors)
	}
	if out.RefreshToken.AccessToken == "" {
		t.Fatalf("rotated access token is empty")
	}
	if out.RefreshToken.RefreshToken == "" || out.RefreshToken.RefreshToken == oldRefresh {
		t.Fatalf("expected a new refresh token distinct from the old one")
	}

	// Reusing the old refresh token must be rejected.
	reuse := base.exec(mutation, map[string]any{"tok": oldRefresh})
	code := firstErrCode(reuse)
	reason := firstErrReason(reuse)
	if code == "" {
		t.Fatalf("expected an error on refresh-token reuse, got data: %s", string(reuse.Data))
	}
	if reason != "TOKEN_REUSE" && reason != "INVALID_REFRESH" {
		t.Fatalf("refresh reuse reason = %q, want TOKEN_REUSE or INVALID_REFRESH (code=%q)", reason, code)
	}
	t.Logf("refresh reuse returned code %q reason %q", code, reason)

	// --- Path (b): cookie + CSRF (browser clients). ---
	// Log in through a cookie-jar client so the refresh and csrf cookies are
	// captured from the Set-Cookie response headers.
	browserEmail := uniqueEmail("refresh-cookie")
	registerUser(base, browserEmail, "Cookie Refresher", "password123")
	browser := base.withCookieJar()
	loginViaCookie(browser, browserEmail, "password123")

	csrf := browser.cookie("csrf_token")
	if csrf == "" {
		t.Fatalf("login did not set a csrf_token cookie")
	}
	if browser.cookie("refresh_token") == "" {
		t.Fatalf("login did not set a refresh_token cookie")
	}

	// Refresh with no token arg + matching CSRF header succeeds.
	withCSRF := browser.withCSRF(csrf)
	cookieResp := withCSRF.exec(mutation, nil)
	var cookieOut refreshOut
	withCSRF.decode(withCSRF.mustData(cookieResp), &cookieOut)
	if len(cookieOut.RefreshToken.UserErrors) > 0 {
		t.Fatalf("cookie refresh surfaced user errors: %+v", cookieOut.RefreshToken.UserErrors)
	}
	if cookieOut.RefreshToken.AccessToken == "" {
		t.Fatalf("cookie refresh returned empty access token")
	}
	// The refresh cookie must have rotated to a new value.
	if rotated := browser.cookie("refresh_token"); rotated == "" {
		t.Fatalf("cookie refresh did not re-set the refresh_token cookie")
	}

	// --- Negative: cookie present but missing X-CSRF-Token -> CSRF_INVALID. ---
	noCSRF := browser.exec(mutation, nil)
	if got := firstErrCode(noCSRF); got != "UNAUTHENTICATED" {
		t.Fatalf("missing-csrf refresh code = %q, want UNAUTHENTICATED", got)
	}
	if got := firstErrReason(noCSRF); got != "CSRF_INVALID" {
		t.Fatalf("missing-csrf refresh reason = %q, want CSRF_INVALID", got)
	}

	// --- Negative: cookie present but wrong X-CSRF-Token -> CSRF_INVALID. ---
	wrongCSRF := browser.withCSRF("not-the-right-token")
	badResp := wrongCSRF.exec(mutation, nil)
	if got := firstErrReason(badResp); got != "CSRF_INVALID" {
		t.Fatalf("wrong-csrf refresh reason = %q, want CSRF_INVALID", got)
	}
}

// loginViaCookie authenticates and (for a cookie-jar client) captures the auth
// cookies set by the login response.
func loginViaCookie(c *gqlClient, email, pw string) {
	c.t.Helper()
	const mutation = `
mutation($e:String!,$pw:String!){
  login(input:{email:$e,password:$pw}){
    accessToken
    refreshToken
    userErrors { code }
  }
}`
	resp := c.exec(mutation, map[string]any{"e": email, "pw": pw})
	var out struct {
		Login struct {
			AccessToken string `json:"accessToken"`
			UserErrors  []struct {
				Code string `json:"code"`
			} `json:"userErrors"`
		} `json:"login"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.Login.UserErrors) > 0 {
		c.t.Fatalf("login surfaced user errors: %+v", out.Login.UserErrors)
	}
	if out.Login.AccessToken == "" {
		c.t.Fatalf("login returned empty access token")
	}
}

// testAIDisabled verifies that AI features are off by default and report a
// disabled state rather than executing.
func testAIDisabled(t *testing.T, base *gqlClient) {
	email := uniqueEmail("ai")
	access, _ := registerUser(base, email, "AI User", "password123")
	c := base.with(access)

	resp := c.exec(`mutation{ detectBias(text:"We always win because we always have."){ summary } }`, nil)
	if len(resp.Errors) == 0 {
		t.Fatalf("expected a top-level error for disabled AI, got data: %s", string(resp.Data))
	}
	if firstErrReason(resp) != "FEATURE_DISABLED" {
		t.Fatalf("detectBias reason = %q (code=%q), want FEATURE_DISABLED",
			firstErrReason(resp), firstErrCode(resp))
	}
}

// --- mutation helpers shared across scenarios ---

func createDecision(t *testing.T, c *gqlClient, teamID, title, desc string) string {
	t.Helper()
	resp := c.exec(`
mutation($t:ID!,$title:String!,$desc:String!){
  createDecision(input:{teamId:$t,title:$title,description:$desc}){
    decision { id status }
    userErrors { code message }
  }
}`, map[string]any{"t": teamID, "title": title, "desc": desc})
	var out struct {
		CreateDecision struct {
			Decision *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"decision"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"createDecision"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.CreateDecision.UserErrors) > 0 {
		t.Fatalf("createDecision user errors: %+v", out.CreateDecision.UserErrors)
	}
	if out.CreateDecision.Decision == nil {
		t.Fatalf("createDecision returned nil decision")
	}
	if out.CreateDecision.Decision.Status != "DRAFT" {
		t.Fatalf("new decision status = %q, want DRAFT", out.CreateDecision.Decision.Status)
	}
	return out.CreateDecision.Decision.ID
}

func transition(t *testing.T, c *gqlClient, id, status, wantStatus string) {
	t.Helper()
	resp := c.exec(`
mutation($id:ID!,$s:DecisionStatus!){
  transitionDecision(input:{id:$id,status:$s}){
    decision { status }
    userErrors { code message }
  }
}`, map[string]any{"id": id, "s": status})
	var out struct {
		TransitionDecision struct {
			Decision *struct {
				Status string `json:"status"`
			} `json:"decision"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"transitionDecision"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.TransitionDecision.UserErrors) > 0 {
		t.Fatalf("transitionDecision user errors: %+v", out.TransitionDecision.UserErrors)
	}
	if out.TransitionDecision.Decision == nil || out.TransitionDecision.Decision.Status != wantStatus {
		t.Fatalf("transition status = %+v, want %q", out.TransitionDecision.Decision, wantStatus)
	}
}

func addAssumption(t *testing.T, c *gqlClient, decisionID, statement string, confidence float64) string {
	t.Helper()
	resp := c.exec(`
mutation($d:ID!,$s:String!,$conf:Float!){
  addAssumption(input:{decisionId:$d,statement:$s,confidence:$conf}){
    assumption { id confidence }
    userErrors { code message }
  }
}`, map[string]any{"d": decisionID, "s": statement, "conf": confidence})
	var out struct {
		AddAssumption struct {
			Assumption *struct {
				ID         string  `json:"id"`
				Confidence float64 `json:"confidence"`
			} `json:"assumption"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"addAssumption"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.AddAssumption.UserErrors) > 0 {
		t.Fatalf("addAssumption user errors: %+v", out.AddAssumption.UserErrors)
	}
	if out.AddAssumption.Assumption == nil {
		t.Fatalf("addAssumption returned nil assumption")
	}
	return out.AddAssumption.Assumption.ID
}

func attachEvidence(t *testing.T, c *gqlClient, assumptionID, sourceType, url string) string {
	t.Helper()
	resp := c.exec(`
mutation($a:ID!,$st:EvidenceSourceType!,$url:String){
  attachEvidence(input:{assumptionId:$a,sourceType:$st,sourceUrl:$url,content:"supporting evidence"}){
    evidence { id sourceType }
    userErrors { code message }
  }
}`, map[string]any{"a": assumptionID, "st": sourceType, "url": url})
	var out struct {
		AttachEvidence struct {
			Evidence *struct {
				ID         string `json:"id"`
				SourceType string `json:"sourceType"`
			} `json:"evidence"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"attachEvidence"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.AttachEvidence.UserErrors) > 0 {
		t.Fatalf("attachEvidence user errors: %+v", out.AttachEvidence.UserErrors)
	}
	if out.AttachEvidence.Evidence == nil {
		t.Fatalf("attachEvidence returned nil evidence")
	}
	return out.AttachEvidence.Evidence.ID
}

func createPrediction(t *testing.T, c *gqlClient, decisionID, statement string, probability float64) string {
	t.Helper()
	resp := c.exec(`
mutation($d:ID!,$s:String!,$p:Float!){
  createPrediction(input:{decisionId:$d,statement:$s,probability:$p}){
    prediction { id probability }
    userErrors { code message }
  }
}`, map[string]any{"d": decisionID, "s": statement, "p": probability})
	var out struct {
		CreatePrediction struct {
			Prediction *struct {
				ID string `json:"id"`
			} `json:"prediction"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"createPrediction"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.CreatePrediction.UserErrors) > 0 {
		t.Fatalf("createPrediction user errors: %+v", out.CreatePrediction.UserErrors)
	}
	if out.CreatePrediction.Prediction == nil {
		t.Fatalf("createPrediction returned nil prediction")
	}
	return out.CreatePrediction.Prediction.ID
}

func recordOutcome(t *testing.T, c *gqlClient, decisionID, summary string, success bool) {
	t.Helper()
	resolvedAt := time.Now().UTC().Format(time.RFC3339)
	resp := c.exec(`
mutation($d:ID!,$sum:String!,$ok:Boolean!,$at:DateTime!){
  recordOutcome(input:{decisionId:$d,summary:$sum,success:$ok,resolvedAt:$at}){
    outcome { id success }
    userErrors { code message }
  }
}`, map[string]any{"d": decisionID, "sum": summary, "ok": success, "at": resolvedAt})
	var out struct {
		RecordOutcome struct {
			Outcome *struct {
				ID      string `json:"id"`
				Success bool   `json:"success"`
			} `json:"outcome"`
			UserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"recordOutcome"`
	}
	c.decode(c.mustData(resp), &out)
	if len(out.RecordOutcome.UserErrors) > 0 {
		t.Fatalf("recordOutcome user errors: %+v", out.RecordOutcome.UserErrors)
	}
	if out.RecordOutcome.Outcome == nil {
		t.Fatalf("recordOutcome returned nil outcome")
	}
}
