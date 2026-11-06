package stage

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestParse_BGCtrlDef_ParsesGroupAndOrderedControllers(t *testing.T) {
	src := `[BGCtrlDef intro]
looptime = 120
ctrlID = 7

[BGCtrl a]
type = Visible
time = 10
ctrlID = 3
value = 0
foo = "bar"

[BGCtrl b]
type = PalFX
time = 20
add = 10, 20, 30
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []BGCtrlDef{{
		Name: "intro", CtrlID: 7, LoopTime: 120,
		Controllers: []BGCtrl{
			{Type: "visible", CtrlID: 3, Time: 10, Params: []BGCtrlParam{{Key: "value", Value: "0"}, {Key: "foo", Value: "bar"}}},
			{Type: "palfx", Time: 20, Params: []BGCtrlParam{{Key: "add", Value: "10, 20, 30"}}},
		},
	}}
	if !reflect.DeepEqual(s.BGCtrlDefs, want) {
		t.Errorf("BGCtrlDefs = %+v, want %+v", s.BGCtrlDefs, want)
	}
}

func TestParse_BGCtrl_MixedCaseHeadersAndKeys_ReusedNamesBelongByPosition(t *testing.T) {
	src := `[BGctrldef first]
LoopTime = 5
[bgctrl 1]
TYPE = visible
Time = 1
Visible = 0
[BGCTRLDEF second]
[BGCtrl 1]
type = VISIBLE
CTRLID = 2
visible = 1
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.BGCtrlDefs) != 2 {
		t.Fatalf("got %d defs, want 2: %+v", len(s.BGCtrlDefs), s.BGCtrlDefs)
	}
	a, b := s.BGCtrlDefs[0], s.BGCtrlDefs[1]
	if a.Name != "first" || a.LoopTime != 5 || len(a.Controllers) != 1 || a.Controllers[0].Type != "visible" ||
		a.Controllers[0].Params[0] != (BGCtrlParam{Key: "visible", Value: "0"}) {
		t.Errorf("first def wrong: %+v", a)
	}
	if b.Name != "second" || len(b.Controllers) != 1 || b.Controllers[0].CtrlID != 2 || b.Controllers[0].Type != "visible" {
		t.Errorf("second def wrong: %+v", b)
	}
	if len(s.Elements) != 0 {
		t.Errorf("BGCtrl headers must not become BG elements: %+v", s.Elements)
	}
}

func TestParse_BGCtrl_UnknownTypeKeptRaw(t *testing.T) {
	src := "[BGCtrlDef x]\n[BGCtrl 1]\ntype = Weird\ntime = 3\nspam = a,b\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := s.BGCtrlDefs[0].Controllers[0]
	if c.Type != "weird" || c.Time != 3 || !reflect.DeepEqual(c.Params, []BGCtrlParam{{Key: "spam", Value: "a,b"}}) {
		t.Errorf("unexpected controller %+v", c)
	}
}

func TestParse_NoBGCtrl_LeavesBGCtrlDefsNil(t *testing.T) {
	s, err := Parse(strings.NewReader("[Info]\nname = x\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.BGCtrlDefs != nil {
		t.Errorf("expected nil, got %+v", s.BGCtrlDefs)
	}
}

func TestParse_BGCtrlBeforeAnyBGCtrlDef_IsIgnored(t *testing.T) {
	src := "[BGCtrl 1]\ntype = Visible\ntime = notanumber\n[BG floor]\ntype = normal\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.BGCtrlDefs != nil {
		t.Errorf("expected nil, got %+v", s.BGCtrlDefs)
	}
	if len(s.Elements) != 1 {
		t.Errorf("BG element lost: %+v", s.Elements)
	}
}

func TestParse_BGCtrl_InvalidIntegers_ReturnLineNumberedError(t *testing.T) {
	cases := map[string]struct{ src, want string }{
		"time":     {"[BGCtrlDef x]\n[BGCtrl 1]\ntype = Visible\ntime = abc\n", `stage: line 4: invalid time "abc"`},
		"looptime": {"[BGCtrlDef x]\nlooptime = zz\n", `stage: line 2: invalid looptime "zz"`},
		"ctrlid":   {"[BGCtrlDef x]\n[BGCtrl 1]\nctrlID = 1,2\n", `stage: line 3: invalid ctrlid "1,2"`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(c.src))
			if err == nil || !strings.HasPrefix(err.Error(), c.want) {
				t.Errorf("err = %v, want prefix %q", err, c.want)
			}
		})
	}
}

