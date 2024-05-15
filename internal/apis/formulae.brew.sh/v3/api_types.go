package v3

import (
	"encoding/json"
	"fmt"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	v2 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v2"
)

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

// CaskJWS represents the cached API data
type CaskJWS struct {
	Payload    FormulaJWS  `json:"payload"`
	Signatures []Signature `json:"signatures"`
}

// FormulaJWS represents the cached API data
type FormulaJWS struct {
	Payload    FormulaPayload `json:"payload"`
	Signatures []Signature    `json:"signatures"`
}

// Signature represents the signature field
type Signature struct {
	Protected string            `json:"protected"`
	Header    map[string]string `json:"header"`
	Signature string            `json:"signature"`
}

// FormulaPayload represents the embedded JSON payload field
type FormulaPayload v1.Index

// UnmarshalJSON implements json.Unmarshaler
func (payload *FormulaPayload) UnmarshalJSON(b []byte) error {
	var data string
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("parsing payload string: %w", err)
	}

	var index v1.Index
	err = json.Unmarshal([]byte(data), &index)
	if err != nil {
		return fmt.Errorf("parsing payload data: %w", err)
	}

	*payload = FormulaPayload(index)

	return nil
}

// MarshalJSON implements json.Marshaler
func (payload *FormulaPayload) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(v1.Index(*payload))
	if err != nil {
		return nil, err
	}
	// Marshal these bytes as a string (embedding the data as escaped JSON)
	return json.Marshal(string(b))
}

// CaskPayload represents the embedded JSON payload field
type CaskPayload []*v2.Cask

// UnmarshalJSON implements json.Unmarshaler
func (payload *CaskPayload) UnmarshalJSON(b []byte) error {
	var data string
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("parsing payload string: %w", err)
	}

	var index []*v2.Cask
	err = json.Unmarshal([]byte(data), &index)
	if err != nil {
		return fmt.Errorf("parsing payload: %w", err)
	}

	*payload = CaskPayload(index)

	return nil
}

// MarshalJSON implements json.Marshaler
func (payload *CaskPayload) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal([]*v2.Cask(*payload))
	if err != nil {
		return nil, err
	}
	// Marshal these bytes as a string (embedding the data as escaped JSON)
	return json.Marshal(string(b))
}
