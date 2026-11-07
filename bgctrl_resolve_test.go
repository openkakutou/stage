package stage

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func parseTargetingFixture(t *testing.T) Stage {
	t.Helper()
	s, err := Parse(strings.NewReader(string(readTestdataFile(t, "mugen-bgctrl-targeting.def"))))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return s
}

// layers indexes: 0 Sky(id1), 1 Neon(id2), 2 Free(id3), 3 Plain(no id).
func TestParse_BGElementID_ReadsIdKeyCaseInsensitively(t *testing.T) {
	s := parseTargetingFixture(t)
	want := []int{1, 2, 3, 0}
	for i, w := range want {
		if s.Elements[i].ID != w {
			t.Errorf("Elements[%d].ID = %d, want %d", i, s.Elements[i].ID, w)
		}
	}
}

func TestSerialize_BGElementID_WrittenOnlyWhenNonzeroAndReparsed(t *testing.T) {
	var b strings.Builder
	s := Stage{Elements: []BGElement{{Name: "a", Type: BGElementNormal, ID: 4}, {Name: "b", Type: BGElementNormal}}}
	if err := Serialize(&b, s); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(strings.ToLower(b.String()), "\nid ="); got != 1 {
		t.Errorf("id lines = %d, want 1 (zero id must not be written) in:\n%s", got, b.String())
	}
	back, err := Parse(strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(back.Elements) != 2 || back.Elements[0].ID != 4 || back.Elements[1].ID != 0 {
		t.Errorf("re-parsed IDs = %+v, want [4 0]", back.Elements)
	}
}

func TestParse_BGElementID_RejectsNonInteger(t *testing.T) {
	_, err := Parse(strings.NewReader("[BG x]\ntype = normal\nid = abc\n"))
	if err == nil || !strings.Contains(err.Error(), "invalid id") {
		t.Errorf("err = %v, want an 'invalid id' error", err)
	}
}

func TestSerializeDef_BGElementID_ByteExactWhenUnchanged(t *testing.T) {
	data := readTestdataFile(t, "mugen-bgctrl-targeting.def")
	s := parseTargetingFixture(t)
	out, err := SerializeDef(data, s)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(data) {
		t.Errorf("SerializeDef changed an unedited document")
	}
}

func TestSerializeDef_RealStageWithBGIDs_ByteExactWhenUnchanged(t *testing.T) {
	data := readTestdataFile(t, "mugen-real-bgctrl-visible-stage.def")
	s, err := Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	out, err := SerializeDef(data, s)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(data) {
		t.Errorf("SerializeDef changed an unedited real stage")
	}
}

func TestResolveBGCtrlState_LoopingVisibleToggle(t *testing.T) {
	s := parseTargetingFixture(t)
	// Sky is id 1; Blink hides it for t in [30,120) of a 120-tick loop.
	cases := []struct {
		tick    int
		visible bool
	}{{0, true}, {29, true}, {30, false}, {119, false}, {120, true}, {130, true}, {150, false}, {239, false}}
	for _, c := range cases {
		got := ResolveBGCtrlState(s, c.tick)[0].Visible
		if got != c.visible {
			t.Errorf("tick %d: Sky.Visible = %v, want %v", c.tick, got, c.visible)
		}
	}
}

func TestResolveBGCtrlState_ReturnsOneStatePerElementWithName(t *testing.T) {
	got := ResolveBGCtrlState(parseTargetingFixture(t), 0)
	names := []string{"Sky", "Neon", "Free", "Plain"}
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}
	for i, n := range names {
		if got[i].Name != n {
			t.Errorf("[%d].Name = %q, want %q", i, got[i].Name, n)
		}
		if !got[i].Visible || !got[i].Enabled || got[i].AnimAction != -1 || got[i].OffsetX != 0 || got[i].OffsetY != 0 {
			t.Errorf("[%d] at tick 0 = %+v, want defaults", i, got[i])
		}
	}
}

func TestResolveBGCtrlState_ZeroIDControllerTargetsAllElements(t *testing.T) {
	s := parseTargetingFixture(t)
	// Again (enabled=0 at t>=5 of loop 240) has no id anywhere: hits all four.
	for _, tc := range []struct {
		tick    int
		enabled bool
	}{{4, true}, {5, false}, {244, true}, {245, false}} {
		for i, st := range ResolveBGCtrlState(s, tc.tick) {
			if st.Enabled != tc.enabled {
				t.Errorf("tick %d layer %d: Enabled = %v, want %v", tc.tick, i, st.Enabled, tc.enabled)
			}
		}
	}
}

