package world

import (
	"os"
	"path/filepath"
)

// Genesis is how the tree grows: a zone-parent writes its own child's README, and the child is
// born. Right-author would read that write as an author on another creature's seat (the child
// is on the roster the moment its doc parses), and the throne's founding of an elf as an
// unpaired re-authoring of a drifted doc. genesisExempt is the one exemption, and it holds only
// while all four conditions do:
//
//  1. the path is the child's own README, on its seat;
//  2. the child names the writer as its one hop up (the throne: a child of rimuru);
//  3. the doc is new — absent from the last commit (genesisUntracked): one commit window, after
//     which the doc is the child's own and re-authoring it is the ordinary arms' business;
//  4. every claim the child makes stands on the live tree, on soil of the writer's own line
//     (the writer and those above it) or open ground, never over a third body's
//     (mintClaimsLiveTree) — a parent hands down what its line holds.
func (w *World) genesisExempt(as, rel string, zone Zone, before map[string][]string) bool {
	child := zone.Creature
	if zone.Soil != SoilSeat || child == nil || child.Doc == "" || normRel(rel) != normRel(child.Doc) || child.Name == as {
		return false
	}
	parent := or(as, Rimuru)
	if w.RankOf(parent).Name == Rimuru {
		parent = Rimuru
	}
	if child.Parent != parent {
		return false
	}
	return w.genesisUntracked(child.Doc, child.Name, before) && w.mintClaimsLiveTree(child, parent)
}

// genesisUntracked: the doc is absent from HEAD (staged or not, it was never committed). With
// no git to ask, the child must be absent from the roster the turn opened with.
func (w *World) genesisUntracked(doc, name string, before map[string][]string) bool {
	if _, err := w.gitOut("rev-parse --is-inside-work-tree"); err != nil {
		_, known := before[name]
		return !known
	}
	_, err := w.gitOut("cat-file -e HEAD:./" + filepath.ToSlash(filepath.Clean(doc)))
	return err != nil
}

// mintClaimsLiveTree: each of the child's claims exists on disk and neither covers nor sits under
// a claim of any living body outside the child and its parent's line.
func (w *World) mintClaimsLiveTree(child *Creature, parent string) bool {
	line := map[string]bool{child.Name: true, parent: true}
	for _, up := range w.lineOf(parent) {
		line[up] = true
	}
	for _, t := range child.Territory {
		if _, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(t))); err != nil {
			return false
		}
		for i := range w.Creatures {
			c := &w.Creatures[i]
			if line[c.Name] {
				continue
			}
			for _, p := range c.Territory {
				if under(normRel(t), p) || under(normRel(p), t) {
					return false
				}
			}
		}
	}
	return true
}
