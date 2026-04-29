package renum

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Options struct {
	StartLine       int
	Increment       int
	FromLine        int
	StrictMSXParity bool
}

type Result struct {
	Text            string
	LineMap         map[int]int
	RenumberedLines int
	UndefinedRefs   []UndefinedReference
}

type Reference struct {
	SourceLine int
	Command    string
	Target     int
	Category   WarningCategory
	Severity   WarningSeverity
}

type DeleteResult struct {
	Text         string
	DeletedLines int
	BlockingRefs []Reference
}

type WarningCategory string

const (
	WarningCategoryFlow    WarningCategory = "flow"
	WarningCategoryListing WarningCategory = "listing"
)

type WarningSeverity string

const (
	WarningSeverityWarning WarningSeverity = "warning"
	WarningSeverityInfo    WarningSeverity = "info"
)

type WarningStats struct {
	Total   int
	Flow    int
	Listing int
	Warning int
	Info    int
}

type UndefinedReference struct {
	SourceLine int
	Command    string
	Target     int
	Category   WarningCategory
	Severity   WarningSeverity
}

type sourceLine struct {
	prefix    string
	oldNumber int
	body      string
	hasNumber bool
}

var numberedLineRE = regexp.MustCompile(`^(\s*)(\d+)(.*)$`)
var gotoOrGosubRE = regexp.MustCompile(`(?i)\b(GOTO|GOSUB)\s+([0-9][0-9\s,]*)`)
var branchDirectRE = regexp.MustCompile(`(?i)\b(THEN|ELSE)\s+([0-9]+)`)
var singleTargetCommandRE = regexp.MustCompile(`(?i)\b(RESTORE|RESUME|RUN|RETURN)\s+([0-9]+)([^0-9]|$)`)
var listingCommandRE = regexp.MustCompile(`(?i)\b(LLIST|LIST)\s+([^:]*)`)
var optionalTargetCommandRE = regexp.MustCompile(`(?i)\b(DELETE|EDIT)\s+([^:]*)`)
var numberRE = regexp.MustCompile(`\d+`)

func Renumber(text string, opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	if err := validateOptions(opts); err != nil {
		return Result{}, err
	}

	hasTrailingNewline := strings.HasSuffix(text, "\n")
	rawLines := strings.Split(text, "\n")
	lines := make([]sourceLine, len(rawLines))

	for i, raw := range rawLines {
		lines[i] = parseSourceLine(raw)
	}

	lineMap := map[int]int{}
	nextLine := opts.StartLine
	renumbered := 0
	for _, line := range lines {
		if !line.hasNumber {
			continue
		}
		newNum := line.oldNumber
		if line.oldNumber >= opts.FromLine {
			newNum = nextLine
			nextLine += opts.Increment
			if newNum != line.oldNumber {
				renumbered++
			}
		}
		lineMap[line.oldNumber] = newNum
	}

	rewritten := make([]string, len(lines))
	undefined := make([]UndefinedReference, 0, 8)
	for i, line := range lines {
		sourceLine := line.oldNumber
		if sourceLine == 0 {
			sourceLine = i + 1
		}
		if !line.hasNumber {
			rewritten[i] = line.prefix + rewriteReferences(line.body, lineMap, sourceLine, &undefined)
			continue
		}
		newNumber := lineMap[line.oldNumber]
		rewritten[i] = fmt.Sprintf("%s%d%s", line.prefix, newNumber, rewriteReferences(line.body, lineMap, sourceLine, &undefined))
	}

	out := strings.Join(rewritten, "\n")
	if hasTrailingNewline && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if opts.StrictMSXParity {
		if err := strictParityError(undefined); err != nil {
			return Result{}, err
		}
	}
	return Result{Text: out, LineMap: lineMap, RenumberedLines: renumbered, UndefinedRefs: undefined}, nil
}