func TestResolveBGCtrlState_IDTargetingLeavesOtherElementsUntouched(t *testing.T) {
	s := parseTargetingFixture(t)
	got := ResolveBGCtrlState(s, 50)
	// Blink (id 1) hid only Sky; Switch/Move (id 2) touch only Neon.
	if got[0].Visible {
		t.Errorf("Sky should be hidden at tick 50")
	}
	for _, i := range []int{1, 2, 3} {
		if !got[i].Visible {
			t.Errorf("layer %d should stay visible at tick 50", i)
		}
	}
	if got[1].AnimAction != 11 {
		t.Errorf("Neon.AnimAction = %d, want 11", got[1].AnimAction)
	}
	for _, i := range []int{0, 2, 3} {
		if got[i].AnimAction != -1 || got[i].OffsetX != 0 || got[i].OffsetY != 0 {
			t.Errorf("layer %d must have no anim/offset, got %+v", i, got[i])
		}
	}
}

func TestResolveBGCtrlState_PositionVelocityAndGroupIDInheritance(t *testing.T) {
	s := parseTargetingFixture(t)
	// Neon (id 2). Hand-derived: posset (5,-3) from t>=10; velset (2,1) from
	// 20, veladd (+1,0) from 40, so velocity (3,1) from 40. Nudge (+10,0) in
	// group Tail (ctrlID inherited, no loop) from tick>=15.
	cases := []struct {
		tick int
		x, y float64
		anim int
	}{
		{9, 0, 0, -1},
		{10, 5, -3, -1},
		{14, 5, -3, -1},
		{15, 15, -3, -1},
		{30, 5 + 20 + 10, -3 + 10, -1},          // 10 ticks at (2,1)
		{49, 5 + 40 + 27 + 10, -3 + 20 + 9, -1}, // 20*(2,1) + 9*(3,1)
		{50, 5 + 40 + 30 + 10, -3 + 20 + 10, 11},
		{240, 10, 0, -1},                          // loop 240 restarts; Tail has no loop and keeps Nudge
		{300, 5 + 40 + 60 + 10, -3 + 20 + 20, 11}, // t=60: 20*(2,1) + 20*(3,1)
	}
	for _, c := range cases {
		got := ResolveBGCtrlState(s, c.tick)[1]
		if got.OffsetX != c.x || got.OffsetY != c.y || got.AnimAction != c.anim {
			t.Errorf("tick %d: Neon = (%v,%v) anim %d, want (%v,%v) anim %d", c.tick, got.OffsetX, got.OffsetY, got.AnimAction, c.x, c.y, c.anim)
		}
	}
}

func TestResolveBGCtrlState_TimeListActiveOnlyInsideWindow(t *testing.T) {
	s := parseTargetingFixture(t)
	for _, c := range []struct {
		tick    int
		visible bool
	}{{99, true}, {100, false}, {109, false}, {110, true}, {340, false}, {350, true}} {
		if got := ResolveBGCtrlState(s, c.tick)[2].Visible; got != c.visible {
			t.Errorf("tick %d: Free.Visible = %v, want %v", c.tick, got, c.visible)
		}
	}
}

func TestResolveBGCtrlState_TimeListLoopRepeatsWindow(t *testing.T) {
	s := Stage{
		Elements: []BGElement{{Name: "e"}},
		BGCtrlDefs: []BGCtrlDef{{Controllers: []BGCtrl{{
			Type: "visible", Params: []BGCtrlParam{{Key: "time", Value: "2,4,10"}, {Key: "value", Value: "0"}},
		}}}},
	}
	for _, c := range []struct {
		tick    int
		visible bool
	}{{1, true}, {2, false}, {3, false}, {4, true}, {12, false}, {14, true}, {23, false}} {
		if got := ResolveBGCtrlState(s, c.tick)[0].Visible; got != c.visible {
			t.Errorf("tick %d: Visible = %v, want %v", c.tick, got, c.visible)
		}
	}
}

