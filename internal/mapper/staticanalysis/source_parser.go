package staticanalysis

import (
	"fmt"
	"sort"
	"strings"
)

type jsTokenKind uint8

const (
	jsTokenIdentifier jsTokenKind = iota
	jsTokenKeyword
	jsTokenString
	jsTokenPunct
	jsTokenOperator
)

type jsToken struct {
	kind  jsTokenKind
	value string
}

var jsKeywords = map[string]struct{}{
	"function": {},
	"class":    {},
	"import":   {},
	"export":   {},
	"default":  {},
	"from":     {},
	"as":       {},
	"const":    {},
	"let":      {},
	"var":      {},
	"async":    {},
}

func parseIndexedBundleSourceAST(source string) (jsBundleExtracted, error) {
	if err := validateBalancedSyntax(source); err != nil {
		return jsBundleExtracted{}, err
	}

	tokens, err := tokenizeJSSource(source)
	if err != nil {
		return jsBundleExtracted{}, err
	}

	functions := make(map[string]struct{})
	classes := make(map[string]struct{})
	exports := make(map[string]struct{})
	importEdges := make(map[string]struct{})
	requireEdges := make(map[string]struct{})
	anchorByKey := make(map[string]jsSignalAnchor)

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		if tok.kind == jsTokenString {
			if anchorType, ok := classifySignalAnchor(tok.value); ok {
				key := anchorType + "\x00" + tok.value
				anchorByKey[key] = jsSignalAnchor{Type: anchorType, Value: tok.value}
			}
		}

		if tok.kind == jsTokenKeyword {
			switch tok.value {
			case "function":
				if name, ok := tokenValue(tokens, i+1, jsTokenIdentifier); ok {
					functions[name] = struct{}{}
				}
			case "class":
				if name, ok := tokenValue(tokens, i+1, jsTokenIdentifier); ok {
					classes[name] = struct{}{}
				}
			case "import":
				consumeImportTokens(tokens, i, importEdges)
			case "export":
				consumeExportTokens(tokens, i, functions, classes, exports)
			case "const", "let", "var":
				consumeVariableDeclarators(tokens, i, functions)
			}
		}

		if tok.kind == jsTokenIdentifier {
			switch tok.value {
			case "require":
				if dep, ok := consumeRequireCall(tokens, i); ok {
					requireEdges[dep] = struct{}{}
				}
			case "exports":
				if name, isFunc, ok := consumeExportsAssignment(tokens, i); ok {
					exports[name] = struct{}{}
					if isFunc {
						functions[name] = struct{}{}
					}
				}
			case "module":
				if consumesModuleExportsAssignment(tokens, i) {
					exports["module.exports"] = struct{}{}
				}
			default:
				if isAssignedFunction(tokens, i) {
					functions[tok.value] = struct{}{}
				}
			}
		}
	}

	anchors := make([]jsSignalAnchor, 0, len(anchorByKey))
	for _, anchor := range anchorByKey {
		anchors = append(anchors, anchor)
	}
	sort.Slice(anchors, func(i, j int) bool {
		if anchors[i].Type == anchors[j].Type {
			return anchors[i].Value < anchors[j].Value
		}
		return anchors[i].Type < anchors[j].Type
	})

	return jsBundleExtracted{
		Functions:    mapKeys(functions),
		Classes:      mapKeys(classes),
		Exports:      mapKeys(exports),
		ImportEdges:  mapKeys(importEdges),
		RequireEdges: mapKeys(requireEdges),
		Anchors:      anchors,
	}, nil
}

