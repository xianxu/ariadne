package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueref"
)

// queueHeadFact comes only from the raw immutable commit object, never a ref or
// replacement-sensitive history query.
type queueHeadFact struct {
	OID, Message string
	Bytes        int
	Parents      []string
}

type queueHeadKind uint8

const (
	queueHeadUnknown queueHeadKind = iota
	queueHeadDirect
	queueHeadNeutral
	queueHeadTrainCar
	queueHeadTrainIntegration
	// A malformed train cannot serve as the nontrain base of another train.
	queueHeadMalformedTrain
)

type queueHeadStep struct {
	Kind       queueHeadKind
	Issues     []int
	Train      int
	NeedParent string
	Reason     string
}

// queueHeadMembership is an immutable command-local value. Neutral descendants
// share Issues; any operation extending membership must allocate its own slice.
type queueHeadMembership struct {
	Known  bool
	Issues []int
	Train  int
	Reason string
}

// parseQueueHeadObject validates the complete header section before extracting
// parents and message. An incomplete object must not turn a missing parent into
// a seemingly definite attribution boundary.
func parseQueueHeadObject(oid string, raw []byte) (queueHeadFact, error) {
	bad := func(reason string) (queueHeadFact, error) {
		return queueHeadFact{}, fmt.Errorf("invalid queue head %s: %s", oid, reason)
	}
	if !landingOIDValid(oid) {
		return bad("invalid full object identity")
	}
	if bytes.IndexByte(raw, 0) >= 0 {
		return bad("NUL in commit object")
	}
	split := bytes.Index(raw, []byte("\n\n"))
	if split < 0 {
		return bad("incomplete commit headers")
	}
	fact := queueHeadFact{OID: oid, Message: string(raw[split+2:]), Bytes: len(raw)}
	seen := map[string]int{}
	previous := ""
	for _, line := range strings.Split(string(raw[:split]), "\n") {
		if strings.HasPrefix(line, " ") {
			if previous == "" || previous == "tree" || previous == "parent" || previous == "author" || previous == "committer" {
				return bad("invalid header continuation")
			}
			continue
		}
		key, value, ok := strings.Cut(line, " ")
		if !ok || key == "" || value == "" || strings.ContainsAny(key, "\t\r") || strings.ContainsRune(value, '\r') {
			return bad("malformed commit header")
		}
		for _, c := range key {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
				return bad("invalid header name")
			}
		}
		previous = key
		seen[key]++
		switch key {
		case "tree", "parent":
			if !landingOIDValid(value) || len(value) != len(oid) {
				return bad("invalid " + key + " identity")
			}
			if key == "parent" {
				fact.Parents = append(fact.Parents, value)
			}
		}
	}
	for _, key := range []string{"tree", "author", "committer"} {
		if seen[key] != 1 {
			return bad("missing or duplicate " + key + " header")
		}
	}
	return fact, nil
}

var queueHeadTrailerLineRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*: .+$`)

var queueHeadCarRE = regexp.MustCompile(`^Landing train ([1-9][0-9]*): #([1-9][0-9]{0,5}) \(PR #([1-9][0-9]*)\)$`)
var queueHeadTrainIntegrationRE = regexp.MustCompile(`^Merge ([0-9a-f]+) into landing train ([1-9][0-9]*) \(sdlc merge, #449\)$`)
var queueHeadIntegrationRE = regexp.MustCompile(`^Merge ([0-9a-f]+) into ([^\s]+) \(sdlc merge, #332\)$`)

// queueHeadTrailer accepts exactly one matching trailer in the final trailer
// paragraph; duplicated fields or body citations do not establish authority.
func queueHeadTrailer(message, key, want string) bool {
	message = strings.TrimRight(message, "\n")
	paragraph := strings.LastIndex(message, "\n\n")
	if paragraph < 0 {
		return false
	}
	count := 0
	for _, line := range strings.Split(message, "\n") {
		if field, _, ok := strings.Cut(line, ":"); ok && strings.EqualFold(field, key) {
			count++
		}
	}
	if count != 1 {
		return false
	}
	found := false
	for _, line := range strings.Split(message[paragraph+2:], "\n") {
		if !queueHeadTrailerLineRE.MatchString(line) {
			return false
		}
		if line == key+": "+want {
			found = true
		}
	}
	return found
}