func TestSerializeDef_UnmodifiedStageWithBGCtrl_ByteExactAfterJSONRoundTrip(t *testing.T) {
	withCtrl := "[Info]\nname = x\n\n[BGCtrlDef a]\nlooptime = 4\n[BGCtrl 1]\ntype = Visible\ntime = 1\nvisible = 0\n"
	without := "[Info]\nname = x\n"
	for name, src := range map[string]string{"with": withCtrl, "without": without} {
		t.Run(name, func(t *testing.T) {
			doc, err := ParseDocument(strings.NewReader(src))
			if err != nil {
				t.Fatal(err)
			}
			back := emptyNonNil(doc.Stage)
			out, err := SerializeDef([]byte(src), back)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, []byte(src)) {
				t.Errorf("not byte-exact:\n%q\nwant\n%q", out, src)
			}
		})
	}
}

func TestSerialize_IgnoresBGCtrlDefs(t *testing.T) {
	plain := Stage{Name: "x"}
	withCtrl := plain
	withCtrl.BGCtrlDefs = []BGCtrlDef{{Name: "a", Controllers: []BGCtrl{{Type: "visible"}}}}
	var a, b bytes.Buffer
	if err := Serialize(&a, plain); err != nil {
		t.Fatal(err)
	}
	if err := Serialize(&b, withCtrl); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Errorf("Serialize output changed by BGCtrlDefs:\n%q\n%q", a.String(), b.String())
	}
}

func TestParse_BGCtrl_TimeList_KeptRawInParams_StageStillParses(t *testing.T) {
	src := "[BGCtrlDef x]\n[BGCtrl 1]\ntype = Visible\ntime = 0, 30, 60\nvisible = 1\n[BG floor]\ntype = normal\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := BGCtrl{Type: "visible", Time: 0, Params: []BGCtrlParam{{Key: "time", Value: "0, 30, 60"}, {Key: "visible", Value: "1"}}}
	if len(s.BGCtrlDefs) != 1 || len(s.BGCtrlDefs[0].Controllers) != 1 || !reflect.DeepEqual(s.BGCtrlDefs[0].Controllers[0], want) {
		t.Errorf("got %+v, want controller %+v", s.BGCtrlDefs, want)
	}
	if len(s.Elements) != 1 {
		t.Errorf("BG element lost: %+v", s.Elements)
	}
}

func TestParse_BGCtrl_StrictIntegers_RejectDecimal(t *testing.T) {
	for name, src := range map[string]string{
		"time":     "[BGCtrlDef x]\n[BGCtrl 1]\ntime = 10.5\n",
		"looptime": "[BGCtrlDef x]\nlooptime = 10.5\n",
		"ctrlid":   "[BGCtrlDef x]\nctrlID = 10.5\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(src)); err == nil {
				t.Error("expected error for 10.5")
			}
		})
	}
}

func TestParse_BGCtrlDef_InvalidGroupCtrlID_ReturnsLineNumberedError(t *testing.T) {
	_, err := Parse(strings.NewReader("[BGCtrlDef x]\nctrlID = 1,2\n"))
	if err == nil || !strings.HasPrefix(err.Error(), `stage: line 2: invalid ctrlid "1,2"`) {
		t.Errorf("err = %v", err)
	}
}