func tokenizeJSSource(source string) ([]jsToken, error) {
	tokens := make([]jsToken, 0, 1024)

	for i := 0; i < len(source); {
		ch := source[i]

		if isWhitespace(ch) {
			i++
			continue
		}

		if ch == '/' && i+1 < len(source) {
			next := source[i+1]
			if next == '/' {
				i += 2
				for i < len(source) && source[i] != '\n' {
					i++
				}
				continue
			}
			if next == '*' {
				i += 2
				closed := false
				for i+1 < len(source) {
					if source[i] == '*' && source[i+1] == '/' {
						i += 2
						closed = true
						break
					}
					i++
				}
				if !closed {
					return nil, fmt.Errorf("unterminated block comment")
				}
				continue
			}
		}

		if isIdentifierStart(ch) {
			start := i
			i++
			for i < len(source) && isIdentifierContinue(source[i]) {
				i++
			}
			value := source[start:i]
			if _, ok := jsKeywords[value]; ok {
				tokens = append(tokens, jsToken{kind: jsTokenKeyword, value: value})
			} else {
				tokens = append(tokens, jsToken{kind: jsTokenIdentifier, value: value})
			}
			continue
		}

		if ch == '\'' || ch == '"' {
			value, nextIndex, err := readQuotedString(source, i, ch)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, jsToken{kind: jsTokenString, value: value})
			i = nextIndex
			continue
		}

		if ch == '`' {
			value, nextIndex, err := readTemplateLiteral(source, i)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(value) != "" {
				tokens = append(tokens, jsToken{kind: jsTokenString, value: value})
			}
			i = nextIndex
			continue
		}

		if ch == '=' && i+1 < len(source) && source[i+1] == '>' {
			tokens = append(tokens, jsToken{kind: jsTokenOperator, value: "=>"})
			i += 2
			continue
		}

		if isPunct(ch) {
			tokens = append(tokens, jsToken{kind: jsTokenPunct, value: string(ch)})
			i++
			continue
		}

		if ch == '=' {
			tokens = append(tokens, jsToken{kind: jsTokenOperator, value: "="})
			i++
			continue
		}

		i++
	}

	return tokens, nil
}

func readQuotedString(source string, start int, quote byte) (string, int, error) {
	i := start + 1
	for i < len(source) {
		if source[i] == '\\' {
			i += 2
			continue
		}
		if source[i] == quote {
			raw := source[start+1 : i]
			return decodeQuotedContent(raw), i + 1, nil
		}
		i++
	}
	return "", 0, fmt.Errorf("unterminated string literal")
}

func readTemplateLiteral(source string, start int) (string, int, error) {
	i := start + 1
	var b strings.Builder

	for i < len(source) {
		ch := source[i]
		if ch == '\\' {
			if i+1 < len(source) {
				b.WriteByte(source[i+1])
				i += 2
				continue
			}
			return "", 0, fmt.Errorf("unterminated template literal")
		}
		if ch == '`' {
			return b.String(), i + 1, nil
		}
		if ch == '$' && i+1 < len(source) && source[i+1] == '{' {
			depth := 1
			i += 2
			for i < len(source) && depth > 0 {
				if source[i] == '\\' {
					i += 2
					continue
				}
				if source[i] == '{' {
					depth++
				} else if source[i] == '}' {
					depth--
				}
				i++
			}
			if depth != 0 {
				return "", 0, fmt.Errorf("unterminated template expression")
			}
			continue
		}
		b.WriteByte(ch)
		i++
	}

	return "", 0, fmt.Errorf("unterminated template literal")
}

func consumeImportTokens(tokens []jsToken, i int, importEdges map[string]struct{}) {
	if tok, ok := tokenAt(tokens, i+1); ok && tok.kind == jsTokenPunct && tok.value == "(" {
		if dep, ok := tokenValue(tokens, i+2, jsTokenString); ok {
			importEdges[dep] = struct{}{}
		}
		return
	}

	for j := i + 1; j < len(tokens); j++ {
		tok := tokens[j]
		if tok.kind == jsTokenPunct && (tok.value == ";" || tok.value == "}") {
			break
		}
		if tok.kind == jsTokenString {
			importEdges[tok.value] = struct{}{}
			break
		}
		if tok.kind == jsTokenKeyword && tok.value == "from" {
			if dep, ok := tokenValue(tokens, j+1, jsTokenString); ok {
				importEdges[dep] = struct{}{}
			}
			break
		}
	}
}