func TestResolveBGCtrlState_UnknownTypeAndUnparsableValueAreIgnored(t *testing.T) {
	s := parseTargetingFixture(t)
	// Bad (value=abc) and Unknown (type wobble, value=0) both target all at
	// t>=1; nothing else touches Plain at tick 4.
	got := ResolveBGCtrlState(s, 4)[3]
	want := BGLayerState{Name: "Plain", Visible: true, Enabled: true, AnimAction: -1}
	if got != want {
		t.Errorf("Plain = %+v, want the default %+v", got, want)
	}
	if len(s.BGCtrlDefs[1].Controllers) < 6 {
		t.Fatalf("fixture lost its ignored controllers")
	}
}

func TestResolveBGCtrlState_NegativeTickIsTreatedAsZero(t *testing.T) {
	s := parseTargetingFixture(t)
	a, b := ResolveBGCtrlState(s, -50), ResolveBGCtrlState(s, 0)
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("layer %d: tick -50 = %+v, tick 0 = %+v", i, a[i], b[i])
		}
	}
}

func TestResolveBGCtrlState_HugeTickWrapsLikeItsRemainder(t *testing.T) {
	s := parseTargetingFixture(t)
	got := ResolveBGCtrlState(s, 240*1000000+50)[1]
	if got.OffsetX != 85 || got.OffsetY != 27 || got.AnimAction != 11 {
		t.Errorf("Neon = %+v, want offset (85,27) anim 11", got)
	}
}

func TestResolveBGCtrlState_GroupWithoutControllersChangesNothing(t *testing.T) {
	s := Stage{
		Elements:   []BGElement{{Name: "e", ID: 7}},
		BGCtrlDefs: []BGCtrlDef{{Name: "Empty", CtrlID: 7, LoopTime: 5}},
	}
	got := ResolveBGCtrlState(s, 3)
	want := BGLayerState{Name: "e", Visible: true, Enabled: true, AnimAction: -1}
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %+v, want [%+v]", got, want)
	}
}

func TestResolveBGCtrlState_LaterControllerOverridesEarlierInFileOrder(t *testing.T) {
	s := Stage{
		Elements: []BGElement{{Name: "e"}},
		BGCtrlDefs: []BGCtrlDef{{Controllers: []BGCtrl{
			{Type: "visible", Time: 0, Params: []BGCtrlParam{{Key: "value", Value: "0"}}},
			{Type: "visible", Time: 5, Params: []BGCtrlParam{{Key: "value", Value: "1"}}},
		}}},
	}
	if ResolveBGCtrlState(s, 4)[0].Visible {
		t.Errorf("tick 4: first controller should hide")
	}
	if !ResolveBGCtrlState(s, 5)[0].Visible {
		t.Errorf("tick 5: second controller should win")
	}
}

func TestResolveBGCtrlState_EmptyStageReturnsEmptySlice(t *testing.T) {
	got := ResolveBGCtrlState(Stage{}, 10)
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
}

func TestResolveBGCtrlState_ElementsWithoutControllersAreDefault(t *testing.T) {
	got := ResolveBGCtrlState(Stage{Elements: []BGElement{{Name: "x"}}}, 99)
	if len(got) != 1 || !got[0].Visible || !got[0].Enabled || got[0].AnimAction != -1 {
		t.Errorf("got %+v", got)
	}
}

func TestResolveBGCtrlState_EnableSpellingIsAcceptedLikeEnabled(t *testing.T) {
	for _, typ := range []string{"enabled", "enable"} {
		s := Stage{
			Elements:   []BGElement{{Name: "e"}},
			BGCtrlDefs: []BGCtrlDef{{Controllers: []BGCtrl{{Type: typ, Time: 3, Params: []BGCtrlParam{{Key: "value", Value: "0"}}}}}},
		}
		if !ResolveBGCtrlState(s, 2)[0].Enabled || ResolveBGCtrlState(s, 3)[0].Enabled {
			t.Errorf("type %q: want enabled at tick 2, disabled at tick 3", typ)
		}
	}
}

func TestResolveBGCtrlState_NonPositiveLoopTimeMeansNoLoop(t *testing.T) {
	for _, lt := range []int{0, -4} {
		s := Stage{
			Elements:   []BGElement{{Name: "e"}},
			BGCtrlDefs: []BGCtrlDef{{LoopTime: lt, Controllers: []BGCtrl{{Type: "visible", Time: 10, Params: []BGCtrlParam{{Key: "value", Value: "0"}}}}}},
		}
		if ResolveBGCtrlState(s, 1000)[0].Visible {
			t.Errorf("looptime %d: controller must persist at tick 1000", lt)
		}
	}
}

