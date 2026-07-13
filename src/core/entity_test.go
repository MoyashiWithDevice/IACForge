package core

import (
	"testing"
)

func TestNewEntity(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	if e.ID != "srv-01" {
		t.Errorf("expected ID srv-01, got %s", e.ID)
	}
	if e.Kind != "server" {
		t.Errorf("expected Kind server, got %s", e.Kind)
	}
	if e.Name != "Server 01" {
		t.Errorf("expected Name Server 01, got %s", e.Name)
	}
	if e.Owner != "" {
		t.Errorf("expected empty Owner, got %s", e.Owner)
	}
	if e.Properties == nil {
		t.Error("expected Properties to be initialized")
	}
}

func TestEntitySetOwner(t *testing.T) {
	e := NewEntity("rack-01", "rack", "Rack 01")
	if !e.IsRoot() {
		t.Error("expected entity to be root initially")
	}
	e.SetOwner("site-01")
	if e.IsRoot() {
		t.Error("expected entity not to be root after setting owner")
	}
	if e.Owner != "site-01" {
		t.Errorf("expected owner site-01, got %s", e.Owner)
	}
}

func TestEntitySetStatus(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetStatus(StatusActive)
	if e.Status != StatusActive {
		t.Errorf("expected status active, got %s", e.Status)
	}
}

func TestEntityTags(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.AddTag("production")
	e.AddTag("ap-northeast-1")

	if len(e.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e.Tags))
	}
	if !e.HasTag("production") {
		t.Error("expected to have tag production")
	}
	if !e.HasTag("ap-northeast-1") {
		t.Error("expected to have tag ap-northeast-1")
	}
	if e.HasTag("nonexistent") {
		t.Error("expected not to have tag nonexistent")
	}
}

func TestEntityLabels(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetLabel("region", "asia-pacific")
	e.SetLabel("tier", "primary")

	if v, ok := e.GetLabel("region"); !ok || v != "asia-pacific" {
		t.Errorf("expected label region=asia-pacific, got %s", v)
	}
	if v, ok := e.GetLabel("tier"); !ok || v != "primary" {
		t.Errorf("expected label tier=primary, got %s", v)
	}
	if _, ok := e.GetLabel("nonexistent"); ok {
		t.Error("expected no label nonexistent")
	}
}

func TestEntityProperties(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetProperty("cpu_cores", 32)
	e.SetProperty("memory_gb", float64(128))

	if v, ok := e.GetProperty("cpu_cores"); !ok || v != 32 {
		t.Errorf("expected property cpu_cores=32, got %v", v)
	}
	if v, ok := e.GetProperty("memory_gb"); !ok || v != float64(128) {
		t.Errorf("expected property memory_gb=128, got %v", v)
	}
	if _, ok := e.GetProperty("nonexistent"); ok {
		t.Error("expected no property nonexistent")
	}
}

func TestEntityValidate(t *testing.T) {
	tests := []struct {
		name    string
		entity  *Entity
		wantErr error
	}{
		{
			name:    "valid entity",
			entity:  NewEntity("srv-01", "server", "Server 01"),
			wantErr: nil,
		},
		{
			name:    "missing ID",
			entity:  NewEntity("", "server", "Server 01"),
			wantErr: ErrEntityMissingID,
		},
		{
			name:    "missing kind",
			entity:  NewEntity("srv-01", "", "Server 01"),
			wantErr: ErrEntityMissingKind,
		},
		{
			name:    "missing name",
			entity:  NewEntity("srv-01", "server", ""),
			wantErr: ErrEntityMissingName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEntityPath(t *testing.T) {
	e := NewEntity("eno1", "interface", "eno1")
	if e.Path() != "" {
		t.Errorf("expected empty path, got %s", e.Path())
	}
	e.SetPath("/site01/rack01/pve01/eno1")
	if e.Path() != "/site01/rack01/pve01/eno1" {
		t.Errorf("expected path /site01/rack01/pve01/eno1, got %s", e.Path())
	}
}

func TestEntityFullPath(t *testing.T) {
	e := NewEntity("eno1", "interface", "eno1")
	if e.FullPath() != "/eno1" {
		t.Errorf("expected full path /eno1, got %s", e.FullPath())
	}
	e.SetPath("/site01/rack01/pve01/eno1")
	if e.FullPath() != "/site01/rack01/pve01/eno1" {
		t.Errorf("expected full path /site01/rack01/pve01/eno1, got %s", e.FullPath())
	}
}

func TestEntityString(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetStatus(StatusActive)
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}

func TestStatusConstants(t *testing.T) {
	statuses := []Status{
		StatusPlanned, StatusActive, StatusMaintenance,
		StatusDeprecated, StatusOffline,
	}
	expected := []string{"planned", "active", "maintenance", "deprecated", "offline"}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("expected status %s, got %s", expected[i], s)
		}
	}
}
