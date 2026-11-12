package users

import "testing"

func TestGenerateRandomPassword(t *testing.T) {
	generatedPasswords := map[string]bool{}

	for range 200 {
		password, err := generateRandomPassword()
		if err != nil {
			t.Fatalf("erro ao gerar senha: %v", err)
		}

		if len(password) != generatedPasswordLength {
			t.Errorf("senha %q tem %d caracteres, esperado %d", password, len(password), generatedPasswordLength)
		}

		if err := ValidatePassword(password); err != nil {
			t.Errorf("senha gerada %q não passa na política: %v", password, err)
		}

		generatedPasswords[password] = true
	}

	if len(generatedPasswords) != 200 {
		t.Errorf("esperado 200 senhas diferentes, gerou %d", len(generatedPasswords))
	}
}
