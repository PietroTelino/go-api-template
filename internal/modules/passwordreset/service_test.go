package passwordreset

import "testing"

func TestHashToken(t *testing.T) {
	firstHash := hashToken("TOKEN-A")

	if len(firstHash) != 64 {
		t.Errorf("o hash tem %d caracteres, esperado 64", len(firstHash))
	}

	if hashToken("TOKEN-A") != firstHash {
		t.Error("o mesmo token gerou hashes diferentes")
	}

	if hashToken("TOKEN-B") == firstHash {
		t.Error("tokens diferentes geraram o mesmo hash")
	}

	if firstHash == "TOKEN-A" {
		t.Error("o hash é igual ao token")
	}
}

func TestBuildResetLink(t *testing.T) {
	testCases := []struct {
		frontendURL string
		want        string
	}{
		{"http://localhost:5173", "http://localhost:5173/reset-password?token=ABC123"},
		{"http://localhost:5173/", "http://localhost:5173/reset-password?token=ABC123"},
		{"https://app.exemplo.com/painel", "https://app.exemplo.com/painel/reset-password?token=ABC123"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.frontendURL, func(t *testing.T) {
			service := &Service{frontendURL: testCase.frontendURL}

			if got := service.buildResetLink("ABC123"); got != testCase.want {
				t.Errorf("buildResetLink = %q, esperado %q", got, testCase.want)
			}
		})
	}
}
