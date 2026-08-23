package stage

import "testing"

func TestScaling_ZeroValue_AllFieldsZero(t *testing.T) {
	var sc Scaling

	if sc.DepthToScreen != 0 || sc.TopZ != 0 || sc.BottomZ != 0 || sc.TopScale != 0 || sc.BottomScale != 0 {
		t.Errorf("expected zero-value Scaling to have all fields at 0, got %+v", sc)
	}
}

func TestScaling_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	// topz/topscale and botz/botscale are deliberately asymmetric (0/1 vs.
	// 50/1.2) so a field-swap bug between the top and bottom pair would be
	// caught rather than passing by coincidence.
	sc := Scaling{
		DepthToScreen: 0.5,
		TopZ:          0,
		BottomZ:       50,
		TopScale:      1,
		BottomScale:   1.2,
	}

	if sc.DepthToScreen != 0.5 {
		t.Errorf("expected DepthToScreen 0.5, got %v", sc.DepthToScreen)
	}
	if sc.TopZ != 0 {
		t.Errorf("expected TopZ 0, got %v", sc.TopZ)
	}
	if sc.BottomZ != 50 {
		t.Errorf("expected BottomZ 50, got %v", sc.BottomZ)
	}
	if sc.TopScale != 1 {
		t.Errorf("expected TopScale 1, got %v", sc.TopScale)
	}
	if sc.BottomScale != 1.2 {
		t.Errorf("expected BottomScale 1.2, got %v", sc.BottomScale)
	}
}