func AnalyzeReferences(text string) []Reference {
	rawLines := strings.Split(text, "\n")
	refs := make([]Reference, 0, 16)
	for i, raw := range rawLines {
		line := parseSourceLine(raw)
		sourceLine := i + 1
		body := line.prefix
		if line.hasNumber {
			sourceLine = line.oldNumber
			body = line.body
		}
		collectReferences(body, sourceLine, &refs)
	}
	return refs
}

func DeleteRange(text string, fromLine, toLine int) (DeleteResult, error) {
	if fromLine <= 0 || toLine <= 0 {
		return DeleteResult{}, fmt.Errorf("delete range lines must be greater than zero")
	}
	if fromLine > toLine {
		fromLine, toLine = toLine, fromLine
	}

	hasTrailingNewline := strings.HasSuffix(text, "\n")
	rawLines := strings.Split(text, "\n")
	lines := make([]sourceLine, len(rawLines))
	deleteSet := make(map[int]struct{})
	deleted := 0
	for i, raw := range rawLines {
		lines[i] = parseSourceLine(raw)
		if lines[i].hasNumber && lines[i].oldNumber >= fromLine && lines[i].oldNumber <= toLine {
			deleteSet[lines[i].oldNumber] = struct{}{}
			deleted++
		}
	}
	if deleted == 0 {
		return DeleteResult{Text: text}, nil
	}

	allRefs := AnalyzeReferences(text)
	blocking := make([]Reference, 0, 8)
	for _, ref := range allRefs {
		if _, targetDeleted := deleteSet[ref.Target]; !targetDeleted {
			continue
		}
		if _, sourceDeleted := deleteSet[ref.SourceLine]; sourceDeleted {
			continue
		}
		blocking = append(blocking, ref)
	}
	if len(blocking) > 0 {
		return DeleteResult{Text: text, DeletedLines: deleted, BlockingRefs: blocking}, nil
	}

	kept := make([]string, 0, len(rawLines)-deleted)
	for i, raw := range rawLines {
		line := lines[i]
		if line.hasNumber && line.oldNumber >= fromLine && line.oldNumber <= toLine {
			continue
		}
		kept = append(kept, raw)
	}
	out := strings.Join(kept, "\n")
	if hasTrailingNewline && out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return DeleteResult{Text: out, DeletedLines: deleted}, nil
}

func normalizeOptions(opts Options) Options {
	if opts.StartLine <= 0 {
		opts.StartLine = 10
	}
	if opts.Increment <= 0 {
		opts.Increment = 10
	}
	if opts.FromLine < 0 {
		opts.FromLine = 0
	}
	return opts
}

func SummarizeWarnings(refs []UndefinedReference) WarningStats {
	stats := WarningStats{Total: len(refs)}
	for _, ref := range refs {
		ref = normalizeWarning(ref)
		switch ref.Category {
		case WarningCategoryListing:
			stats.Listing++
		default:
			stats.Flow++
		}
		switch ref.Severity {
		case WarningSeverityInfo:
			stats.Info++
		default:
			stats.Warning++
		}
	}
	return stats
}

func SummarizeReferences(refs []Reference) WarningStats {
	stats := WarningStats{Total: len(refs)}
	for _, ref := range refs {
		ref = normalizeReference(ref)
		switch ref.Category {
		case WarningCategoryListing:
			stats.Listing++
		default:
			stats.Flow++
		}
		switch ref.Severity {
		case WarningSeverityInfo:
			stats.Info++
		default:
			stats.Warning++
		}
	}
	return stats
}

func validateOptions(opts Options) error {
	if opts.StartLine <= 0 {
		return fmt.Errorf("start line must be greater than zero")
	}
	if opts.Increment <= 0 {
		return fmt.Errorf("increment must be greater than zero")
	}
	if opts.FromLine < 0 {
		return fmt.Errorf("renumber from line must be zero or greater")
	}
	return nil
}

func parseSourceLine(raw string) sourceLine {
	m := numberedLineRE.FindStringSubmatch(raw)
	if m == nil {
		return sourceLine{prefix: raw}
	}
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return sourceLine{prefix: raw}
	}
	return sourceLine{
		prefix:    m[1],
		oldNumber: n,
		body:      m[3],
		hasNumber: true,
	}
}

