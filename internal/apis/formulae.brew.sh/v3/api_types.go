package v3

// This uses a JSON Web Signature:
//
// Standard: https://openid.net/specs/draft-jones-json-web-signature-04.html
//
// Homebrew's signing script:
// https://github.com/Homebrew/formulae.brew.sh/blob/master/script/sign-json.rb
//
// OPA Go impl (internal):
// https://github.com/open-policy-agent/opa/blob/main/internal/jwx/jws/jws.go
//
// https://pkg.go.dev/github.com/lestrrat-go/jwx/v2/jws#section-readme
//
// TODO: use the signature to verify Homebrew's API responses

// Response represents the v3 API's response format
type Response struct {
	// Payload is the requested tap's information
	Payload Tap `json:"payload"`

	// Signatures is a list of JSON Web Signatures that can be used to verify the payload
	Signatures []Signature `json:"signatures"`
}

// Signature represents the signature field
type Signature struct {
	Protected string            `json:"protected"`
	Header    map[string]string `json:"header"`
	Signature string            `json:"signature"`
}