func TestResolveBGCtrlState_TimeListWithoutEndPersistsFromStart(t *testing.T) {
	s := Stage{
		Elements:   []BGElement{{Name: "e"}},
		BGCtrlDefs: []BGCtrlDef{{Controllers: []BGCtrl{{Type: "visible", Params: []BGCtrlParam{{Key: "time", Value: "7"}, {Key: "value", Value: "0"}}}}}},
	}
	if !ResolveBGCtrlState(s, 6)[0].Visible || ResolveBGCtrlState(s, 5000)[0].Visible {
		t.Errorf("time list with only a start must hide from tick 7 on")
	}
}

// ctrlStage builds a one-element stage with the given controllers in a
// single group without loop.
func ctrlStage(cs ...BGCtrl) Stage {
	return Stage{Elements: []BGElement{{Name: "e"}}, BGCtrlDefs: []BGCtrlDef{{Controllers: cs}}}
}

func params(kv ...string) []BGCtrlParam {
	var out []BGCtrlParam
	for i := 0; i+1 < len(kv); i += 2 {
		out = append(out, BGCtrlParam{Key: kv[i], Value: kv[i+1]})
	}
	return out
}

func TestResolveBGCtrlState_ChronologicalOrderBeatsFileOrder(t *testing.T) {
	// The controller that fired last wins, wherever it sits in the file.
	vis := ctrlStage(
		BGCtrl{Type: "visible", Time: 50, Params: params("value", "0")},
		BGCtrl{Type: "visible", Time: 10, Params: params("value", "1")},
	)
	if !ResolveBGCtrlState(vis, 30)[0].Visible {
		t.Errorf("tick 30: only the time=10 controller fired, want visible")
	}
	if ResolveBGCtrlState(vis, 60)[0].Visible {
		t.Errorf("tick 60: time=50 fired after time=10, want hidden")
	}
	anim := ctrlStage(
		BGCtrl{Type: "anim", Time: 50, Params: params("value", "3")},
		BGCtrl{Type: "anim", Time: 10, Params: params("value", "7")},
	)
	if got := ResolveBGCtrlState(anim, 30)[0].AnimAction; got != 7 {
		t.Errorf("tick 30: AnimAction = %d, want 7", got)
	}
	if got := ResolveBGCtrlState(anim, 60)[0].AnimAction; got != 3 {
		t.Errorf("tick 60: AnimAction = %d, want 3", got)
	}
	pos := ctrlStage(
		BGCtrl{Type: "posset", Time: 50, Params: params("x", "9")},
		BGCtrl{Type: "posset", Time: 10, Params: params("x", "4")},
		BGCtrl{Type: "posadd", Time: 20, Params: params("x", "1")},
	)
	// 10: x=4; 20: +1 = 5; 50: posset 9 replaces everything.
	if got := ResolveBGCtrlState(pos, 30)[0].OffsetX; got != 5 {
		t.Errorf("tick 30: OffsetX = %v, want 5", got)
	}
	if got := ResolveBGCtrlState(pos, 60)[0].OffsetX; got != 9 {
		t.Errorf("tick 60: OffsetX = %v, want 9", got)
	}
}

func TestResolveBGCtrlState_SameTimeControllersFallBackToFileOrder(t *testing.T) {
	s := ctrlStage(
		BGCtrl{Type: "visible", Time: 5, Params: params("value", "0")},
		BGCtrl{Type: "visible", Time: 5, Params: params("value", "1")},
	)
	if !ResolveBGCtrlState(s, 9)[0].Visible {
		t.Errorf("tie on time: the later one in the file must win")
	}
}

func TestResolveBGCtrlState_LoopedWindowCountsAsFiringAtItsCurrentPeriod(t *testing.T) {
	s := ctrlStage(
		BGCtrl{Type: "visible", Time: 3, Params: params("value", "0")},
		BGCtrl{Type: "visible", Params: params("time", "2,,10", "value", "1")},
	)
	// Tick 5: starts 3 (single) vs 2 (looped, period 0) -> single wins.
	if ResolveBGCtrlState(s, 5)[0].Visible {
		t.Errorf("tick 5: want hidden (time=3 fired after the looped start at 2)")
	}
	// Tick 13: looped window restarted at 12 > 3 -> looped wins.
	if !ResolveBGCtrlState(s, 13)[0].Visible {
		t.Errorf("tick 13: want visible (looped window re-fired at 12)")
	}
}

