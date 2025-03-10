package crypto

import "testing"

func TestGenerateOtp(t *testing.T) {
	otp, err := GenerateOtp(6)
	if err != nil {
		t.Errorf("error generate otp : %s", err)
	}
	t.Log(otp)
}

func TestGenerateTokenHash(t *testing.T) {
	tokenHash := GenerateTokenHash("email@example.com", "123456")
	t.Log(tokenHash)
}
