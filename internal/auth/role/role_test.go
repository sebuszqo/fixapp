package role

import (
	"encoding/json"
	"testing"
)

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{Unknown, "unknown"},
		{User, "user"},
		{Handyman, "handyman"},
		{Admin, "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  Role
	}{
		{"user", User},
		{"USER", User},
		{"User", User},
		{" user ", User},
		{"handyman", Handyman},
		{"HANDYMAN", Handyman},
		{"admin", Admin},
		{"ADMIN", Admin},
		{"invalid", Unknown},
		{"", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Parse(tt.input); got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	// Test valid roles
	if got := MustParse("user"); got != User {
		t.Errorf("MustParse(user) = %v, want %v", got, User)
	}

	// Test panic on invalid role
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse(invalid) did not panic")
		}
	}()
	MustParse("invalid")
}

func TestRoleIsValid(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{Unknown, false},
		{User, true},
		{Handyman, true},
		{Admin, true},
		{Role(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.role.String(), func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("Role.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleIsAtLeast(t *testing.T) {
	tests := []struct {
		role   Role
		target Role
		want   bool
	}{
		{Admin, Admin, true},
		{Admin, Handyman, true},
		{Admin, User, true},
		{Handyman, Handyman, true},
		{Handyman, User, true},
		{Handyman, Admin, false},
		{User, User, true},
		{User, Handyman, false},
		{User, Admin, false},
	}

	for _, tt := range tests {
		t.Run(tt.role.String()+"_"+tt.target.String(), func(t *testing.T) {
			if got := tt.role.IsAtLeast(tt.target); got != tt.want {
				t.Errorf("Role(%v).IsAtLeast(%v) = %v, want %v", tt.role, tt.target, got, tt.want)
			}
		})
	}
}

func TestRoleCanActAs(t *testing.T) {
	// Admin can act as any role
	if !Admin.CanActAs(Admin) || !Admin.CanActAs(Handyman) || !Admin.CanActAs(User) {
		t.Error("Admin should be able to act as any role")
	}

	// Handyman cannot act as Admin
	if Handyman.CanActAs(Admin) {
		t.Error("Handyman should not be able to act as Admin")
	}

	// User cannot act as Admin or Handyman
	if User.CanActAs(Admin) || User.CanActAs(Handyman) {
		t.Error("User should not be able to act as Admin or Handyman")
	}
}

func TestAll(t *testing.T) {
	roles := All()
	if len(roles) != 3 {
		t.Errorf("All() returned %d roles, want 3", len(roles))
	}

	expected := []Role{User, Handyman, Admin}
	for i, r := range expected {
		if roles[i] != r {
			t.Errorf("All()[%d] = %v, want %v", i, roles[i], r)
		}
	}
}

func TestRoleJSONMarshal(t *testing.T) {
	type testStruct struct {
		Role Role `json:"role"`
	}

	ts := testStruct{Role: Admin}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `{"role":"admin"}`
	if string(data) != expected {
		t.Errorf("json.Marshal = %s, want %s", string(data), expected)
	}
}

func TestRoleJSONUnmarshal(t *testing.T) {
	type testStruct struct {
		Role Role `json:"role"`
	}

	var ts testStruct
	err := json.Unmarshal([]byte(`{"role":"handyman"}`), &ts)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if ts.Role != Handyman {
		t.Errorf("json.Unmarshal role = %v, want %v", ts.Role, Handyman)
	}
}

func TestRoleJSONUnmarshalInvalid(t *testing.T) {
	type testStruct struct {
		Role Role `json:"role"`
	}

	var ts testStruct
	err := json.Unmarshal([]byte(`{"role":"invalid"}`), &ts)
	if err == nil {
		t.Error("json.Unmarshal should fail for invalid role")
	}
}

