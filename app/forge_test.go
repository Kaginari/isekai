package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/provider"
	"github.com/Kaginari/isekai/provider/mock"
	"github.com/Kaginari/isekai/tool"
	"github.com/Kaginari/isekai/wire"
	"github.com/Kaginari/isekai/world"
)

// forgeWorld is the isekai test world where elf-core wears ttl-craft (its doc names the Mind),
// dispatch reaches three hops, and the provider is a recording mock the test scripts.
func forgeWorld(t *testing.T) (*testWorld, *App) {
	t.Helper()
	files := isekaiCreatures()
	files[".isekai/elf/core/README.md"] = "# elf-core\n\n- **Rank:** Elf\n- **Territory:** `src/`\n- **Minds:** ttl-craft\n- **Purpose:** the shared mind\n"
	w := newTestWorld(t, "isekai", files)
	w.script(".isekai/tmp/s.json", text("nothing"))
	w.write(".isekai/config.yaml", "models: {default: mock/m}\nproviders: {mock: {enabled: true, script: .isekai/tmp/s.json}}\ntools: {bash: {sandbox: none}, dispatch: {maxDepth: 3}}\nui: {board: {autostart: false}}\nlaw: {forge: {demandAtSteps: 0}}\n")
	return w, w.open()
}

// throne runs one ask on the throne with every body answering from the same recording mock.
func throne(t *testing.T, a *App, ask string, script ...provider.Response) *mock.Provider {
	t.Helper()
	m := mock.New(script...)
	b := a.Build()
	b.Provider, b.Models = m, nil
	e := a.World.Engine(world.Rimuru, b)
	defer a.World.Close(e)
	if _, err := e.Run(context.Background(), ask); err != nil {
		t.Fatal(err)
	}
	return m
}

func dispatchCall(id, body, ask string) provider.Response {
	return mock.Call(id, "dispatch", map[string]interface{}{"body": body, "ask": ask})
}

func forgeNotes(t *testing.T, w *testWorld) []world.ForgeNote {
	t.Helper()
	return world.ReadForgeLedger(filepath.Join(w.root, ".isekai", "memory", "shared", "notes.jsonl"))
}

func desk(a *App, args ...string) (int, string, string) {
	var out, errw bytes.Buffer
	code := a.forgeRun(&out, &errw, args, nil, "/forge")
	return code, out.String(), errw.String()
}

var deskID = regexp.MustCompile(`(?m)^([0-9a-f]{7})  (\S+)`)

// idOf finds a listed draft's id by name.
func idOf(t *testing.T, a *App, name string) string {
	t.Helper()
	_, out, _ := desk(a, "list")
	for _, m := range deskID.FindAllStringSubmatch(out, -1) {
		if m[2] == name {
			return m[1]
		}
	}
	t.Fatalf("%s not on the desk:\n%s", name, out)
	return ""
}

