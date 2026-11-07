package stage

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// BGLayerState is one BG element's controller-driven state at a tick, as
// returned by ResolveBGCtrlState.
type BGLayerState struct {
	// Name is the element's section name (BGElement.Name).
	Name string `json:"name"`
	// Visible is false once a "visible" controller with value 0 is in
	// effect; true by default.
	Visible bool `json:"visible"`
	// Enabled is false once an "enabled" controller with value 0 is in
	// effect; true by default.
	Enabled bool `json:"enabled"`
	// OffsetX and OffsetY are the position offset added to the element's
	// start position by posset/posadd/velset/veladd; 0 by default.
	OffsetX float64 `json:"offsetX"`
	OffsetY float64 `json:"offsetY"`
	// AnimAction is the action number set by an "anim" controller, or -1
	// when no controller overrides the element's own action.
	AnimAction int `json:"animAction"`
}

// ResolveBGCtrlState resolves, for every element of s (same order as
// s.Elements), the state its BGCtrl controllers put it in at tick. Pure and
// total: it never errors or panics. Rules (see decision 012):
//
//   - Group tick t = tick mod LoopTime when LoopTime > 0, else tick; a
//     negative tick counts as 0.
//   - A controller targets an element when its effective ID (its own CtrlID
//     if nonzero, else its group's) is 0 (all elements) or equals the
//     element's ID.
//   - A single "time" T is in effect once t >= T. A "time" list
//     "start,end,loop" (kept raw in Params) is in effect while
//     start <= t' < end, t' = t mod loop when loop > 0; an omitted end means
//     "no end".
//   - Controllers in effect are folded in chronological order of when they
//     last fired (window start, shifted to the current loop period), file
//     order breaking ties, so the latest one wins for an attribute. Groups
//     are folded in file order.
//   - Handled types: visible, enabled (also spelled "enable", as the vendored
//     fixture does), posset, posadd, velset, veladd, anim.
//     Anything else, or an unparsable value, is ignored.
func ResolveBGCtrlState(s Stage, tick int) []BGLayerState {
	states := make([]BGLayerState, len(s.Elements))
	if tick < 0 {
		tick = 0
	}
	for i, el := range s.Elements {
		st := BGLayerState{Name: el.Name, Visible: true, Enabled: true, AnimAction: -1}
		var dx, dy float64
		for _, g := range s.BGCtrlDefs {
			gx, gy := resolveGroup(&st, g, el.ID, tick)
			dx += gx
			dy += gy
		}
		// Velocity displacement rides on top of the folded base position.
		st.OffsetX = finite(st.OffsetX + dx)
		st.OffsetY = finite(st.OffsetY + dy)
		states[i] = st
	}
	return states
}

// finite clamps an overflowed offset to ±MaxFloat64 (NaN to 0), so the result
// always survives JSON encoding.
func finite(v float64) float64 {
	switch {
	case math.IsNaN(v):
		return 0
	case math.IsInf(v, 1):
		return math.MaxFloat64
	case math.IsInf(v, -1):
		return -math.MaxFloat64
	}
	return v
}

// ctrlTargets reports whether controller c of group g applies to an element
// whose ID is elemID: the effective ID (c's own if nonzero, else g's) is 0
// (everything) or equals elemID.
func ctrlTargets(g BGCtrlDef, c BGCtrl, elemID int) bool {
	id := c.CtrlID
	if id == 0 {
		id = g.CtrlID
	}
	return id == 0 || id == elemID
}

// resolveGroup folds the controllers of g that target an element into st at
// the tick and returns the velocity displacement of that group.
func resolveGroup(st *BGLayerState, g BGCtrlDef, elemID, tick int) (dx, dy float64) {
	t := tick
	if g.LoopTime > 0 {
		t = tick % g.LoopTime
	}
	var targeted []BGCtrl
	for _, c := range g.Controllers {
		if ctrlTargets(g, c, elemID) {
			targeted = append(targeted, c)
		}
	}
	all := timedCtrls(targeted)
	posStart := 0
	for _, f := range firedAt(all, t) {
		applyBGCtrl(st, f.c)
		if f.c.Type == "posset" {
			posStart = f.start
		}
	}
	var vel []timedCtrl
	for _, tc := range all {
		if tc.c.Type == "velset" || tc.c.Type == "veladd" {
			vel = append(vel, tc)
		}
	}
	// A posset restarts the position, so only the displacement since the
	// latest posset counts.
	return velocityDisplacement(vel, posStart, t)
}

// ctrlWindow is a controller's activity window on the group tick: in effect
// while start <= t' < end (end < 0 = no end), t' = t mod loop when loop > 0.
type ctrlWindow struct {
	start, end, loop int
}

// timedCtrl is a controller with its parsed window. start is filled by
// firedAt: the tick the controller last fired.
type timedCtrl struct {
	c     BGCtrl
	w     ctrlWindow
	start int
}

// timedCtrls pairs each controller with its window, dropping those whose
// "time" list is unparsable.
func timedCtrls(ctrls []BGCtrl) []timedCtrl {
	var out []timedCtrl
	for _, c := range ctrls {
		if w, ok := bgCtrlWindow(c); ok {
			out = append(out, timedCtrl{c: c, w: w})
		}
	}
	return out
}