func TestParse_BGCtrlDef_NamelessAndTabSeparatedHeaders(t *testing.T) {
	s, err := Parse(strings.NewReader("[BGCtrlDef]\nlooptime = 3\n[BGCtrlDef\tTabbed]\nlooptime = 5\n[BGCtrl\t7]\ntype = visible\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.BGCtrlDefs) != 2 || s.BGCtrlDefs[0].Name != "" || s.BGCtrlDefs[0].LoopTime != 3 ||
		s.BGCtrlDefs[1].Name != "Tabbed" || s.BGCtrlDefs[1].LoopTime != 5 || len(s.BGCtrlDefs[1].Controllers) != 1 {
		t.Errorf("got %+v", s.BGCtrlDefs)
	}
}

func TestParse_BGBlockAfterBGCtrl_DoesNotLeakKeysIntoParams(t *testing.T) {
	s, err := Parse(strings.NewReader("[BGCtrlDef x]\n[BGCtrl 1]\ntype = visible\n[BG floor]\ntype = normal\nvisible = 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := s.BGCtrlDefs[0].Controllers[0].Params; len(got) != 0 {
		t.Errorf("params leaked: %+v", got)
	}
}

func TestSerializeDef_HandWrittenEmptyJSONLists_ByteExact(t *testing.T) {
	withCtrl := "[Info]\nname = x\n\n[BGCtrlDef a]\nlooptime = 4\n[BGCtrl 1]\ntype = Visible\ntime = 1\n\n[BGCtrlDef b]\nlooptime = 2\n"
	without := "[Info]\nname = x\n"
	for name, src := range map[string]string{"with": withCtrl, "without": without} {
		t.Run(name, func(t *testing.T) {
			doc, err := ParseDocument(strings.NewReader(src))
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(doc.Stage)
			if err != nil {
				t.Fatal(err)
			}
			var generic map[string]any
			if err := json.Unmarshal(raw, &generic); err != nil {
				t.Fatal(err)
			}
			generic["elements"] = []any{}
			if defs, ok := generic["bgCtrlDefs"].([]any); ok {
				for _, d := range defs {
					dm := d.(map[string]any)
					if cs, ok := dm["controllers"].([]any); ok {
						for _, c := range cs {
							c.(map[string]any)["params"] = []any{}
						}
					}
					if cs, _ := dm["controllers"].([]any); len(cs) == 0 {
						dm["controllers"] = []any{}
					}
				}
			} else {
				generic["bgCtrlDefs"] = []any{}
			}
			raw, _ = json.Marshal(generic)
			var back Stage
			if err := json.Unmarshal(raw, &back); err != nil {
				t.Fatal(err)
			}
			out, err := SerializeDef([]byte(src), back)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, []byte(src)) {
				t.Errorf("not byte-exact:\n%q\nwant\n%q", out, src)
			}
		})
	}
}

// emptyNonNil returns a deep copy of s with every nil slice/map reachable
// from it replaced by an empty non-nil one (built by reflection, so a new
// collection field is covered automatically).
func emptyNonNil(s Stage) Stage {
	var fill func(v reflect.Value)
	fill = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Slice:
			if v.IsNil() {
				v.Set(reflect.MakeSlice(v.Type(), 0, 0))
			}
			for i := 0; i < v.Len(); i++ {
				fill(v.Index(i))
			}
		case reflect.Map:
			if v.IsNil() {
				v.Set(reflect.MakeMap(v.Type()))
			}
			for _, k := range v.MapKeys() {
				e := reflect.New(v.Type().Elem()).Elem()
				e.Set(v.MapIndex(k))
				fill(e)
				v.SetMapIndex(k, e)
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanSet() {
					fill(v.Field(i))
				}
			}
		}
	}
	// deep copy through the normalizer's inverse-free path: copy, then fill.
	cp := reflect.New(reflect.TypeOf(s)).Elem()
	copyDeep(cp, reflect.ValueOf(s))
	fill(cp)
	return cp.Interface().(Stage)
}

func copyDeep(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Slice:
		if src.IsNil() {
			return
		}
		n := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
		for i := 0; i < src.Len(); i++ {
			copyDeep(n.Index(i), src.Index(i))
		}
		dst.Set(n)
	case reflect.Map:
		if src.IsNil() {
			return
		}
		n := reflect.MakeMap(src.Type())
		for _, k := range src.MapKeys() {
			e := reflect.New(src.Type().Elem()).Elem()
			copyDeep(e, src.MapIndex(k))
			n.SetMapIndex(k, e)
		}
		dst.Set(n)
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if dst.Field(i).CanSet() {
				copyDeep(dst.Field(i), src.Field(i))
			}
		}
	default:
		dst.Set(src)
	}
}

const allCollectionsSrc = "[Info]\nname = x\n\n[BGdef]\nspr = a.sff\n\n[BG bg1]\ntype = normal\nspriteno = 0, 0\n\n[Begin Action 5]\n\n[BGCtrlDef a]\nlooptime = 4\n\n[BGCtrlDef b]\nlooptime = 2\n[BGCtrl 1]\ntype = Visible\ntime = 1\n\n[BGCtrl 2]\ntype = Null\n"

func TestSerializeDef_AllNilCollectionsEmpty_ByteExact(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader(allCollectionsSrc))
	if err != nil {
		t.Fatal(err)
	}
	st := emptyNonNil(doc.Stage)
	if reflect.DeepEqual(st, doc.Stage) {
		t.Fatal("fixture has no nil collection; test is vacuous")
	}
	out, err := SerializeDef([]byte(allCollectionsSrc), st)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, []byte(allCollectionsSrc)) {
		t.Errorf("not byte-exact:\n%q\nwant\n%q", out, allCollectionsSrc)
	}
}

func TestSerializeDef_DoesNotMutateCallerStage(t *testing.T) {
	s, err := Parse(strings.NewReader(allCollectionsSrc))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SerializeDef([]byte(allCollectionsSrc), s); err != nil {
		t.Fatal(err)
	}
	want, _ := Parse(strings.NewReader(allCollectionsSrc))
	if !reflect.DeepEqual(s, want) {
		t.Errorf("caller Stage mutated:\n%+v\nwant\n%+v", s, want)
	}
}