// The forge end to end: a Court's draft is filed from its report, the desk lists it by a
// stable id, the mint writes a real SKILL.md in Claude Code's format, and a later Court that
// wears the Mind carries `@T mind ttl-craft` in its prompt while skill{name} returns the body.
func TestForgeDraftMintWear(t *testing.T) {
	w, a := forgeWorld(t)
	throne(t, a, "work the token ttl",
		dispatchCall("d1", "elf-core", "findings on the token ttl"),
		mock.Text("@S DONE\n@F src/auth/login.go:1 ttl read\n@FORGE skill:ttl-craft | Use when a token TTL is set, read or questioned | 1. read src/auth/login.go 2. keep the ttl at one hour 3. say where it is set\n@U colony the ttl is an hour\n@E 0"),
		mock.Text("done"))
	notes := forgeNotes(t, w)
	if len(notes) != 1 || notes[0].By != "elf-core" || notes[0].Run == "" || notes[0].Hash == "" {
		t.Fatalf("one draft filed by its body with its run and hash: %+v", notes)
	}
	id := idOf(t, a, "ttl-craft")
	if !strings.HasPrefix(notes[0].Hash, id) {
		t.Fatalf("the desk's id is the draft's hash: %s vs %s", id, notes[0].Hash)
	}
	if code, out, errs := desk(a, "mint", id); code != 0 || !strings.Contains(out, "minted skill:ttl-craft v1") {
		t.Fatalf("mint: %d\n%s\n%s", code, out, errs)
	}
	skill := w.read(".isekai/crafted-skills/ttl-craft/SKILL.md")
	for _, want := range []string{"---\nname: ttl-craft\n", "description: \"Use when a token TTL is set, read or questioned\"", "version: 1\n", "status: active\n", "by: \"elf-core\"", "born: \"" + notes[0].Run + "\"", "minted: {by: \"veldora\"", "# ttl-craft\n\n1. read src/auth/login.go"} {
		if !strings.Contains(skill, want) {
			t.Fatalf("SKILL.md lacks %q:\n%s", want, skill)
		}
	}
	if _, out, _ := desk(a, "list"); strings.Contains(out, "ttl-craft") {
		t.Fatalf("a minted draft leaves the desk:\n%s", out)
	}
	// a later Court that wears it: its prompt carries the Mind
	m := throne(t, a, "work the token ttl again",
		dispatchCall("d1", "elf-core", "findings on the token ttl, again"),
		mock.Text("@S DONE\n@F src/auth/login.go:1 still an hour\n@U colony nothing new\n@E 0"),
		mock.Text("done"))
	court := m.Requests[1]
	all := court.System + court.SystemTail
	for _, msg := range court.Messages {
		all += "\n" + msg.Text
	}
	if !strings.Contains(all, "@T mind ttl-craft") {
		t.Fatalf("the wearing Court's prompt must carry @T mind ttl-craft:\n%s", all)
	}
	res := a.skillTool("elf-core").Run(context.Background(), tool.Env{Root: a.Root, WorldDir: ".isekai"}, json.RawMessage(`{"name":"ttl-craft"}`))
	if res.Err || !strings.Contains(res.Output, "keep the ttl at one hour") {
		t.Fatalf("skill{name} returns the body: %+v", res)
	}
}

// a report over the 600-byte envelope (and over @CAP) still keeps its draft: the harness files
// from the report it holds, never from the digest the parent reads.
func TestForgeDraftSurvivesTheEnvelope(t *testing.T) {
	w, a := forgeWorld(t)
	long := "@S DONE\n"
	for len(long) < 2600 {
		long += "@F src/auth/login.go:1 a long finding that fills the report past the envelope and the cap\n"
	}
	long += "@FORGE skill:auth-tour | Use when touring src/auth | read login.go first\n@U colony long\n@E 0"
	m := throne(t, a, "tour auth",
		dispatchCall("d1", "elf-core", "findings on auth"),
		mock.Text(long),
		mock.Text("done"))
	if seen := m.Requests[2].Messages[len(m.Requests[2].Messages)-1].ToolResults[0].Content; strings.Contains(seen, "@FORGE") {
		t.Fatalf("the test needs the draft out of the parent's view:\n%s", seen)
	}
	notes := forgeNotes(t, w)
	if len(notes) != 1 || notes[0].By != "elf-core" || !strings.HasPrefix(notes[0].Text, "skill:auth-tour") {
		t.Fatalf("the draft is filed although the parent saw only the envelope: %+v", notes)
	}
}

// a draft relayed up three hops is filed once, as the slime's: the first filer wins.
func TestForgeRelayFilesOnce(t *testing.T) {
	w, a := forgeWorld(t)
	line := "@FORGE tool:token-grep | find token ttl lines | grep -rn ttl {{dir}}"
	throne(t, a, "the ttl ladder",
		dispatchCall("d1", "elf-core", "draft on the ttl"),
		dispatchCall("d2", "orc-security", "verdict on the ttl"),
		dispatchCall("d3", "slime-auth", "findings on the ttl"),
		mock.Text("@S DONE\n@F src/auth/login.go:1 ttl\n"+line+"\n@U colony slime\n@E 0"),
		mock.Text("@S DONE\n@F relayed\n"+line+"\n@U colony orc\n@E 0"),
		mock.Text("@S DONE\n@F relayed\n"+line+"\n@U colony elf\n@E 0"),
		mock.Text("done\n"+line))
	notes := forgeNotes(t, w)
	if len(notes) != 1 || notes[0].By != "slime-auth" {
		t.Fatalf("one filing, by the slime: %+v", notes)
	}
}