func consumeExportTokens(tokens []jsToken, i int, functions, classes, exports map[string]struct{}) {
	next, ok := tokenAt(tokens, i+1)
	if !ok {
		return
	}

	if next.kind == jsTokenKeyword && next.value == "default" {
		exports["default"] = struct{}{}
		return
	}

	if next.kind == jsTokenPunct && next.value == "{" {
		for j := i + 2; j < len(tokens); j++ {
			tok := tokens[j]
			if tok.kind == jsTokenPunct && tok.value == "}" {
				break
			}
			if tok.kind != jsTokenIdentifier {
				continue
			}
			nextTok, hasNext := tokenAt(tokens, j+1)
			if hasNext && nextTok.kind == jsTokenKeyword && nextTok.value == "as" {
				continue
			}
			exports[tok.value] = struct{}{}
		}
		return
	}

	if next.kind == jsTokenKeyword && next.value == "function" {
		if name, ok := tokenValue(tokens, i+2, jsTokenIdentifier); ok {
			functions[name] = struct{}{}
			exports[name] = struct{}{}
		}
		return
	}

	if next.kind == jsTokenKeyword && next.value == "class" {
		if name, ok := tokenValue(tokens, i+2, jsTokenIdentifier); ok {
			classes[name] = struct{}{}
			exports[name] = struct{}{}
		}
		return
	}

	if next.kind == jsTokenKeyword && (next.value == "const" || next.value == "let" || next.value == "var") {
		for _, decl := range scanDeclarators(tokens, i+2) {
			exports[decl.name] = struct{}{}
			if decl.functionLike {
				functions[decl.name] = struct{}{}
			}
		}
	}
}

func consumeVariableDeclarators(tokens []jsToken, i int, functions map[string]struct{}) {
	for _, decl := range scanDeclarators(tokens, i+1) {
		if decl.functionLike {
			functions[decl.name] = struct{}{}
		}
	}
}

type declarator struct {
	name         string
	functionLike bool
}

func scanDeclarators(tokens []jsToken, start int) []declarator {
	out := make([]declarator, 0)

	for i := start; i < len(tokens); {
		tok := tokens[i]
		if tok.kind == jsTokenPunct && (tok.value == ";" || tok.value == "}") {
			break
		}
		if tok.kind != jsTokenIdentifier {
			i++
			continue
		}

		decl := declarator{name: tok.value}
		if eq, ok := tokenAt(tokens, i+1); ok && eq.kind == jsTokenOperator && eq.value == "=" {
			decl.functionLike = rhsLooksFunctionLike(tokens, i+2)
		}
		out = append(out, decl)

		next := i + 1
		depthParen, depthBrace, depthBracket := 0, 0, 0
		for ; next < len(tokens); next++ {
			nt := tokens[next]
			if nt.kind == jsTokenPunct {
				switch nt.value {
				case "(":
					depthParen++
				case ")":
					if depthParen > 0 {
						depthParen--
					}
				case "{":
					depthBrace++
				case "}":
					if depthBrace == 0 && depthParen == 0 && depthBracket == 0 {
						break
					}
					if depthBrace > 0 {
						depthBrace--
					}
				case "[":
					depthBracket++
				case "]":
					if depthBracket > 0 {
						depthBracket--
					}
				case ",":
					if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
						next++
						goto advance
					}
				case ";":
					if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
						goto end
					}
				}
			}
		}
	advance:
		i = next
		continue
	}

end:
	return out
}

func consumeRequireCall(tokens []jsToken, i int) (string, bool) {
	open, ok := tokenAt(tokens, i+1)
	if !ok || open.kind != jsTokenPunct || open.value != "(" {
		return "", false
	}
	dep, ok := tokenValue(tokens, i+2, jsTokenString)
	if !ok {
		return "", false
	}
	close, ok := tokenAt(tokens, i+3)
	if !ok || close.kind != jsTokenPunct || close.value != ")" {
		return "", false
	}
	return dep, true
}

