package stage

import "testing"

func TestModel_ZeroValue_AllFieldsZero(t *testing.T) {
	var m Model

	if m.OffsetX != 0 || m.OffsetY != 0 || m.OffsetZ != 0 {
		t.Errorf("expected zero-value Model to have Offset 0,0,0, got %v,%v,%v", m.OffsetX, m.OffsetY, m.OffsetZ)
	}
	if m.ScaleX != 0 || m.ScaleY != 0 || m.ScaleZ != 0 {
		t.Errorf("expected zero-value Model to have Scale 0,0,0, got %v,%v,%v", m.ScaleX, m.ScaleY, m.ScaleZ)
	}
	if m.Environment != "" {
		t.Errorf("expected zero-value Model to have empty Environment, got %q", m.Environment)
	}
	if m.EnvironmentIntensity != 0 {
		t.Errorf("expected zero-value Model to have EnvironmentIntensity 0, got %v", m.EnvironmentIntensity)
	}
}

func TestModel_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	m := Model{
		OffsetX: 0, OffsetY: -0.25, OffsetZ: -1,
		ScaleX: 0.5, ScaleY: 0.5, ScaleZ: 0.5,
		Environment:          "stage.hdr",
		EnvironmentIntensity: 1.2,
	}

	if m.OffsetX != 0 || m.OffsetY != -0.25 || m.OffsetZ != -1 {
		t.Errorf("expected Offset 0,-0.25,-1, got %v,%v,%v", m.OffsetX, m.OffsetY, m.OffsetZ)
	}
	if m.ScaleX != 0.5 || m.ScaleY != 0.5 || m.ScaleZ != 0.5 {
		t.Errorf("expected Scale 0.5,0.5,0.5, got %v,%v,%v", m.ScaleX, m.ScaleY, m.ScaleZ)
	}
	if m.Environment != "stage.hdr" {
		t.Errorf("expected Environment %q, got %q", "stage.hdr", m.Environment)
	}
	if m.EnvironmentIntensity != 1.2 {
		t.Errorf("expected EnvironmentIntensity 1.2, got %v", m.EnvironmentIntensity)
	}
}
