package agent

// Editor edits on top of the host-rendered config ("base + edits", AGE-540).
//
// The OpAMP server keeps the changes a user makes in the bkoff Config tab as a
// path-based diff against the config this agent renders from the ingestion
// rules (the base), and sends it next to the rules as the `mw.edits` file.
// `mw-agent format --edits` applies it on a copy of the rendered rules, after
// ECS and before the feature restrictions and host tags, and validates the
// result. Anything wrong with the edits leaves the agent on the base.
//
// The rules below are the spec in mw-vault
// 13-opamp-remote-management/plan-persistent-config-edits.md, "Data and
// protocol" -> "Exact edit rules"; the server's diff (step 4 of that plan)
// produces these documents, so keep both in sync.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// Results of one edit entry, reported in EditsStatus.Ops.
const (
	EditApplied  = "applied"
	EditSkipped  = "skipped"
	EditConflict = "conflict"
	EditRefused  = "refused"
)

// maxEditsErrorLen caps EditsStatus.Error (in characters).
const maxEditsErrorLen = 2048

// ConfigRender is the output of RenderOtelConfig.
type ConfigRender struct {
	// Base is the config rendered from the ingestion rules alone; it is what
	// BuildOtelConfig returns.
	Base []byte
	// Final is what the agent runs: Base with the edits applied, or Base when
	// there are no edits or they could not be used.
	Final []byte
	// Status is nil when no edits were given.
	Status *EditsStatus
}

// EditsStatus is the `--edits-status` file (reported to the server as
// `mw.edits.status`).
type EditsStatus struct {
	// EditsHash is the lowercase hex sha256 of the edits file bytes.
	EditsHash string `json:"edits_hash"`
	// Applied is true when Final contains the edits, even if some entries
	// were skipped or refused.
	Applied bool           `json:"applied"`
	Error   string         `json:"error"`
	Ops     []EditOpResult `json:"ops"`
}

// EditOpResult is the result of the edit entry at index I.
type EditOpResult struct {
	I      int    `json:"i"`
	Result string `json:"result"`
	Reason string `json:"reason,omitempty"`
}

// configEdit is one entry of the edits document.
type configEdit struct {
	Op    string          `json:"op"`
	Path  []string        `json:"path"`
	Value json.RawMessage `json:"value"`
}

// plannedEdit is a parsed, checked entry, ready to apply.
type plannedEdit struct {
	op   string
	path []string
	// comp is the length of the component path that path starts with.
	comp  int
	value interface{}
	// refused is the reason the entry is refused; it is then not applied.
	refused string
}

var editSections = map[string]bool{
	"receivers": true, "processors": true, "exporters": true,
	"extensions": true, "connectors": true, Service: true,
}

// ownedEditPaths are refused for every op: the supervisor and the agent own
// them and re-add them on every render.
var ownedEditPaths = []struct {
	path   []string
	reason string
}{
	{[]string{"extensions", "opamp"}, "extensions.opamp is owned by the supervisor"},
	{[]string{Service, "telemetry", "resource", "service.instance.id"}, "the instance id is owned by the supervisor"},
	{[]string{Processors, "resource/host_tags"}, "processors.resource/host_tags is owned by the agent"},
}

// hostTagsProcessor is dropped from list_set values; the host-tags step
// re-adds it.
const hostTagsProcessor = "resource/host_tags"

// editStepTestHook lets tests inject a panic into the edit step.
var editStepTestHook func()