func TestResolveBGCtrlState_PosSetResetsAccumulatedVelocity(t *testing.T) {
	for _, c := range []struct {
		posX float64
		tick int
		want float64
	}{{0, 49, 49}, {0, 50, 0}, {0, 60, 10}, {7, 60, 17}} {
		s := ctrlStage(
			BGCtrl{Type: "velset", Time: 0, Params: params("x", "1")},
			BGCtrl{Type: "posset", Time: 50, Params: params("x", strconv.FormatFloat(c.posX, 'f', -1, 64))},
		)
		if got := ResolveBGCtrlState(s, c.tick)[0].OffsetX; got != c.want {
			t.Errorf("posset x=%v tick %d: OffsetX = %v, want %v", c.posX, c.tick, got, c.want)
		}
	}
}

func TestResolveBGCtrlState_VelocityTimeListLoopRepeatsWindow(t *testing.T) {
	s := ctrlStage(BGCtrl{Type: "velset", Params: params("time", "2,4,10", "x", "3")})
	// Velocity 3 on ticks u with u mod 10 in [2,4): 2,3,12,13,22,23.
	for _, c := range []struct {
		tick int
		want float64
	}{{2, 0}, {4, 6}, {12, 6}, {13, 9}, {25, 18}} {
		if got := ResolveBGCtrlState(s, c.tick)[0].OffsetX; got != c.want {
			t.Errorf("tick %d: OffsetX = %v, want %v", c.tick, got, c.want)
		}
	}
}

func TestResolveBGCtrlState_VelocityTimeListEndIsExclusive(t *testing.T) {
	s := ctrlStage(BGCtrl{Type: "velset", Params: params("time", "0,5", "x", "2")})
	for _, c := range []struct {
		tick int
		want float64
	}{{4, 8}, {5, 10}, {6, 10}, {100, 10}} {
		if got := ResolveBGCtrlState(s, c.tick)[0].OffsetX; got != c.want {
			t.Errorf("tick %d: OffsetX = %v, want %v", c.tick, got, c.want)
		}
	}
}

func TestResolveBGCtrlState_UnparsableTimeListLeavesDefaults(t *testing.T) {
	want := BGLayerState{Name: "e", Visible: true, Enabled: true, AnimAction: -1}
	for _, list := range []string{"abc", "5,xyz", "5,10,q"} {
		for _, c := range []BGCtrl{
			{Type: "visible", Params: params("time", list, "value", "0")},
			{Type: "posset", Params: params("time", list, "x", "3", "y", "4")},
			{Type: "velset", Params: params("time", list, "x", "1")},
		} {
			if got := ResolveBGCtrlState(ctrlStage(c), 100)[0]; got != want {
				t.Errorf("%s time=%q: got %+v, want the default", c.Type, list, got)
			}
		}
	}
}

func TestResolveBGCtrlState_ControllerCtrlIDWinsOverNonzeroGroupCtrlID(t *testing.T) {
	s := Stage{
		Elements: []BGElement{{Name: "one", ID: 1}, {Name: "two", ID: 2}},
		BGCtrlDefs: []BGCtrlDef{{CtrlID: 1, Controllers: []BGCtrl{
			{Type: "visible", CtrlID: 2, Params: params("value", "0")},
		}}},
	}
	got := ResolveBGCtrlState(s, 1)
	if !got[0].Visible || got[1].Visible {
		t.Errorf("visible = [%v %v], want [true false] (controller ctrlID 2 overrides group ctrlID 1)", got[0].Visible, got[1].Visible)
	}
}

// parseRealVisibleStage parses the unmodified real community stage. Reading
// its text by hand: 50 BG elements carry "ID = 50" .. "ID = 1" (one each);
// 50 groups "[BGCtrlDef Back]" each have "ctrlid = N" and "looptime = 201"
// (N = 1..50), with two Visible controllers: "time = 4(N-1) value = 1" then
// "time = 4N value = 0" (group 50: 196 then 200).
func parseRealVisibleStage(t *testing.T) (Stage, map[int]int) {
	t.Helper()
	s, err := Parse(strings.NewReader(string(readTestdataFile(t, "mugen-real-bgctrl-visible-stage.def"))))
	if err != nil {
		t.Fatalf("parse real fixture: %v", err)
	}
	byID := map[int]int{}
	for i, el := range s.Elements {
		if el.ID != 0 {
			byID[el.ID] = i
		}
	}
	return s, byID
}

