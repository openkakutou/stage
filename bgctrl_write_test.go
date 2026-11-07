package stage

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

const bgctrlFixture = "mugen-bgctrl-stage.def"

func mustParseBytes(t *testing.T, data []byte) Stage {
	t.Helper()
	s, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return s
}

func TestSerialize_BGCtrlDefs_WritesPinnedText(t *testing.T) {
	s := Stage{BGCtrlDefs: []BGCtrlDef{{
		Name: "Lights", CtrlID: 3, LoopTime: 100,
		Controllers: []BGCtrl{{Type: "visible", CtrlID: 3, Time: 5, Params: []BGCtrlParam{{Key: "value", Value: "0"}}}},
	}}}
	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatal(err)
	}
	want := "[BGCtrlDef Lights]\nctrlID = 3\nlooptime = 100\n\n[BGCtrl 1]\ntype = visible\nctrlID = 3\ntime = 5\nvalue = 0\n\n"
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("missing block in output:\n%s", buf.String())
	}
}

func TestSerialize_BGCtrlDefs_OmitsZeroKeysAndEmptyName(t *testing.T) {
	s := Stage{BGCtrlDefs: []BGCtrlDef{{Controllers: []BGCtrl{{Type: "enable"}}}}}
	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "[BGCtrlDef]\n\n[BGCtrl 1]\ntype = enable\n\n") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	for _, k := range []string{"ctrlID", "looptime", "time"} {
		if strings.Contains(out, k) {
			t.Errorf("zero %s must be omitted:\n%s", k, out)
		}
	}
}

func TestSerialize_BGCtrlDefs_RoundTripsThroughParse(t *testing.T) {
	want := []BGCtrlDef{
		{Name: "Lights", CtrlID: 1, LoopTime: 60, Controllers: []BGCtrl{
			{Type: "palfx", Time: 0, Params: []BGCtrlParam{{Key: "time", Value: "60,90,120"}, {Key: "add", Value: "10, 20, 30"}}},
			{Type: "visible", CtrlID: 2, Time: 7, Params: []BGCtrlParam{{Key: "value", Value: " padded "}, {Key: "q", Value: `"inner"`}, {Key: "empty", Value: ""}}},
		}},
		{Name: "", CtrlID: 0, LoopTime: 0},
		{Name: "Lights", LoopTime: 5, Controllers: []BGCtrl{{Type: "enable"}}},
	}
	var buf bytes.Buffer
	if err := Serialize(&buf, Stage{BGCtrlDefs: want}); err != nil {
		t.Fatal(err)
	}
	got := mustParseBytes(t, buf.Bytes()).BGCtrlDefs
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch\n got: %#v\nwant: %#v\ntext:\n%s", got, want, buf.String())
	}
}

func TestDocumentAndSerializeDef_BGCtrlFixture_UnmodifiedIsByteExact(t *testing.T) {
	lf := readTestdataFile(t, bgctrlFixture)
	crlf := bytes.ReplaceAll(lf, []byte("\n"), []byte("\r\n"))
	for name, data := range map[string][]byte{"LF": lf, "CRLF": crlf} {
		t.Run(name, func(t *testing.T) {
			doc, err := ParseDocument(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if len(doc.Stage.BGCtrlDefs) != 3 {
				t.Fatalf("fixture should hold 3 groups, got %d", len(doc.Stage.BGCtrlDefs))
			}
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf.Bytes(), data) {
				t.Error("Document.Serialize is not byte-exact")
			}
			out, err := SerializeDef(data, doc.Stage)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, data) {
				t.Error("SerializeDef of unmodified stage is not byte-exact")
			}
		})
	}
}

