package main

import "testing"

func TestInitDb(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"Empty URI": {
			input:   "",
			wantErr: true,
		},
		"Wrong URI": {
			input:   "wronguri",
			wantErr: true,
		},
		"Given URI": {
			input:   "mongodb://root:password@localhost:27017",
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := InitDb(tc.input); tc.wantErr != (err != nil) {
				t.Errorf("expected err: %v, got err: %v", tc.wantErr, err)
			}
		})
	}
}