func rewriteReferences(body string, lineMap map[int]int, sourceLine int, undefined *[]UndefinedReference) string {
	code, comment := splitCodeAndComment(body)
	if code == "" {
		return body
	}
	segments := splitByQuotedStrings(code)
	for i := 0; i < len(segments); i += 2 {
		segments[i] = rewriteOutsideStrings(segments[i], lineMap, sourceLine, undefined)
	}
	return strings.Join(segments, "") + comment
}

func collectReferences(body string, sourceLine int, refs *[]Reference) {
	code, _ := splitCodeAndComment(body)
	if code == "" {
		return
	}
	segments := splitByQuotedStrings(code)
	for i := 0; i < len(segments); i += 2 {
		collectOutsideStringsReferences(segments[i], sourceLine, refs)
	}
}

func splitCodeAndComment(body string) (string, string) {
	inString := false
	for i := 0; i < len(body); i++ {
		ch := body[i]
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '\'' {
			return body[:i], body[i:]
		}
		if i+2 < len(body) && equalsFold3(body[i], body[i+1], body[i+2], 'R', 'E', 'M') {
			beforeOK := i == 0 || !isWordChar(body[i-1])
			afterIdx := i + 3
			afterOK := afterIdx >= len(body) || !isWordChar(body[afterIdx])
			if beforeOK && afterOK {
				return body[:i], body[i:]
			}
		}
	}
	return body, ""
}

func splitByQuotedStrings(code string) []string {
	if code == "" {
		return []string{""}
	}
	segments := make([]string, 0, 8)
	start := 0
	inString := false
	for i := 0; i < len(code); i++ {
		if code[i] != '"' {
			continue
		}
		if inString {
			segments = append(segments, code[start:i+1])
			start = i + 1
			inString = false
		} else {
			if start <= i {
				segments = append(segments, code[start:i])
			}
			start = i
			inString = true
		}
	}
	if start <= len(code) {
		segments = append(segments, code[start:])
	}
	if len(segments) == 0 {
		return []string{code}
	}
	if segments[0] != "" && strings.HasPrefix(segments[0], "\"") {
		segments = append([]string{""}, segments...)
	}
	return segments
}

func rewriteOutsideStrings(code string, lineMap map[int]int, sourceLine int, undefined *[]UndefinedReference) string {
	code = gotoOrGosubRE.ReplaceAllStringFunc(code, func(m string) string {
		parts := gotoOrGosubRE.FindStringSubmatch(m)
		if len(parts) < 3 {
			return m
		}
		repl := replaceNumberList(parts[2], parts[1], lineMap, sourceLine, undefined, true)
		return strings.Replace(m, parts[2], repl, 1)
	})
	code = branchDirectRE.ReplaceAllStringFunc(code, func(m string) string {
		parts := branchDirectRE.FindStringSubmatch(m)
		if len(parts) < 3 {
			return m
		}
		oldN, err := strconv.Atoi(parts[2])
		if err != nil {
			return m
		}
		newN, ok := mapReference(oldN, parts[1], lineMap, sourceLine, undefined, true)
		if !ok {
			return m
		}
		return strings.Replace(m, parts[2], strconv.Itoa(newN), 1)
	})
	code = rewriteListingCommands(code, lineMap, sourceLine, undefined)
	code = optionalTargetCommandRE.ReplaceAllStringFunc(code, func(m string) string {
		parts := optionalTargetCommandRE.FindStringSubmatch(m)
		if len(parts) < 3 {
			return m
		}
		rewritten := replaceNumberList(parts[2], parts[1], lineMap, sourceLine, undefined, true)
		return strings.Replace(m, parts[2], rewritten, 1)
	})
	code = singleTargetCommandRE.ReplaceAllStringFunc(code, func(m string) string {
		parts := singleTargetCommandRE.FindStringSubmatch(m)
		if len(parts) < 4 {
			return m
		}
		oldN, err := strconv.Atoi(parts[2])
		if err != nil {
			return m
		}
		newN, ok := mapReference(oldN, parts[1], lineMap, sourceLine, undefined, true)
		if !ok {
			return m
		}
		return strings.Replace(m, parts[2], strconv.Itoa(newN), 1)
	})
	return code
}