// RenderOtelConfig renders the base config exactly as BuildOtelConfig does
// and, when edits is non-nil, the final config with the edits applied:
//
//	variant -> credential files -> ECS -> fork
//	  base:   restrictions -> host tags -> validate (may report track status)
//	  edited: edits -> restrictions -> host tags -> validate (no side effects)
//
// Only a base that fails to render or validate returns an error, even when the
// edits would fix it. Malformed edits, a panic in the edit step or an invalid
// edited config make Final = Base, with Status.Applied false and the error.
func (c *HostAgent) RenderOtelConfig(body []byte, configType string, edits []byte) (*ConfigRender, error) {
	apiYAMLConfig, err := c.renderRulesConfig(body, configType)
	if err != nil {
		return nil, err
	}

	// The restriction and host-tag steps modify their input, so the edited
	// branch gets its own copy before either runs.
	var edited map[string]interface{}
	if edits != nil {
		edited = deepCopyConfigValue(apiYAMLConfig).(map[string]interface{})
	}

	base, err := c.finishOtelConfig(apiYAMLConfig)
	if err != nil {
		return nil, err
	}
	render := &ConfigRender{Base: base, Final: base}
	if edits == nil {
		return render, nil
	}

	status := &EditsStatus{EditsHash: editsHash(edits), Ops: []EditOpResult{}}
	render.Status = status
	final, ops, err := c.renderEdited(edited, edits)
	if ops != nil {
		status.Ops = ops
	}
	if err != nil {
		status.Error = capEditsError(err.Error())
		c.logger.Warn("editor edits not applied; using the config rendered from the ingestion rules alone", zap.Error(err))
		return render, nil
	}
	status.Applied = true
	render.Final = final
	c.logger.Info("applied editor edits", zap.Int("entries", len(ops)), zap.Any("results", countEditResults(ops)))
	return render, nil
}

// renderEdited applies the edits to apiYAMLConfig and finishes it like the
// base, but validates without UpdateAgentTrackStatus. ops is nil when the
// document is malformed or the edit step panicked.
func (c *HostAgent) renderEdited(apiYAMLConfig map[string]interface{}, doc []byte) (out []byte, ops []EditOpResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			out, ops, err = nil, nil, fmt.Errorf("edit step panicked: %v", r)
		}
	}()

	planned, err := parseConfigEdits(doc)
	if err != nil {
		return nil, nil, err
	}
	if editStepTestHook != nil {
		editStepTestHook()
	}
	ops = applyConfigEdits(apiYAMLConfig, planned)

	out, err = c.overlayAndMarshal(apiYAMLConfig)
	if err != nil {
		return nil, ops, fmt.Errorf("edited config: %w", err)
	}
	invalid, err := c.validateOtelConfig(out)
	if err != nil {
		return nil, ops, fmt.Errorf("edited config is invalid: %w", err)
	}
	if invalid != nil {
		return nil, ops, fmt.Errorf("edited config is invalid: %w", invalid)
	}
	return out, ops, nil
}

func editsHash(doc []byte) string {
	sum := sha256.Sum256(doc)
	return hex.EncodeToString(sum[:])
}

func capEditsError(msg string) string {
	r := []rune(msg)
	if len(r) <= maxEditsErrorLen {
		return msg
	}
	return string(r[:maxEditsErrorLen])
}

func countEditResults(ops []EditOpResult) map[string]int {
	counts := map[string]int{}
	for _, op := range ops {
		counts[op.Result]++
	}
	return counts
}

// parseConfigEdits decodes and checks the whole document before anything is
// applied: one malformed entry rejects the document. Refused entries are
// marked, not rejected.
func parseConfigEdits(doc []byte) ([]plannedEdit, error) {
	trimmed := bytes.TrimSpace(doc)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, fmt.Errorf("malformed edits: not a JSON array")
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(trimmed, &entries); err != nil {
		return nil, fmt.Errorf("malformed edits: %w", err)
	}
	planned := make([]plannedEdit, 0, len(entries))
	for i, raw := range entries {
		var e configEdit
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("malformed edits: entry %d: %w", i, err)
		}
		p, err := planConfigEdit(e)
		if err != nil {
			return nil, fmt.Errorf("malformed edits: entry %d: %w", i, err)
		}
		planned = append(planned, p)
	}
	return planned, nil
}

