package rank

import "testing"

func TestCanonOneSpelling(t *testing.T) {
	for in, want := range map[string]string{
		"dark-elf": "dark-elf", "dark_elf": "dark-elf", "darkelf": "dark-elf", "DarkElf": "dark-elf", " Dark_Elf ": "dark-elf",
		"high_orc": "high-orc", "HighOrc": "high-orc", "highelf": "high-elf", "Slime": "slime", "rimuru": "rimuru",
		"night_watch": "night-watch", "NightWatch": "night-watch", "SLIME": "slime", "": "",
	} {
		if got := Canon(in); got != want {
			t.Errorf("Canon(%q) = %q, want %q", in, got, want)
		}
	}
	al := Aliases("dark_elf")
	if len(al) != 3 || al[0] != "dark-elf" || al[1] != "dark_elf" || al[2] != "darkelf" {
		t.Fatalf("aliases: %v", al)
	}
	if !Equal("DarkElf", "dark_elf") || Equal("elf", "dark-elf") {
		t.Fatal("Equal")
	}
}

func TestFamilyFromFlags(t *testing.T) {
	for want, f := range map[string]Flags{
		FamilyRoot:    {Root: true, Commands: true},
		FamilyAuthor:  {Authors: true},
		FamilyGate:    {HoldsGate: true, Commands: true},
		FamilyPlanner: {Commands: true, ReportsToRoot: true},
		FamilySteward: {Authors: true, HoldsGate: true, ReportsToRoot: true},
		FamilyAuditor: {ReadOnly: true, ReportsToRoot: true},
		FamilyGeneric: {},
	} {
		if got := FamilyOf(f); got != want {
			t.Errorf("%+v → %s, want %s", f, got, want)
		}
	}
	if !ReadOnly([]string{"read", "glob", "grep"}) || ReadOnly([]string{"read", "bash"}) || ReadOnly([]string{"*"}) || ReadOnly(nil) || ReadOnly([]string{"wr*"}) {
		t.Fatal("ReadOnly")
	}
}
