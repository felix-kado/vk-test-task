package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr bool
	}{
		{"valid login", "user123", false},
		{"valid login with underscore", "user_name", false},
		{"too short", "ab", true},
		{"too long", "thisisaverylongloginthatiswaylongerthanfiftycharacterssothisshouldfail", true},
		{"starts with number", "1user", true},
		{"contains special chars", "user@name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogin(tt.login)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "Password1!", false},
		{"missing uppercase", "password1!", true},
		{"missing lowercase", "PASSWORD1!", true},
		{"missing number", "Password!", true},
		{"missing special char", "Password1", true},
		{"too short", "Pwd1!", true},
		{"too long", "ThisIsAVeryLongPasswordThatIsWayLongerThanSeventyTwoCharactersAndShouldFail123!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAdRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *AdRequest
		wantErr bool
	}{
		{
			"valid request",
			&AdRequest{
				Title: "Test Ad",
				Text:  "This is a test ad",
				Price: 1000,
			},
			false,
		},
		{
			"missing title",
			&AdRequest{
				Title: "",
				Text:  "This is a test ad",
				Price: 1000,
			},
			true,
		},
		{
			"missing text",
			&AdRequest{
				Title: "Test Ad",
				Text:  "",
				Price: 1000,
			},
			true,
		},
		{
			"negative price",
			&AdRequest{
				Title: "Test Ad",
				Text:  "This is a test ad",
				Price: -100,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAdRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateListParams(t *testing.T) {
	tests := []struct {
		name    string
		sortBy  string
		order   string
		wantErr bool
	}{
		{
			name:    "valid params",
			sortBy:  "price",
			order:   "asc",
			wantErr: false,
		},
		{
			name:    "valid params with created_at",
			sortBy:  "created_at",
			order:   "desc",
			wantErr: false,
		},
		{
			name:    "invalid sort field",
			sortBy:  "invalid",
			order:   "asc",
			wantErr: true,
		},
		{
			name:    "invalid order",
			sortBy:  "price",
			order:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListParams(tt.sortBy, tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
