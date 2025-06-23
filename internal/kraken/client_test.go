package kraken

import (
	"testing"
)

func TestRequestGeneration(t *testing.T) {
	fakePublicKey := "1212121"
	fakePrivateKey := "kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg=="
	desiredApiSign := "BVT3EumzzXSJHlvrinSwICz5uKKlSZPXL9cJIuKqn7ZMhHSbbtXhGdvwoDBmRz6ALXI+GVxFD0ZlGuOmfi2bxA=="
	payload := map[string]interface{}{
		"nonce":     "1616492376594",
		"ordertype": "limit",
		"pair":      "XBTUSD",
		"price":     37500,
		"type":      "buy",
		"volume":    1.25,
	}
	req, err := generateRequest(&requestParams{
		method:      "POST",
		path:        "/0/private/AddOrder",
		body:        payload,
		publicKey:   fakePublicKey,
		privateKey:  fakePrivateKey,
		environment: BaseURL,
	})
	if err != nil {
		t.Errorf("failed generating request: %v", err)
	}
	if req.Header.Values("API-Sign")[0] != desiredApiSign {
		t.Errorf("incorrect sign header: %v", req.Header)
	}
}