// firedAt returns the controllers in effect at tick t, each with start set
// to when it last fired, ordered oldest first (input order breaks ties).
func firedAt(tcs []timedCtrl, t int) []timedCtrl {
	var out []timedCtrl
	for _, tc := range tcs {
		if tc.w.activeAt(t) {
			tc.start = tc.w.lastStart(t)
			out = append(out, tc)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].start < out[j].start })
	return out
}

// bgCtrlWindow reads c's timing: the raw "time" list from Params when
// present, else the single Time. ok is false when the list is unparsable.
func bgCtrlWindow(c BGCtrl) (w ctrlWindow, ok bool) {
	w = ctrlWindow{start: c.Time, end: -1}
	for _, p := range c.Params {
		if p.Key != "time" {
			continue
		}
		parts := strings.Split(p.Value, ",")
		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return w, false
		}
		w.start = start
		for j, dst := range []*int{&w.end, &w.loop} {
			if j+1 >= len(parts) || strings.TrimSpace(parts[j+1]) == "" {
				continue
			}
			n, err := strconv.Atoi(strings.TrimSpace(parts[j+1]))
			if err != nil {
				return w, false
			}
			*dst = n
		}
		return w, true
	}
	return w, true
}

// activeAt reports whether the window is in effect at group tick t.
func (w ctrlWindow) activeAt(t int) bool {
	if w.loop > 0 {
		t %= w.loop
	}
	return t >= w.start && (w.end < 0 || t < w.end)
}

// lastStart is the tick, at most t, where the window's current activation
// began (its start shifted to the current loop period).
func (w ctrlWindow) lastStart(t int) int {
	if w.loop > 0 {
		return t - t%w.loop + w.start
	}
	return w.start
}

// ctrlParamFloat returns the lenient float value of key in c's Params.
func ctrlParamFloat(c BGCtrl, key string) (float64, bool) {
	for _, p := range c.Params {
		if p.Key == key {
			f, err := strconv.ParseFloat(strings.TrimSpace(p.Value), 64)
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				return 0, false
			}
			return f, true
		}
	}
	return 0, false
}

// applyBGCtrl folds controller c, known to be in effect, into st. Types
// without a base-state effect (velset, veladd, unknown) are ignored.
func applyBGCtrl(st *BGLayerState, c BGCtrl) {
	switch c.Type {
	case "visible":
		if v, ok := ctrlParamFloat(c, "value"); ok {
			st.Visible = v != 0
		}
	case "enabled", "enable":
		if v, ok := ctrlParamFloat(c, "value"); ok {
			st.Enabled = v != 0
		}
	case "anim":
		if v, ok := ctrlParamFloat(c, "value"); ok {
			st.AnimAction = int(v)
		}
	case "posset":
		x, _ := ctrlParamFloat(c, "x")
		y, _ := ctrlParamFloat(c, "y")
		st.OffsetX, st.OffsetY = x, y
	case "posadd":
		x, _ := ctrlParamFloat(c, "x")
		y, _ := ctrlParamFloat(c, "y")
		st.OffsetX += x
		st.OffsetY += y
	}
}

// velocityDisplacement integrates the velset/veladd controllers vel of one
// group over the ticks from..t-1: the sum over u of the velocity at tick u,
// itself folded from the controllers in effect at u in chronological order
// (velset replaces, veladd adds). It is computed in closed form per
// constant-velocity segment.
// shortcut: segment count grows with t/loop for a looped velocity time
// list, upgrade to a period reduction if huge ticks without a group
// looptime matter.
func velocityDisplacement(vel []timedCtrl, from, t int) (dx, dy float64) {
	bounds := velBoundaries(vel, from, t)
	for i := 0; i+1 < len(bounds); i++ {
		vx, vy := velocityAt(vel, bounds[i])
		n := float64(bounds[i+1] - bounds[i])
		dx += vx * n
		dy += vy * n
	}
	return dx, dy
}

// velBoundaries returns the sorted ticks in [from, t] where the velocity can
// change: from, t, and every window start, end and loop period start
// strictly between them.
func velBoundaries(vel []timedCtrl, from, t int) []int {
	if from >= t {
		return nil
	}
	set := map[int]bool{from: true, t: true}
	add := func(p int) {
		if p > from && p < t {
			set[p] = true
		}
	}
	for _, v := range vel {
		w := v.w
		if w.loop <= 0 {
			add(w.start)
			add(w.end)
			continue
		}
		for k := from - from%w.loop; k < t; k += w.loop {
			add(k)
			add(k + w.start)
			add(k + w.end)
		}
	}
	bounds := make([]int, 0, len(set))
	for p := range set {
		bounds = append(bounds, p)
	}
	sort.Ints(bounds)
	return bounds
}

// velocityAt folds the velocity controllers in effect at tick u.
func velocityAt(vel []timedCtrl, u int) (vx, vy float64) {
	for _, f := range firedAt(vel, u) {
		x, _ := ctrlParamFloat(f.c, "x")
		y, _ := ctrlParamFloat(f.c, "y")
		if f.c.Type == "velset" {
			vx, vy = x, y
		} else {
			vx += x
			vy += y
		}
	}
	return vx, vy
}