func planConfigEdit(e configEdit) (plannedEdit, error) {
	p := plannedEdit{op: e.Op, path: e.Path}
	switch e.Op {
	case "set", "unset", "add", "remove", "list_add", "list_remove", "list_set":
	default:
		return p, fmt.Errorf("unknown op %q", e.Op)
	}
	if len(e.Path) == 0 {
		return p, fmt.Errorf("%s: empty path", e.Op)
	}
	for _, key := range e.Path {
		if key == "" {
			return p, fmt.Errorf("%s: empty key in path %q", e.Op, e.Path)
		}
	}
	if !editSections[e.Path[0]] {
		return p, fmt.Errorf("%s: unknown section %q", e.Op, e.Path[0])
	}
	p.comp = 2
	if e.Path[0] == Service && len(e.Path) >= 2 && e.Path[1] == Pipelines {
		p.comp = 3
	}

	for _, owned := range ownedEditPaths {
		if hasPathPrefix(e.Path, owned.path) {
			p.refused = owned.reason
			return p, nil
		}
	}

	// The lists of service: pipeline receivers/exporters and service.extensions
	// take list_add/list_remove; pipeline processors (ordered) take list_set.
	// Any other op on or under them is refused.
	if list, wantOp := editListAt(e.Path); list != nil {
		isListOp := e.Op == "list_add" || e.Op == "list_remove" || e.Op == "list_set"
		if len(e.Path) != len(list) || !isListOp || (wantOp == "list_set") != (e.Op == "list_set") {
			p.refused = fmt.Sprintf("%s does not apply to %s; use %s", e.Op, strings.Join(list, "."), wantOp)
			return p, nil
		}
	} else if strings.HasPrefix(e.Op, "list_") {
		return p, fmt.Errorf("%s: path %q is not a pipeline receivers/exporters/processors list or service.extensions", e.Op, e.Path)
	}

	switch e.Op {
	case "add", "remove":
		if len(e.Path) != p.comp {
			return p, fmt.Errorf("%s: path %q is not a component path", e.Op, e.Path)
		}
	case "set", "unset":
		if len(e.Path) <= p.comp {
			return p, fmt.Errorf("%s: path %q is not inside a component", e.Op, e.Path)
		}
	}

	var value interface{}
	if len(e.Value) > 0 {
		if err := json.Unmarshal(e.Value, &value); err != nil {
			return p, fmt.Errorf("%s: value: %w", e.Op, err)
		}
	}
	switch e.Op {
	case "add":
		if _, ok := value.(map[string]interface{}); !ok {
			return p, fmt.Errorf("add: value must be an object")
		}
	case "set":
		if len(e.Value) == 0 {
			return p, fmt.Errorf("set: value missing")
		}
	case "list_add", "list_remove":
		name, ok := value.(string)
		if !ok || name == "" {
			return p, fmt.Errorf("%s: value must be a non-empty string", e.Op)
		}
		if e.Path[1] == "extensions" && name == "opamp" {
			p.refused = "the opamp extension is owned by the supervisor"
		}
	case "list_set":
		items, ok := value.([]interface{})
		if !ok {
			return p, fmt.Errorf("list_set: value must be an array of strings")
		}
		kept := make([]interface{}, 0, len(items))
		for _, item := range items {
			name, ok := item.(string)
			if !ok || name == "" {
				return p, fmt.Errorf("list_set: value must be an array of strings")
			}
			if name != hostTagsProcessor {
				kept = append(kept, name)
			}
		}
		value = kept
	}
	p.value = value
	return p, nil
}

// editListAt returns the service list that path is on or under, and the op
// that list takes; nil when path is not on such a list.
func editListAt(path []string) ([]string, string) {
	if len(path) >= 2 && path[0] == Service && path[1] == "extensions" {
		return path[:2], "list_add/list_remove"
	}
	if len(path) >= 4 && path[0] == Service && path[1] == Pipelines {
		switch path[3] {
		case Receivers, "exporters":
			return path[:4], "list_add/list_remove"
		case Processors:
			return path[:4], "list_set"
		}
	}
	return nil, ""
}