// editFixture parses the fixture, applies edit and runs SerializeDef.
func editFixture(t *testing.T, data []byte, edit func(*Stage)) string {
	t.Helper()
	s := mustParseBytes(t, data)
	edit(&s)
	out, err := SerializeDef(data, s)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestSerializeDef_BGCtrlValueEdit_ChangesOnlyThatLine(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs[0].Controllers[0].Params[0].Value = "1"
	})
	want := strings.Replace(string(data), "value = 0   ; hidden", "value = 1   ; hidden", 1)
	if got != want {
		t.Fatalf("unexpected patch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSerializeDef_EditInSecondOfThreeControllers_KeepsOthersAndCommentsIntact(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		c := &s.BGCtrlDefs[0].Controllers[1]
		c.Params[1].Value = "11, 22, 33" // add
	})
	want := strings.Replace(string(data), "add = 10, 20, 30  ; brighten", "add = 11, 22, 33  ; brighten", 1)
	if got != want {
		t.Fatalf("unexpected patch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSerializeDef_TimeEdit_ReplacesTypedKeyAndKeepsCRLF(t *testing.T) {
	data := bytes.ReplaceAll(readTestdataFile(t, bgctrlFixture), []byte("\n"), []byte("\r\n"))
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs[0].Controllers[0].Time = 45
		s.BGCtrlDefs[0].LoopTime = 99
	})
	want := strings.Replace(string(data), "time = 30\r\n", "time = 45\r\n", 1)
	want = strings.Replace(want, "looptime = 120\r\n", "looptime = 99\r\n", 1)
	if got != want {
		t.Fatalf("unexpected patch\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestSerializeDef_KeyAdditionAndRemoval_AppendsAfterLastKeyLineAndDeletesLine(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		c := &s.BGCtrlDefs[0].Controllers[2] // Shake
		c.Params = append(c.Params, BGCtrlParam{Key: "ignorehit", Value: "1"})
		f := &s.BGCtrlDefs[0].Controllers[1] // Fade: drop mul
		f.Params = f.Params[:2]
	})
	want := strings.Replace(string(data), "mul = 256,256,256\n", "", 1)
	want = strings.Replace(want, "value = 2\n", "value = 2\nignorehit = 1\n", 1)
	if got != want {
		t.Fatalf("unexpected patch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSerializeDef_NewTypedKey_IsAppendedToSection(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs[1].CtrlID = 4 // group "Lights" #2 had no ctrlID
		s.BGCtrlDefs[1].Controllers[0].CtrlID = 8
	})
	want := strings.Replace(string(data), "looptime = 240\n", "looptime = 240\nctrlID = 4\n", 1)
	want = strings.Replace(want, "value = \"1\"\n", "value = \"1\"\nctrlID = 8\n", 1)
	if got != want {
		t.Fatalf("unexpected patch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSerializeDef_QuotedValueEdit_KeepsOutputParseable(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	var edited Stage
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs[1].Controllers[0].Params[0].Value = " spaced "
		edited = *s
	})
	if !strings.Contains(got, `value = " spaced "`) {
		t.Fatalf("expected quoted value, got:\n%s", got)
	}
	if !reflect.DeepEqual(mustParseBytes(t, []byte(got)).BGCtrlDefs, edited.BGCtrlDefs) {
		t.Error("patched text does not parse back to the edited controllers")
	}
	if !strings.HasPrefix(got, string(data[:strings.Index(string(data), "[BGCtrlDef Lights]")])) {
		t.Error("text before the groups changed")
	}
}

func TestSerializeDef_ControllerCountChange_FallsBackToFreshSerialize(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs[0].Controllers = s.BGCtrlDefs[0].Controllers[:2]
	})
	if strings.Contains(got, "; toggled on a timer") {
		t.Error("count change must use the fresh write path (comments dropped)")
	}
	if !strings.Contains(got, "[BGCtrl 1]\ntype = visible") {
		t.Errorf("expected fresh BGCtrl blocks, got:\n%s", got)
	}
}

func TestSerializeDef_NonBGCtrlEditAlongsideBGCtrl_FallsBackToFreshSerialize(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		s.Name = "Renamed"
		s.BGCtrlDefs[0].Controllers[0].Time = 1
	})
	if strings.Contains(got, "; trailing comment") || !strings.Contains(got, `name = "Renamed"`) {
		t.Errorf("expected fresh serialization, got:\n%s", got)
	}
}

func TestSerializeDef_GroupCountChange_FallsBackToFreshSerialize(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs = s.BGCtrlDefs[:2]
	})
	if strings.Contains(got, "; lights controller") || strings.Contains(got, "[BGCtrlDef Empty]") {
		t.Errorf("expected fresh serialization of two groups, got:\n%s", got)
	}
}

func TestSerializeDef_UnpatchableValue_FallsBackToFreshSerialize(t *testing.T) {
	data := readTestdataFile(t, bgctrlFixture)
	got := editFixture(t, data, func(s *Stage) {
		s.BGCtrlDefs[0].Controllers[0].Params[0].Value = "a;b"
	})
	if strings.Contains(got, "; toggled on a timer") {
		t.Error("an unreconcilable edit must not be patched in place")
	}
}

func TestSerializeDef_MalformedOriginal_ReturnsError(t *testing.T) {
	_, err := SerializeDef([]byte("[BGCtrlDef x\nlooptime = 1\n"), Stage{})
	if err == nil {
		t.Fatal("expected an error for a malformed original")
	}
}

const looptimeParamSrc = "[Info]\nname = \"Test\"\n\n; group comment\n[BGCtrlDef a]\nctrlID = 1 ; group id\n\n; controller comment\n[BGCtrl 1]\ntype = visible\nlooptime = 30 ; a param here\nvalue = 1 ; keep me\n\n[Shadow]\ncolor = 0,0,0\n"

func editLooptimeParamStage(t *testing.T, mutate func(c *BGCtrl)) string {
	t.Helper()
	st := mustParseBytes(t, []byte(looptimeParamSrc))
	mutate(&st.BGCtrlDefs[0].Controllers[0])
	out, err := SerializeDef([]byte(looptimeParamSrc), st)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestSerializeDef_BGCtrlLooptimeParam_EditOtherParam_PatchesSingleLine(t *testing.T) {
	got := editLooptimeParamStage(t, func(c *BGCtrl) {
		if len(c.Params) != 2 || c.Params[0].Key != "looptime" || c.Params[1].Key != "value" {
			t.Fatalf("unexpected params %+v", c.Params)
		}
		c.Params[1].Value = "0"
	})
	want := strings.Replace(looptimeParamSrc, "value = 1 ; keep me", "value = 0 ; keep me", 1)
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSerializeDef_BGCtrlLooptimeParam_EditLooptimeValue_PatchesSingleLine(t *testing.T) {
	got := editLooptimeParamStage(t, func(c *BGCtrl) { c.Params[0].Value = "45" })
	want := strings.Replace(looptimeParamSrc, "looptime = 30 ; a param here", "looptime = 45 ; a param here", 1)
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