func consumeExportsAssignment(tokens []jsToken, i int) (string, bool, bool) {
	dot, ok := tokenAt(tokens, i+1)
	if !ok || dot.kind != jsTokenPunct || dot.value != "." {
		return "", false, false
	}
	name, ok := tokenValue(tokens, i+2, jsTokenIdentifier)
	if !ok {
		return "", false, false
	}
	eq, ok := tokenAt(tokens, i+3)
	if !ok || eq.kind != jsTokenOperator || eq.value != "=" {
		return "", false, false
	}
	return name, rhsLooksFunctionLike(tokens, i+4), true
}

func consumesModuleExportsAssignment(tokens []jsToken, i int) bool {
	dot, ok := tokenAt(tokens, i+1)
	if !ok || dot.kind != jsTokenPunct || dot.value != "." {
		return false
	}
	exportsTok, ok := tokenAt(tokens, i+2)
	if !ok || exportsTok.kind != jsTokenIdentifier || exportsTok.value != "exports" {
		return false
	}
	eq, ok := tokenAt(tokens, i+3)
	return ok && eq.kind == jsTokenOperator && eq.value == "="
}

func isAssignedFunction(tokens []jsToken, i int) bool {
	eq, ok := tokenAt(tokens, i+1)
	if !ok || eq.kind != jsTokenOperator || eq.value != "=" {
		return false
	}
	return rhsLooksFunctionLike(tokens, i+2)
}

func rhsLooksFunctionLike(tokens []jsToken, start int) bool {
	first, ok := tokenAt(tokens, start)
	if !ok {
		return false
	}
	if first.kind == jsTokenKeyword && first.value == "function" {
		return true
	}
	if first.kind == jsTokenKeyword && first.value == "async" {
		if second, ok := tokenAt(tokens, start+1); ok && second.kind == jsTokenKeyword && second.value == "function" {
			return true
		}
	}
	return hasArrowBeforeBoundary(tokens, start)
}

func hasArrowBeforeBoundary(tokens []jsToken, start int) bool {
	parenDepth := 0
	braceDepth := 0
	bracketDepth := 0

	for j := start; j < len(tokens); j++ {
		tok := tokens[j]
		if tok.kind == jsTokenOperator && tok.value == "=>" {
			return true
		}
		if tok.kind != jsTokenPunct {
			continue
		}
		switch tok.value {
		case "(":
			parenDepth++
		case ")":
			if parenDepth > 0 {
				parenDepth--
			}
		case "{":
			braceDepth++
		case "}":
			if braceDepth == 0 && parenDepth == 0 && bracketDepth == 0 {
				return false
			}
			if braceDepth > 0 {
				braceDepth--
			}
		case "[":
			bracketDepth++
		case "]":
			if bracketDepth > 0 {
				bracketDepth--
			}
		case ";", ",":
			if parenDepth == 0 && braceDepth == 0 && bracketDepth == 0 {
				return false
			}
		}
	}

	return false
}

func tokenAt(tokens []jsToken, i int) (jsToken, bool) {
	if i < 0 || i >= len(tokens) {
		return jsToken{}, false
	}
	return tokens[i], true
}

func tokenValue(tokens []jsToken, i int, kind jsTokenKind) (string, bool) {
	tok, ok := tokenAt(tokens, i)
	if !ok || tok.kind != kind {
		return "", false
	}
	return tok.value, true
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}

func isIdentifierStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}

func isIdentifierContinue(ch byte) bool {
	return isIdentifierStart(ch) || (ch >= '0' && ch <= '9')
}

func isPunct(ch byte) bool {
	switch ch {
	case '{', '}', '(', ')', '[', ']', ';', ',', '.':
		return true
	default:
		return false
	}
}
