package stage

import "testing"

func TestCameraBounds_ZeroValue_AllFieldsZero(t *testing.T) {
	var c CameraBounds

	if c.Left != 0 || c.Right != 0 || c.High != 0 || c.Low != 0 {
		t.Errorf("expected zero-value CameraBounds to have all edges at 0, got %+v", c)
	}
}

func TestCameraBounds_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	// boundleft/boundhigh are conventionally negative, boundright/boundlow
	// conventionally 0 or positive — exercise that asymmetry rather than a
	// symmetric fixture, since a sign-flip bug would otherwise go unnoticed.
	c := CameraBounds{Left: -180, Right: 180, High: -240, Low: 0}

	if c.Left != -180 {
		t.Errorf("expected Left -180, got %d", c.Left)
	}
	if c.Right != 180 {
		t.Errorf("expected Right 180, got %d", c.Right)
	}
	if c.High != -240 {
		t.Errorf("expected High -240, got %d", c.High)
	}
	if c.Low != 0 {
		t.Errorf("expected Low 0, got %d", c.Low)
	}
}

func TestStageBoundaries_ZeroValue_AllFieldsZero(t *testing.T) {
	var s StageBoundaries

	if s.Left != 0 || s.Right != 0 {
		t.Errorf("expected zero-value StageBoundaries to have Left=0, Right=0, got %+v", s)
	}
}

func TestStageBoundaries_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	s := StageBoundaries{Left: -1000, Right: 1000}

	if s.Left != -1000 {
		t.Errorf("expected Left -1000, got %d", s.Left)
	}
	if s.Right != 1000 {
		t.Errorf("expected Right 1000, got %d", s.Right)
	}
}