func parseQueueHead(f queueHeadFact, selfRepo string) queueHeadStep {
	subject, _, _ := strings.Cut(f.Message, "\n")
	unknown := queueHeadStep{Reason: "commit has no definite issue lead or recognized neutral authority"}
	malformed := queueHeadStep{Kind: queueHeadMalformedTrain, Reason: "malformed landing train authority"}
	parentsValid := true
	for _, p := range f.Parents {
		if !landingOIDValid(p) || len(p) != len(f.OID) {
			parentsValid = false
		}
	}
	mergeValid := parentsValid && len(f.Parents) == 2
	firstParent := func() string {
		if len(f.Parents) > 0 {
			return f.Parents[0]
		}
		return ""
	}
	purportedTrain := strings.HasPrefix(subject, "Landing train") || strings.Contains(subject, " into landing train ") || strings.Contains(f.Message, "Landing-Train:")
	if purportedTrain {
		if !mergeValid {
			return malformed
		}
		if m := queueHeadCarRE.FindStringSubmatch(subject); m != nil {
			train, e1 := strconv.Atoi(m[1])
			issue, e2 := strconv.Atoi(m[2])
			_, e3 := strconv.Atoi(m[3])
			if e1 != nil || e2 != nil || e3 != nil || !queueHeadTrailer(f.Message, "Landing-Train", m[1]) || !queueHeadTrailer(f.Message, "Landing-PR", m[3]) || strings.Contains(f.Message, "Integrated-Main:") {
				return malformed
			}
			return queueHeadStep{Kind: queueHeadTrainCar, Issues: []int{issue}, Train: train, NeedParent: firstParent()}
		}
		if m := queueHeadTrainIntegrationRE.FindStringSubmatch(subject); m != nil {
			train, err := strconv.Atoi(m[2])
			if err != nil || m[1] != f.Parents[1] || !queueHeadTrailer(f.Message, "Integrated-Main", m[1]) || !queueHeadTrailer(f.Message, "Landing-Train", m[2]) || strings.Contains(f.Message, "Landing-PR:") {
				return malformed
			}
			return queueHeadStep{Kind: queueHeadTrainIntegration, Train: train, NeedParent: firstParent()}
		}
		return malformed
	}
	if len(issueref.Lead(subject)) > 0 {
		step := queueHeadStep{Kind: queueHeadDirect}
		for _, n := range issueref.LeadLocalNums(subject, selfRepo) {
			i, err := strconv.Atoi(n)
			if err != nil || i <= 0 {
				return unknown
			}
			step.Issues = append(step.Issues, i)
		}
		return step
	}
	if subject == defaultSyncSubject || subject == archiveCommitMessage {
		if parentsValid && len(f.Parents) == 1 {
			return queueHeadStep{Kind: queueHeadNeutral, NeedParent: firstParent()}
		}
		return unknown
	}
	if m := queueHeadIntegrationRE.FindStringSubmatch(subject); m != nil && mergeValid && m[1] == f.Parents[1] && queueHeadTrailer(f.Message, "Integrated-Main", m[1]) {
		return queueHeadStep{Kind: queueHeadNeutral, NeedParent: firstParent()}
	}
	return unknown
}

// queueHeadNeedsParentMembership distinguishes a required parent's immutable
// classification from its membership: a train's base must be read, but must not
// be walked or unioned into the train.
func queueHeadNeedsParentMembership(step, parent queueHeadStep) bool {
	if step.Kind == queueHeadNeutral {
		return true
	}
	return (step.Kind == queueHeadTrainCar || step.Kind == queueHeadTrainIntegration) &&
		(parent.Kind == queueHeadTrainCar || parent.Kind == queueHeadTrainIntegration) && step.Train == parent.Train
}

func selectQueueHeadMembership(step queueHeadStep, parentStep *queueHeadStep, parentMembership *queueHeadMembership) queueHeadMembership {
	unknown := func(reason string) queueHeadMembership { return queueHeadMembership{Reason: reason} }
	switch step.Kind {
	case queueHeadDirect:
		return queueHeadMembership{Known: true, Issues: append([]int(nil), step.Issues...)}
	case queueHeadNeutral:
		if parentStep == nil || parentStep.Kind == queueHeadUnknown || parentStep.Kind == queueHeadMalformedTrain || parentMembership == nil || !parentMembership.Known {
			return unknown("neutral head has unresolved first-parent membership")
		}
		return queueHeadMembership{Known: true, Issues: parentMembership.Issues, Train: parentMembership.Train}
	case queueHeadTrainCar, queueHeadTrainIntegration:
		if parentStep == nil {
			return unknown("train head has unread first parent")
		}
		if parentStep.Kind == queueHeadMalformedTrain {
			return unknown("train prefix has ambiguous first-parent authority")
		}
		// Only consecutive same-train composition belongs to this prefix. A main
		// tip may itself be neutral bookkeeping; its membership remains excluded.
		issues := append([]int(nil), step.Issues...)
		if queueHeadNeedsParentMembership(step, *parentStep) {
			if parentMembership == nil || !parentMembership.Known || parentMembership.Train != step.Train {
				return unknown("train prefix has unresolved membership")
			}
			issues = append(issues, parentMembership.Issues...)
		} else if step.Kind == queueHeadTrainIntegration {
			return unknown("train integration has no matching composed prefix")
		}
		if len(issues) == 0 || len(issues) > maxTrainCars {
			return unknown("train membership exceeds car bound")
		}
		seen := map[int]bool{}
		for _, issue := range issues {
			if issue <= 0 || seen[issue] {
				return unknown("duplicate or invalid train car membership")
			}
			seen[issue] = true
		}
		return queueHeadMembership{Known: true, Issues: issues, Train: step.Train}
	default:
		return unknown(step.Reason)
	}
}