func hasPathPrefix(path, prefix []string) bool {
	if len(path) < len(prefix) {
		return false
	}
	for i := range prefix {
		if path[i] != prefix[i] {
			return false
		}
	}
	return true
}

// applyConfigEdits applies the planned entries in order and returns one result
// per entry.
func applyConfigEdits(cfg map[string]interface{}, planned []plannedEdit) []EditOpResult {
	ops := make([]EditOpResult, 0, len(planned))
	for i, p := range planned {
		res := EditOpResult{I: i}
		if p.refused != "" {
			res.Result, res.Reason = EditRefused, p.refused
		} else {
			res.Result, res.Reason = applyConfigEdit(cfg, p)
		}
		ops = append(ops, res)
	}
	return ops
}

func applyConfigEdit(cfg map[string]interface{}, p plannedEdit) (string, string) {
	compName := strings.Join(p.path[:p.comp], "/")
	notInBase := func() (string, string) {
		return EditSkipped, fmt.Sprintf("component %s not in base", compName)
	}

	switch p.op {
	case "add":
		parent, ok := editWalk(cfg, p.path[:p.comp-1], true)
		if !ok {
			return EditSkipped, fmt.Sprintf("%s is not a map", strings.Join(p.path[:p.comp-1], "."))
		}
		name := p.path[p.comp-1]
		_, exists := editGet(parent, name)
		editPut(parent, name, p.value)
		if exists {
			return EditConflict, fmt.Sprintf("component %s already in base; replaced by the edit", compName)
		}
		return EditApplied, ""

	case "remove":
		parent, ok := editWalk(cfg, p.path[:p.comp-1], false)
		if !ok {
			return notInBase()
		}
		name := p.path[p.comp-1]
		if _, exists := editGet(parent, name); !exists {
			return notInBase()
		}
		editDelete(parent, name)
		return EditApplied, ""

	case "set", "unset":
		comp, exists := editComponent(cfg, p.path[:p.comp])
		if !exists {
			return notInBase()
		}
		if comp == nil {
			return EditSkipped, fmt.Sprintf("component %s is not a map", compName)
		}
		inner, key := p.path[p.comp:len(p.path)-1], p.path[len(p.path)-1]
		parent, ok := editWalk(comp, inner, p.op == "set")
		if !ok {
			if p.op == "set" {
				return EditSkipped, fmt.Sprintf("path %s crosses a non-map", strings.Join(p.path, "."))
			}
			return EditSkipped, fmt.Sprintf("%s not in base", strings.Join(p.path, "."))
		}
		if p.op == "set" {
			editPut(parent, key, p.value)
			return EditApplied, ""
		}
		if _, exists := editGet(parent, key); !exists {
			return EditSkipped, fmt.Sprintf("%s not in base", strings.Join(p.path, "."))
		}
		editDelete(parent, key)
		return EditApplied, ""

	case "list_add", "list_remove", "list_set":
		listPath := p.path
		owner, ok := editWalk(cfg, listPath[:len(listPath)-1], false)
		if !ok {
			if p.comp == 3 {
				return notInBase()
			}
			return EditSkipped, fmt.Sprintf("%s not in base", strings.Join(listPath[:len(listPath)-1], "."))
		}
		key := listPath[len(listPath)-1]
		if p.op == "list_set" {
			editPut(owner, key, p.value)
			return EditApplied, ""
		}
		current, _ := editGet(owner, key)
		items, ok := editListItems(current)
		if !ok {
			return EditSkipped, fmt.Sprintf("%s is not a list", strings.Join(listPath, "."))
		}
		name := p.value.(string)
		idx := -1
		for i, item := range items {
			if item == name {
				idx = i
				break
			}
		}
		if p.op == "list_add" {
			if idx >= 0 {
				return EditSkipped, fmt.Sprintf("%s already in %s", name, strings.Join(listPath, "."))
			}
			editPut(owner, key, append(items, name))
			return EditApplied, ""
		}
		if idx < 0 {
			return EditSkipped, fmt.Sprintf("%s not in %s", name, strings.Join(listPath, "."))
		}
		editPut(owner, key, append(items[:idx:idx], items[idx+1:]...))
		return EditApplied, ""
	}
	return EditSkipped, fmt.Sprintf("unknown op %q", p.op)
}