func collectOutsideStringsReferences(code string, sourceLine int, refs *[]Reference) {
	for _, parts := range gotoOrGosubRE.FindAllStringSubmatch(code, -1) {
		if len(parts) < 3 {
			continue
		}
		appendNumberListReferences(refs, sourceLine, parts[1], parts[2])
	}
	for _, parts := range branchDirectRE.FindAllStringSubmatch(code, -1) {
		if len(parts) < 3 {
			continue
		}
		appendSingleReference(refs, sourceLine, parts[1], parts[2])
	}
	appendListingCommandReferences(code, sourceLine, refs)
	for _, parts := range optionalTargetCommandRE.FindAllStringSubmatch(code, -1) {
		if len(parts) < 3 {
			continue
		}
		appendNumberListReferences(refs, sourceLine, parts[1], parts[2])
	}
	for _, parts := range singleTargetCommandRE.FindAllStringSubmatch(code, -1) {
		if len(parts) < 4 {
			continue
		}
		appendSingleReference(refs, sourceLine, parts[1], parts[2])
	}
}

func rewriteListingCommands(code string, lineMap map[int]int, sourceLine int, undefined *[]UndefinedReference) string {
	matches := listingCommandRE.FindAllStringSubmatchIndex(code, -1)
	if len(matches) == 0 {
		return code
	}

	var b strings.Builder
	last := 0
	for _, idx := range matches {
		if len(idx) < 6 {
			continue
		}
		fullEnd := idx[1]
		cmdStart, cmdEnd := idx[2], idx[3]
		argsStart, argsEnd := idx[4], idx[5]
		if !isListingCommandContext(code, cmdStart) {
			continue
		}
		b.WriteString(code[last:argsStart])
		command := code[cmdStart:cmdEnd]
		args := code[argsStart:argsEnd]
		rewritten := replaceNumberList(args, command, lineMap, sourceLine, undefined, true)
		b.WriteString(rewritten)
		last = fullEnd
	}
	if last == 0 {
		return code
	}
	b.WriteString(code[last:])
	return b.String()
}

func appendListingCommandReferences(code string, sourceLine int, refs *[]Reference) {
	for _, idx := range listingCommandRE.FindAllStringSubmatchIndex(code, -1) {
		if len(idx) < 6 {
			continue
		}
		cmdStart, cmdEnd := idx[2], idx[3]
		argsStart, argsEnd := idx[4], idx[5]
		if !isListingCommandContext(code, cmdStart) {
			continue
		}
		command := code[cmdStart:cmdEnd]
		args := code[argsStart:argsEnd]
		appendNumberListReferences(refs, sourceLine, command, args)
	}
}

func appendNumberListReferences(refs *[]Reference, sourceLine int, command, list string) {
	for _, n := range numberRE.FindAllString(list, -1) {
		appendSingleReference(refs, sourceLine, command, n)
	}
}

func appendSingleReference(refs *[]Reference, sourceLine int, command, rawTarget string) {
	target, err := strconv.Atoi(rawTarget)
	if err != nil || target == 0 {
		return
	}
	*refs = append(*refs, newReference(sourceLine, command, target))
}

func replaceNumberList(list, command string, lineMap map[int]int, sourceLine int, undefined *[]UndefinedReference, warn bool) string {
	return numberRE.ReplaceAllStringFunc(list, func(n string) string {
		oldN, err := strconv.Atoi(n)
		if err != nil {
			return n
		}
		newN, ok := mapReference(oldN, command, lineMap, sourceLine, undefined, warn)
		if !ok {
			return n
		}
		return strconv.Itoa(newN)
	})
}