// the throne's own final answer files its drafts (prose, not the wire).
func TestForgeThroneAnswerFiles(t *testing.T) {
	w, a := forgeWorld(t)
	throne(t, a, "say it", mock.Text("Here is the pattern.\n@FORGE skill:say-it | Use when asked to say it | say it once"))
	notes := forgeNotes(t, w)
	if len(notes) != 1 || notes[0].By != "rimuru" || notes[0].Text != "skill:say-it | Use when asked to say it | say it once" {
		t.Fatalf("the throne's draft: %+v", notes)
	}
}

func fileDraft(t *testing.T, a *App, by string, line string) {
	t.Helper()
	fs := wire.ForgesOf("@FORGE " + line)
	if len(fs) != 1 {
		t.Fatalf("draft line %q", line)
	}
	if holes := a.World.FileForges(by, "run-1", fs); len(holes) > 0 {
		t.Fatal(holes)
	}
}

// A tool mint is safe: shell quoting, the name checked first, the class from the classifier,
// interpreter-eval refused unless forced, outward needs --yes, the YAML printed, provenance kept.
func TestForgeToolMintIsSafe(t *testing.T) {
	w, a := forgeWorld(t)
	fileDraft(t, a, "slime-auth", "tool:todo-grep | find the TODOs | grep -rn 'TODO: fix' {{dir}}")
	fileDraft(t, a, "slime-auth", "tool:../evil | escape | ls")
	fileDraft(t, a, "slime-auth", "tool:runner | run anything | bash -c {{cmd}}")
	fileDraft(t, a, "slime-auth", "tool:pusher | push a branch | git push origin {{branch}}")
	fileDraft(t, a, "slime-auth", "tool:piped | count | grep x | wc -l")
	cfgBefore := w.read(".isekai/config.yaml")

	code, out, errs := desk(a, "mint", idOf(t, a, "todo-grep"))
	if code != 0 || !strings.Contains(out, "tools.custom.todo-grep: {") || !strings.Contains(out, `"TODO: fix"`) {
		t.Fatalf("the YAML is printed before it is written: %d\n%s\n%s", code, out, errs)
	}
	y := w.read(".isekai/config.yaml")
	if !strings.Contains(y, "TODO: fix") || !strings.Contains(y, "class: read") || !strings.Contains(y, "crafted:") || !strings.Contains(y, "draft:") || !strings.Contains(y, "dir:") {
		t.Fatalf("argv quoted as a shell would, class read from the classifier, provenance kept:\n%s", y)
	}
	if tl, ok := a.Cfg.Tools.Custom["todo-grep"]; !ok || len(tl.Run) != 4 || tl.Run[2] != "TODO: fix" || tl.Crafted == nil || tl.Crafted.By != "slime-auth" || tl.Crafted.Run != "run-1" {
		t.Fatalf("reloaded tools.custom.todo-grep: %+v", tl)
	}

	// the name is checked before any path is built
	code, _, errs = desk(a, "mint", idOf(t, a, "../evil"))
	if code == 0 || !strings.Contains(errs, "tools.custom's rule") {
		t.Fatalf("a bad name is refused: %d %s", code, errs)
	}
	// a shell pipeline is not an argv
	code, _, errs = desk(a, "mint", idOf(t, a, "piped"))
	if code == 0 || !strings.Contains(errs, "shell operator") {
		t.Fatalf("a pipeline is refused: %d %s", code, errs)
	}
	// interpreter-eval: refused, then forced
	runner := idOf(t, a, "runner")
	if code, _, errs = desk(a, "mint", runner); code == 0 || !strings.Contains(errs, "interpreter (bash -c)") {
		t.Fatalf("bash -c is refused unless forced: %d %s", code, errs)
	}
	if code, out, errs = desk(a, "mint", runner, "--force"); code != 0 {
		t.Fatalf("--force mints it: %d %s %s", code, out, errs)
	}
	// outward: headless needs --yes, and lands outward
	pusher := idOf(t, a, "pusher")
	if code, _, errs = desk(a, "mint", pusher); code == 0 || !strings.Contains(errs, "outward") || !strings.Contains(errs, "--yes") {
		t.Fatalf("an outward mint asks the human gate: %d %s", code, errs)
	}
	if strings.Contains(w.read(".isekai/config.yaml"), "pusher") {
		t.Fatal("a refused mint wrote nothing")
	}
	if code, out, errs = desk(a, "mint", pusher, "--yes"); code != 0 || a.Cfg.Tools.Custom["pusher"] == nil || a.Cfg.Tools.Custom["pusher"].Class != "outward" {
		t.Fatalf("--yes lands it with its class: %d %s %s %+v", code, out, errs, a.Cfg.Tools.Custom["pusher"])
	}
	if !strings.Contains(cfgBefore, "models:") {
		t.Fatal("fixture")
	}
}

