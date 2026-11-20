package stage

import (
	"strings"
	"testing"
)

// Ikemen GO stage.go:1340-1344 reads [Bound] screenleft and screenright, the
// distance kept between a fighter and the screen edge.
func TestParse_BoundScreenMargins(t *testing.T) {
	s, err := Parse(strings.NewReader("[Bound]\nscreenleft = 20\nscreenright = 25\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.StageBoundaries.ScreenLeft == nil || *s.StageBoundaries.ScreenLeft != 20 || s.StageBoundaries.ScreenRight == nil || *s.StageBoundaries.ScreenRight != 25 {
		t.Fatalf("margins = %v %v, want 20 and 25", s.StageBoundaries.ScreenLeft, s.StageBoundaries.ScreenRight)
	}
}

func TestParse_NoBoundSectionLeavesTheMarginsUnset(t *testing.T) {
	s, err := Parse(strings.NewReader("[PlayerInfo]\nleftbound = -100\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.StageBoundaries.ScreenLeft != nil || s.StageBoundaries.ScreenRight != nil {
		t.Fatal("no [Bound]: the margins are unset, the engine's default applies")
	}
}

func TestParse_InvalidScreenLeftNamesItsLine(t *testing.T) {
	_, err := Parse(strings.NewReader("[Bound]\nscreenleft = wide\n"))
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("err = %v, want one naming line 2", err)
	}
}

func TestSerialize_BoundRoundTrips(t *testing.T) {
	left, right := 20, 25
	s := Stage{}
	s.StageBoundaries.ScreenLeft, s.StageBoundaries.ScreenRight = &left, &right
	var b strings.Builder
	if err := Serialize(&b, s); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "[Bound]\nscreenleft = 20\nscreenright = 25\n") {
		t.Fatalf("written = %q, want a [Bound] section", b.String())
	}
	back, err := Parse(strings.NewReader(b.String()))
	if err != nil || *back.StageBoundaries.ScreenLeft != 20 || *back.StageBoundaries.ScreenRight != 25 {
		t.Fatalf("read back %v %v (%v)", back.StageBoundaries.ScreenLeft, back.StageBoundaries.ScreenRight, err)
	}
	b.Reset()
	if err := Serialize(&b, Stage{}); err != nil || strings.Contains(b.String(), "[Bound]") {
		t.Fatalf("unset margins: no [Bound] section, got %q", b.String())
	}
}
