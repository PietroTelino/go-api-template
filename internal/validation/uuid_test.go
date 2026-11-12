package validation

import "testing"

func TestIsUUID(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{"uuid minúsculo", "3f2b8c1e-9a4d-4e2f-8b1a-2c3d4e5f6a7b", true},
		{"uuid maiúsculo", "3F2B8C1E-9A4D-4E2F-8B1A-2C3D4E5F6A7B", true},
		{"vazio", "", false},
		{"texto qualquer", "abc", false},
		{"sem hífens", "3f2b8c1e9a4d4e2f8b1a2c3d4e5f6a7b", false},
		{"caractere inválido", "3f2b8c1e-9a4d-4e2f-8b1a-2c3d4e5f6a7z", false},
		{"com espaço no fim", "3f2b8c1e-9a4d-4e2f-8b1a-2c3d4e5f6a7b ", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := IsUUID(testCase.value); got != testCase.want {
				t.Errorf("IsUUID(%q) = %v, esperado %v", testCase.value, got, testCase.want)
			}
		})
	}
}