// The desk: stable ids, reject with a why, a newer draft of a minted name shows as v2 and
// mints only with --version (the old file kept), never an overwrite.
func TestForgeDeskRejectAndVersions(t *testing.T) {
	w, a := forgeWorld(t)
	fileDraft(t, a, "slime-auth", "skill:ttl-craft | Use when the ttl moves | keep it at an hour")
	fileDraft(t, a, "slime-api", "skill:api-tour | Use when touring the api | read server.go")
	tour := idOf(t, a, "api-tour")
	if code, out, errs := desk(a, "reject", tour, "--why", "too narrow"); code != 0 || !strings.Contains(out, "rejected") {
		t.Fatalf("reject: %d %s %s", code, out, errs)
	}
	var out, errw bytes.Buffer
	a.forgeCmd(&out, &errw, "list")
	if strings.Contains(out.String(), "api-tour") || !strings.Contains(out.String(), "ttl-craft") || !strings.Contains(out.String(), "slime-auth · run-1") {
		t.Fatalf("the rejected draft leaves; the list shows by · run · age · kind:\n%s", out.String())
	}
	marks := forgeNotes(t, w)
	if last := marks[len(marks)-1]; last.Tag != "forge-rejected" || !strings.HasPrefix(last.Hash, tour) || !strings.Contains(last.Text, "too narrow") {
		t.Fatalf("the reject mark names the hash and the why: %+v", last)
	}
	if code, _, _ := desk(a, "mint", idOf(t, a, "ttl-craft")); code != 0 {
		t.Fatal("v1 mint")
	}
	v1 := w.read(".isekai/crafted-skills/ttl-craft/SKILL.md")
	// a newer draft of the minted name: v2 on the desk, refused with the diff, then --version
	fileDraft(t, a, "slime-auth", "skill:ttl-craft | Use when the ttl moves or is audited | keep it at an hour; log the change")
	_, list, _ := desk(a, "list")
	if !regexp.MustCompile(`ttl-craft\s+v2`).MatchString(list) {
		t.Fatalf("a newer draft of a minted name shows as v2:\n%s", list)
	}
	id2 := idOf(t, a, "ttl-craft")
	code, out2, errs := desk(a, "mint", id2)
	if code == 0 || !strings.Contains(errs, "nothing overwritten") || !strings.Contains(out2, "  - description:") || !strings.Contains(out2, "  + description:") {
		t.Fatalf("an existing crafted skill is refused with the diff: %d\n%s\n%s", code, out2, errs)
	}
	if w.read(".isekai/crafted-skills/ttl-craft/SKILL.md") != v1 {
		t.Fatal("the refusal overwrote")
	}
	if code, _, errs = desk(a, "mint", id2, "--version"); code != 0 {
		t.Fatalf("--version mints v2: %s", errs)
	}
	if s := w.read(".isekai/crafted-skills/ttl-craft/SKILL.md"); !strings.Contains(s, "version: 2") || !strings.Contains(s, "log the change") {
		t.Fatalf("v2:\n%s", s)
	}
	if w.read(".isekai/crafted-skills/ttl-craft/versions/v1.md") != v1 {
		t.Fatal("v1 is kept beside it")
	}
}

// a name another home already wears is refused: a crafted skill would shadow it.
func TestForgeSkillNeverShadows(t *testing.T) {
	w, a := forgeWorld(t)
	w.write(".claude/skills/ttl-craft/SKILL.md", "---\nname: ttl-craft\ndescription: the host's own\n---\nbody\n")
	fileDraft(t, a, "slime-auth", "skill:ttl-craft | Use when | how")
	fileDraft(t, a, "slime-auth", "skill:Bad_Name | Use when | how")
	code, _, errs := desk(a, "mint", idOf(t, a, "ttl-craft"))
	if code == 0 || !strings.Contains(errs, ".claude/skills/ttl-craft/SKILL.md") {
		t.Fatalf("shadowing refused: %d %s", code, errs)
	}
	if code, _, errs = desk(a, "mint", idOf(t, a, "Bad_Name")); code == 0 || !strings.Contains(errs, "[a-z0-9-]{1,64}") {
		t.Fatalf("Claude Code's name rule: %d %s", code, errs)
	}
	if _, err := os.Stat(filepath.Join(w.root, ".isekai/crafted-skills")); err == nil {
		if ents, _ := os.ReadDir(filepath.Join(w.root, ".isekai/crafted-skills")); len(ents) > 0 {
			t.Fatalf("nothing minted: %v", ents)
		}
	}
}

