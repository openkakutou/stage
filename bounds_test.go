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

func TestStageBoundaries_ZAxisFields_ZeroValueAndAssignable(t *testing.T) {
	var s StageBoundaries
	if s.TopBound != 0 || s.BottomBound != 0 {
		t.Errorf("expected zero-value TopBound/BottomBound 0/0, got %v/%v", s.TopBound, s.BottomBound)
	}

	s = StageBoundaries{Left: -1000, Right: 1000, TopBound: -50, BottomBound: 50}
	if s.TopBound != -50 {
		t.Errorf("expected TopBound -50, got %v", s.TopBound)
	}
	if s.BottomBound != 50 {
		t.Errorf("expected BottomBound 50, got %v", s.BottomBound)
	}
	// The X-axis and Z-axis clamps must stay independently settable — a bug
	// that aliased TopBound/BottomBound onto Left/Right would go unnoticed
	// without this cross-check.
	if s.TopBound == float64(s.Left) {
		t.Errorf("expected TopBound (%v) independent from Left (%d)", s.TopBound, s.Left)
	}
}

func TestPlayerStartZ_ZeroValue_AllFieldsZero(t *testing.T) {
	var p PlayerStartZ
	if p.P1 != 0 || p.P2 != 0 || p.P3 != 0 || p.P4 != 0 || p.P5 != 0 || p.P6 != 0 || p.P7 != 0 || p.P8 != 0 {
		t.Errorf("expected zero-value PlayerStartZ to have all players at 0, got %+v", p)
	}
}

func TestPlayerStartZ_WithValues_PreservesPerPlayerValues(t *testing.T) {
	p := PlayerStartZ{P1: -10, P2: 10, P3: 0, P4: 0, P5: 0, P6: 0, P7: 0, P8: 5}

	if p.P1 != -10 {
		t.Errorf("expected P1 -10, got %d", p.P1)
	}
	if p.P2 != 10 {
		t.Errorf("expected P2 10, got %d", p.P2)
	}
	if p.P8 != 5 {
		t.Errorf("expected P8 5, got %d", p.P8)
	}
}