func TestParse_RealVisibleStage_ReadsIDsGroupsAndControllers(t *testing.T) {
	s, byID := parseRealVisibleStage(t)
	if len(byID) != 50 {
		t.Errorf("distinct element ids = %d, want 50", len(byID))
	}
	if len(s.BGCtrlDefs) != 50 {
		t.Fatalf("BGCtrlDefs = %d, want 50", len(s.BGCtrlDefs))
	}
	g := s.BGCtrlDefs[49]
	if g.Name != "Back" || g.CtrlID != 50 || g.LoopTime != 201 || len(g.Controllers) != 2 {
		t.Errorf("last group = %+v, want Back ctrlid 50 looptime 201 with 2 controllers", g)
	}
	if c := g.Controllers[1]; c.Type != "visible" || c.Time != 200 || c.CtrlID != 50 {
		t.Errorf("last controller = %+v, want visible time 200 ctrlid 50", c)
	}
}

func TestResolveBGCtrlState_RealVisibleStage_HidesEachIDInTurnAndWrapsAtLoopTime(t *testing.T) {
	s, byID := parseRealVisibleStage(t)
	// Hand-derived from the file: element id N is visible before time 4N of
	// the 201-tick loop and hidden from 4N on, until the wrap.
	cases := []struct {
		tick    int
		id      int
		visible bool
	}{
		{0, 1, true}, {3, 1, true}, {4, 1, false}, {5, 2, true}, {7, 2, true}, {8, 2, false},
		{100, 25, false}, {99, 25, true}, {100, 26, true}, {195, 49, true}, {196, 49, false}, {199, 50, true},
		{200, 50, false}, {200, 1, false},
		// 201 wraps to group tick 0: everything visible again.
		{201, 1, true}, {201, 50, true}, {204, 1, true}, {205, 1, false}, {206, 2, true},
		// 603 = 3*201: group tick 0; 803 = 3*201+200: group tick 200.
		{603, 50, true}, {803, 50, false}, {803, 49, false},
	}
	for _, c := range cases {
		got := ResolveBGCtrlState(s, c.tick)[byID[c.id]]
		if got.Visible != c.visible {
			t.Errorf("tick %d id %d: Visible = %v, want %v", c.tick, c.id, got.Visible, c.visible)
		}
	}
}

func TestResolveBGCtrlState_RealVisibleStage_HiddenCountAndUntargetedElements(t *testing.T) {
	s, _ := parseRealVisibleStage(t)
	// At group tick 100 ids 1..25 are hidden (4N <= 100); ids 26..50 and
	// every element without an id (not targeted by any group) stay visible.
	hidden := 0
	for _, st := range ResolveBGCtrlState(s, 100+201) {
		if !st.Visible {
			hidden++
		}
	}
	if hidden != 25 {
		t.Errorf("hidden at tick 301 = %d, want 25", hidden)
	}
	for i, st := range ResolveBGCtrlState(s, 150) {
		if s.Elements[i].ID == 0 && (!st.Visible || !st.Enabled || st.AnimAction != -1 || st.OffsetX != 0 || st.OffsetY != 0) {
			t.Fatalf("element %d (%s) without id changed: %+v", i, st.Name, st)
		}
	}
}

func TestResolveBGCtrlState_OverflowingOffsetStaysFinite(t *testing.T) {
	s := Stage{
		Elements: []BGElement{{Name: "a"}},
		BGCtrlDefs: []BGCtrlDef{{Controllers: []BGCtrl{
			{Type: "posadd", Time: 0, Params: []BGCtrlParam{{Key: "x", Value: "1e308"}}},
			{Type: "posadd", Time: 0, Params: []BGCtrlParam{{Key: "x", Value: "1e308"}}},
		}}},
	}
	got := ResolveBGCtrlState(s, 5)[0].OffsetX
	if math.IsInf(got, 0) || math.IsNaN(got) {
		t.Errorf("OffsetX = %v, want a finite value", got)
	}
	if got != math.MaxFloat64 {
		t.Errorf("OffsetX = %v, want it clamped to MaxFloat64", got)
	}
}