// drafts filed before the hash rode the note still list, and a legacy "kind:name" mark still
// closes them.
func TestForgeLegacyLedger(t *testing.T) {
	w, a := forgeWorld(t)
	w.write(".isekai/memory/shared/notes.jsonl", `{"at":"2026-09-28T21:00:00Z","by":"slime-auth","world":"x","tag":"forge","kind":"colony","text":"tool:ui-wrangler | rewrite bare /v1 strings | ./wrangle {{path}}"}
{"at":"2026-09-28T21:01:00Z","by":"slime-auth","world":"x","tag":"forge","kind":"colony","text":"skill:ttl-craft | one hour is the shelf life"}
{"at":"2026-09-28T21:02:00Z","by":"slime-auth","world":"x","tag":"forge-minted","kind":"colony","text":"skill:ttl-craft"}
`)
	_, out, _ := desk(a, "list")
	if !strings.Contains(out, "ui-wrangler") || strings.Contains(out, "ttl-craft") || !strings.Contains(out, "write") {
		t.Fatalf("legacy ledger:\n%s", out)
	}
}

// remember refuses the harness's own tags.
func TestRememberRefusesForgeTags(t *testing.T) {
	w, a := forgeWorld(t)
	for _, tag := range []string{"forge", "forge-minted", "skill", "Forge"} {
		res := a.rememberTool("slime-auth").Run(context.Background(), tool.Env{Root: a.Root, WorldDir: ".isekai"}, json.RawMessage(`{"text":"skill:x | y","tag":"`+tag+`"}`))
		if !res.Err || !strings.Contains(res.Output, "harness's own") {
			t.Fatalf("tag %s: %+v", tag, res)
		}
	}
	if strings.Contains(w.read(".isekai/memory/shared/notes.jsonl"), "skill:x") {
		t.Fatal("nothing filed")
	}
	if res := a.rememberTool("slime-auth").Run(context.Background(), tool.Env{Root: a.Root, WorldDir: ".isekai"}, json.RawMessage(`{"text":"plain","tag":"mine"}`)); res.Err {
		t.Fatalf("other tags stand: %+v", res)
	}
}

// One registry: status's minds and the skill tool's known list include crafted and base skills;
// a name with two homes is a hole naming both.
func TestSkillRegistryIncludesCrafted(t *testing.T) {
	files := isekaiCreatures()
	files[".isekai/crafted-skills/ttl-craft/SKILL.md"] = "---\nname: ttl-craft\ndescription: crafted\n---\nthe crafted body\n"
	files[".isekai/base-skills/auth-tour/SKILL.md"] = "---\nname: auth-tour\ndescription: base\n---\nbody\n"
	files[".claude/skills/ttl-craft/SKILL.md"] = "---\nname: ttl-craft\ndescription: host copy\n---\nthe host body\n"
	w := newTestWorld(t, "isekai", files)
	w.script(".isekai/tmp/s.json", text("nothing"))
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	a := w.open()
	var out bytes.Buffer
	a.cmdStatus(IO{Out: &out, Err: &out})
	s := out.String()
	n := len(a.Found.Skills)
	if !strings.Contains(s, " "+strconv.Itoa(n)+" minds") || n < 3 {
		t.Fatalf("status counts %d minds:\n%s", n, s)
	}
	if !strings.Contains(s, `@? skill "ttl-craft" has two homes`) || !strings.Contains(s, ".isekai/crafted-skills/ttl-craft/SKILL.md wins over") {
		t.Fatalf("two homes are a hole naming both:\n%s", s)
	}
	sk, ok := a.Found.Skill("ttl-craft")
	if !ok || sk.Source != "crafted" {
		t.Fatalf("crafted wins: %+v", sk)
	}
	res := a.skillTool("rimuru").Run(context.Background(), tool.Env{Root: a.Root, WorldDir: ".isekai"}, json.RawMessage(`{"name":"nope"}`))
	if !strings.Contains(res.Output, "ttl-craft") || !strings.Contains(res.Output, "auth-tour") {
		t.Fatalf("the skill tool's known list: %s", res.Output)
	}
}