func mapReference(target int, command string, lineMap map[int]int, sourceLine int, undefined *[]UndefinedReference, warn bool) (int, bool) {
	if target == 0 {
		return 0, false
	}
	newN, ok := lineMap[target]
	if ok {
		return newN, true
	}
	if warn && undefined != nil {
		*undefined = append(*undefined, newUndefinedReference(sourceLine, command, target))
	}
	return 0, false
}

func newUndefinedReference(sourceLine int, command string, target int) UndefinedReference {
	category, severity := classifyWarning(command)
	return UndefinedReference{
		SourceLine: sourceLine,
		Command:    strings.ToUpper(strings.TrimSpace(command)),
		Target:     target,
		Category:   category,
		Severity:   severity,
	}
}

func newReference(sourceLine int, command string, target int) Reference {
	category, severity := classifyWarning(command)
	return Reference{
		SourceLine: sourceLine,
		Command:    strings.ToUpper(strings.TrimSpace(command)),
		Target:     target,
		Category:   category,
		Severity:   severity,
	}
}

func normalizeWarning(ref UndefinedReference) UndefinedReference {
	if ref.Category != "" && ref.Severity != "" {
		return ref
	}
	category, severity := classifyWarning(ref.Command)
	if ref.Category == "" {
		ref.Category = category
	}
	if ref.Severity == "" {
		ref.Severity = severity
	}
	if ref.Command != "" {
		ref.Command = strings.ToUpper(strings.TrimSpace(ref.Command))
	}
	return ref
}

func normalizeReference(ref Reference) Reference {
	if ref.Category != "" && ref.Severity != "" {
		return ref
	}
	category, severity := classifyWarning(ref.Command)
	if ref.Category == "" {
		ref.Category = category
	}
	if ref.Severity == "" {
		ref.Severity = severity
	}
	if ref.Command != "" {
		ref.Command = strings.ToUpper(strings.TrimSpace(ref.Command))
	}
	return ref
}

func classifyWarning(command string) (WarningCategory, WarningSeverity) {
	switch strings.ToUpper(strings.TrimSpace(command)) {
	case "LIST", "LLIST":
		return WarningCategoryListing, WarningSeverityInfo
	default:
		return WarningCategoryFlow, WarningSeverityWarning
	}
}

func strictParityError(refs []UndefinedReference) error {
	flow := make([]UndefinedReference, 0, len(refs))
	for _, ref := range refs {
		normalized := normalizeWarning(ref)
		if normalized.Category == WarningCategoryFlow {
			flow = append(flow, normalized)
		}
	}
	if len(flow) == 0 {
		return nil
	}
	first := flow[0]
	if len(flow) == 1 {
		return fmt.Errorf("Undefined line %d in %d", first.Target, first.SourceLine)
	}
	return fmt.Errorf("Undefined line %d in %d (and %d more)", first.Target, first.SourceLine, len(flow)-1)
}

func isListingCommandContext(code string, cmdStart int) bool {
	i := cmdStart - 1
	for i >= 0 && (code[i] == ' ' || code[i] == '\t') {
		i--
	}
	if i < 0 || code[i] == ':' {
		return true
	}
	end := i + 1
	for i >= 0 && isWordChar(code[i]) {
		i--
	}
	if end <= i+1 {
		return false
	}
	prev := strings.ToUpper(code[i+1 : end])
	return prev == "THEN" || prev == "ELSE"
}

func equalsFold3(a, b, c, x, y, z byte) bool {
	return toUpperASCII(a) == x && toUpperASCII(b) == y && toUpperASCII(c) == z
}

func toUpperASCII(ch byte) byte {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

func isWordChar(ch byte) bool {
	if ch >= 'a' && ch <= 'z' {
		return true
	}
	if ch >= 'A' && ch <= 'Z' {
		return true
	}
	if ch >= '0' && ch <= '9' {
		return true
	}
	return ch == '_' || ch == '$'
}