// editComponent returns the map of the component at compPath, or nil when the
// component exists but is not a map. A component whose value is null (e.g.
// `nop:`) counts as an empty map.
func editComponent(cfg map[string]interface{}, compPath []string) (interface{}, bool) {
	parent, ok := editWalk(cfg, compPath[:len(compPath)-1], false)
	if !ok {
		return nil, false
	}
	name := compPath[len(compPath)-1]
	comp, exists := editGet(parent, name)
	if !exists {
		return nil, false
	}
	if comp == nil {
		comp = map[string]interface{}{}
		editPut(parent, name, comp)
	}
	if !editIsMap(comp) {
		return nil, true
	}
	return comp, true
}

// editWalk follows keys from m and returns the map at the end. With create,
// missing or null keys become empty maps; it fails when a key is missing
// (without create) or holds a non-map.
func editWalk(m interface{}, keys []string, create bool) (interface{}, bool) {
	cur := m
	for _, key := range keys {
		next, exists := editGet(cur, key)
		if !exists || next == nil {
			if !create {
				return nil, false
			}
			next = map[string]interface{}{}
			editPut(cur, key, next)
		}
		if !editIsMap(next) {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

// The rendered config holds map[string]interface{} (rules JSON) and
// map[interface{}]interface{} (credential files read with yaml.v2).

func editIsMap(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}, map[interface{}]interface{}:
		return true
	}
	return false
}

func editGet(m interface{}, key string) (interface{}, bool) {
	switch mm := m.(type) {
	case map[string]interface{}:
		v, ok := mm[key]
		return v, ok
	case map[interface{}]interface{}:
		v, ok := mm[key]
		return v, ok
	}
	return nil, false
}

func editPut(m interface{}, key string, v interface{}) {
	switch mm := m.(type) {
	case map[string]interface{}:
		mm[key] = v
	case map[interface{}]interface{}:
		mm[key] = v
	}
}

func editDelete(m interface{}, key string) {
	switch mm := m.(type) {
	case map[string]interface{}:
		delete(mm, key)
	case map[interface{}]interface{}:
		delete(mm, key)
	}
}

// editListItems returns a copy of a service list as []interface{}; a missing
// or null list is empty. ECS stores the metrics receivers as []string.
func editListItems(v interface{}) ([]interface{}, bool) {
	switch l := v.(type) {
	case nil:
		return []interface{}{}, true
	case []interface{}:
		return append([]interface{}{}, l...), true
	case []string:
		items := make([]interface{}, 0, len(l))
		for _, s := range l {
			items = append(items, s)
		}
		return items, true
	}
	return nil, false
}

// deepCopyConfigValue copies every map and slice type the render steps put
// into the config, so the base and edited branches share nothing.
func deepCopyConfigValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		if t == nil {
			return t
		}
		out := make(map[string]interface{}, len(t))
		for k, x := range t {
			out[k] = deepCopyConfigValue(x)
		}
		return out
	case map[interface{}]interface{}:
		if t == nil {
			return t
		}
		out := make(map[interface{}]interface{}, len(t))
		for k, x := range t {
			out[k] = deepCopyConfigValue(x)
		}
		return out
	case []interface{}:
		if t == nil {
			return t
		}
		out := make([]interface{}, len(t))
		for i, x := range t {
			out[i] = deepCopyConfigValue(x)
		}
		return out
	case []string:
		if t == nil {
			return t
		}
		return append([]string{}, t...)
	case []map[string]interface{}:
		if t == nil {
			return t
		}
		out := make([]map[string]interface{}, len(t))
		for i, x := range t {
			out[i] = deepCopyConfigValue(x).(map[string]interface{})
		}
		return out
	}
	return v
}