// the CLI desk, headless: list, show, mint --yes.
func TestForgeCLI(t *testing.T) {
	w, a := forgeWorld(t)
	fileDraft(t, a, "slime-auth", "skill:ttl-craft | Use when the ttl moves | keep it at an hour")
	a.Close()
	run := func(args ...string) (int, string, string) {
		var out, errb bytes.Buffer
		code := Main("isekai", append([]string{"--root", w.root, "--quiet", "--no-board"}, args...), IO{In: strings.NewReader(""), Out: &out, Err: &errb, Env: func(k string) string {
			if k == "HOME" {
				return w.home
			}
			return ""
		}}, Version{Version: "test"})
		return code, out.String(), errb.String()
	}
	code, out, errs := run("forge", "list")
	m := deskID.FindStringSubmatch(out)
	if code != 0 || m == nil || m[2] != "ttl-craft" {
		t.Fatalf("forge list: %d\n%s\n%s", code, out, errs)
	}
	if code, out, errs = run("forge", "show", m[1]); code != 0 || !strings.Contains(out, "crafted-skills/ttl-craft/SKILL.md") {
		t.Fatalf("forge show: %d\n%s\n%s", code, out, errs)
	}
	if code, out, errs = run("forge", "mint", m[1], "--version"); code != 0 || !strings.Contains(out, "minted skill:ttl-craft v1") {
		t.Fatalf("forge mint (--version is the desk's flag, not the binary's): %d\n%s\n%s", code, out, errs)
	}
	if code, out, _ = run("--help"); !strings.Contains(out, "forge") {
		t.Fatalf("help lists forge: %s", out)
	}
}

// Under Claude Code the Stop hook files the turn's final answer's drafts: the transcript's last
// assistant text (after the last user entry), by the session — once, however often stop runs.
func TestClaudeStopFilesTheFinalAnswersDrafts(t *testing.T) {
	w := claudeWorld(t)
	tp := filepath.Join(w.home, "transcript.jsonl")
	w.write("~/transcript.jsonl", `{"type":"user","message":{"role":"user","content":"tour the auth zone"}}
{"type":"assistant","message":{"content":[{"type":"text","text":"@FORGE skill:early-idea | not final | x"},{"type":"tool_use","name":"Read"}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"ok"}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Toured.\n@FORGE skill:auth-tour | Use when touring src/auth | read login.go first"}]}}
`)
	w.hook("prompt", map[string]interface{}{"hook_event_name": "UserPromptSubmit", "prompt": "tour the auth zone"})
	for i := 0; i < 2; i++ {
		w.hook("stop", map[string]interface{}{"hook_event_name": "Stop", "transcript_path": tp})
	}
	notes := forgeNotes(t, w)
	if len(notes) != 1 || notes[0].By != "rimuru" || notes[0].Run != "cc-1" || !strings.HasPrefix(notes[0].Text, "skill:auth-tour") {
		t.Fatalf("the final answer's draft, once: %+v", notes)
	}
	// the host's own field wins when it sends one
	w.hook("stop", map[string]interface{}{"hook_event_name": "Stop", "last_assistant_message": "@FORGE tool:auth-grep | grep auth | grep -rn auth src"})
	if notes = forgeNotes(t, w); len(notes) != 2 || !strings.HasPrefix(notes[1].Text, "tool:auth-grep") {
		t.Fatalf("last_assistant_message: %+v", notes)
	}
}

// law.forge.demandAtSteps: 0 is reported as switched off, with where, like every law.
func TestForgeDemandOffIsReported(t *testing.T) {
	_, a := forgeWorld(t) // its config sets law.forge.demandAtSteps: 0
	var out bytes.Buffer
	a.cmdStatus(IO{Out: &out, Err: &out})
	if !strings.Contains(out.String(), "law.forge.demandAtSteps") || !strings.Contains(out.String(), "config.yaml") {
		t.Fatalf("status names the switched-off law and where:\n%s", out.String())
	}
	if a.Build().ForgeDemandAt != 0 {
		t.Fatal("the build carries the law's value")
	}
}
