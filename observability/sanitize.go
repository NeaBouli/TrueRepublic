package observability

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxCollectionItems = 64
	maxSanitizeDepth   = 12
	maxTextBytes       = 8 * 1024
)

var (
	privatePEMPattern = regexp.MustCompile(
		`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY(?: BLOCK)?-----.*?-----END [A-Z0-9 ]*PRIVATE KEY(?: BLOCK)?-----`,
	)
	authorizationPattern = regexp.MustCompile(
		`(?i)\bauthorization[ \t]*[:=][ \t]*(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\r\n,;]+)`,
	)
	urlCredentialPattern = regexp.MustCompile(
		`(?i)\b(https?://)[^/\s:@]+:[^@\s/]+@`,
	)
	mnemonicPattern = regexp.MustCompile(
		`(?im)\b(mnemonic|seed[ _-]*phrase)([ \t]*[:=][ \t]*)[^\r\n]*`,
	)
	labeledValuePattern = regexp.MustCompile(
		`(?i)\b(password|passwd|passphrase|secret(?:[ _-]*key)?|credential|private[ _-]*key|privkey|validator[ _-]*key|signing[ _-]*key|signer[ _-]*state|keyring|keystore|raw[ _-]*(?:tx|transaction)|tx[ _-]*bytes|txs?|transactions?|proof|signature|cookie|session|token|api[ _-]*(?:key|token)|auth[ _-]*token|access[ _-]*token|refresh[ _-]*token|id[ _-]*token|jwt)([ \t]*[:=][ \t]*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\r\n,;]+)`,
	)
	schemePattern      = regexp.MustCompile(`(?i)\b(bearer|basic)([ \t]+)[A-Za-z0-9._~+/=-]{8,}`)
	transactionPattern = regexp.MustCompile(
		`(?i)\b(raw[ _-]*(?:tx|transaction)|tx[ _-]*bytes|transaction[ _-]*(?:data|payload|bytes))([ \t]+)[A-Za-z0-9+/=._~-]{16,}`,
	)
)

var sensitiveKeyFragments = []string{
	"password",
	"passwd",
	"passphrase",
	"secret",
	"credential",
	"authentication",
	"authorization",
	"mnemonic",
	"seedphrase",
	"privatekey",
	"privkey",
	"validatorkey",
	"signingkey",
	"signerstate",
	"keyring",
	"keystore",
	"rawtransaction",
	"rawtx",
	"txbytes",
	"transactionbytes",
	"transactiondata",
	"transactionpayload",
	"proof",
	"signature",
	"cookie",
	"session",
	"authtoken",
	"accesstoken",
	"refreshtoken",
	"idtoken",
	"bearertoken",
	"apitoken",
	"apikey",
	"jwt",
}

var sensitiveExactKeys = map[string]struct{}{
	"auth":         {},
	"bearer":       {},
	"body":         {},
	"data":         {},
	"payload":      {},
	"private":      {},
	"pwd":          {},
	"seed":         {},
	"sig":          {},
	"token":        {},
	"transaction":  {},
	"transactions": {},
	"tx":           {},
	"txs":          {},
}

// Sanitize removes common labeled credentials, private material, proofs,
// signatures, and transaction payloads from free-form text. It is defensive
// minimization, not a general data-loss-prevention engine.
func Sanitize(input string) string {
	output := privatePEMPattern.ReplaceAllString(input, Redacted)
	output = authorizationPattern.ReplaceAllString(output, "authorization="+Redacted)
	output = urlCredentialPattern.ReplaceAllString(output, "${1}"+Redacted+"@")
	output = mnemonicPattern.ReplaceAllString(output, "${1}${2}"+Redacted)
	output = labeledValuePattern.ReplaceAllString(output, "${1}${2}"+Redacted)
	output = schemePattern.ReplaceAllString(output, "${1}${2}"+Redacted)
	output = transactionPattern.ReplaceAllString(output, "${1}${2}"+Redacted)
	if len(output) <= maxTextBytes {
		return output
	}
	end := maxTextBytes
	for end > 0 && !utf8.RuneStart(output[end]) {
		end--
	}
	return output[:end] + Redacted
}

func sanitizeKeyValues(input []any) []any {
	if len(input) == 0 {
		return nil
	}
	output := make([]any, 0, len(input)+(len(input)%2))
	for index := 0; index < len(input); index += 2 {
		rawKey := safeKey(input[index])
		key := Sanitize(rawKey)
		if key == "" {
			key = "invalid_field"
		}
		output = append(output, key)
		if index+1 >= len(input) {
			output = append(output, Redacted)
			continue
		}
		if isSensitiveKey(rawKey) {
			output = append(output, Redacted)
			continue
		}
		output = append(output, sanitizeValue(input[index+1], rawKey))
	}
	return output
}

func safeKey(value any) (output string) {
	defer func() {
		if recover() != nil {
			output = "invalid_field"
		}
	}()
	if key, ok := value.(string); ok {
		return key
	}
	if err, ok := value.(error); ok {
		return safeError(err)
	}
	if stringer, ok := value.(fmt.Stringer); ok {
		return safeString(stringer)
	}
	return fmt.Sprint(value)
}

func normalizeKey(key string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, key)
}

func isSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	if isExplicitPublicField(normalized) {
		return false
	}
	if _, sensitive := sensitiveExactKeys[normalized]; sensitive {
		return true
	}
	for _, fragment := range sensitiveKeyFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func isExplicitPublicField(normalized string) bool {
	switch normalized {
	case "accountpubkey",
		"apphash",
		"blockhash",
		"commithash",
		"consensushash",
		"consensuspubkey",
		"evidencehash",
		"nodeid",
		"nodepubkey",
		"operatorpubkey",
		"peerid",
		"peerpubkey",
		"proposalhash",
		"publichash",
		"publickey",
		"pubkey",
		"transactionhash",
		"txhash",
		"validatorpubkey":
		return true
	default:
		return false
	}
}

func isExplicitPublicBinaryKey(key string) bool {
	return isExplicitPublicField(normalizeKey(key))
}

func sanitizeValue(value any, key string) (output any) {
	defer func() {
		if recover() != nil {
			output = Redacted
		}
	}()
	return walkValue(reflect.ValueOf(value), key, 0, make(map[visit]struct{}))
}

type visit struct {
	kind reflect.Kind
	typ  reflect.Type
	ptr  uintptr
}

func walkValue(value reflect.Value, key string, depth int, seen map[visit]struct{}) any {
	if !value.IsValid() {
		return nil
	}
	if depth > maxSanitizeDepth {
		return Redacted
	}
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return walkValue(value.Elem(), key, depth+1, seen)
	case reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		current := visit{kind: value.Kind(), typ: value.Type(), ptr: value.Pointer()}
		if _, exists := seen[current]; exists {
			return Redacted
		}
		seen[current] = struct{}{}
		defer delete(seen, current)
		return walkValue(value.Elem(), key, depth+1, seen)
	}

	if value.CanInterface() {
		switch typed := value.Interface().(type) {
		case string:
			return Sanitize(typed)
		case error:
			return Sanitize(safeError(typed))
		case fmt.Stringer:
			return Sanitize(safeString(typed))
		case json.Number:
			return typed
		}
	}

	switch value.Kind() {
	case reflect.Slice:
		if value.IsNil() {
			return nil
		}
		if value.Type().Elem().Kind() == reflect.Uint8 {
			if isExplicitPublicBinaryKey(key) {
				copyOfBytes := make([]byte, value.Len())
				reflect.Copy(reflect.ValueOf(copyOfBytes), value)
				return copyOfBytes
			}
			return Redacted
		}
		current := visit{kind: value.Kind(), typ: value.Type(), ptr: value.Pointer()}
		if current.ptr != 0 {
			if _, exists := seen[current]; exists {
				return Redacted
			}
			seen[current] = struct{}{}
			defer delete(seen, current)
		}
		length := min(value.Len(), maxCollectionItems)
		result := make([]any, 0, length+1)
		for index := 0; index < length; index++ {
			result = append(result, walkValue(value.Index(index), "", depth+1, seen))
		}
		if value.Len() > length {
			result = append(result, Redacted)
		}
		return result
	case reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			if !isExplicitPublicBinaryKey(key) {
				return Redacted
			}
			result := make([]byte, value.Len())
			for index := range result {
				result[index] = byte(value.Index(index).Uint())
			}
			return result
		}
		length := min(value.Len(), maxCollectionItems)
		result := make([]any, 0, length+1)
		for index := 0; index < length; index++ {
			result = append(result, walkValue(value.Index(index), "", depth+1, seen))
		}
		if value.Len() > length {
			result = append(result, Redacted)
		}
		return result
	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		current := visit{kind: value.Kind(), typ: value.Type(), ptr: uintptr(value.UnsafePointer())}
		if _, exists := seen[current]; exists {
			return Redacted
		}
		seen[current] = struct{}{}
		defer delete(seen, current)
		result := make(map[string]any, min(value.Len(), maxCollectionItems)+1)
		iterator := value.MapRange()
		for count := 0; count < maxCollectionItems && iterator.Next(); count++ {
			rawKey := safeKeyValue(iterator.Key())
			outputKey := Sanitize(rawKey)
			if outputKey == "" {
				outputKey = "invalid_field"
			}
			if isSensitiveKey(rawKey) {
				result[outputKey] = Redacted
			} else {
				result[outputKey] = walkValue(iterator.Value(), rawKey, depth+1, seen)
			}
		}
		if value.Len() > maxCollectionItems {
			result["truncated"] = Redacted
		}
		return result
	case reflect.Struct:
		result := make(map[string]any, value.NumField())
		valueType := value.Type()
		for index := 0; index < value.NumField(); index++ {
			fieldType := valueType.Field(index)
			if !fieldType.IsExported() {
				continue
			}
			fieldName := fieldType.Name
			if jsonName := strings.Split(fieldType.Tag.Get("json"), ",")[0]; jsonName != "" && jsonName != "-" {
				fieldName = jsonName
			}
			if fieldName == "-" {
				continue
			}
			if isSensitiveKey(fieldName) {
				result[fieldName] = Redacted
			} else {
				result[fieldName] = walkValue(value.Field(index), fieldName, depth+1, seen)
			}
		}
		return result
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.String:
		return Sanitize(value.String())
	default:
		return Redacted
	}
}

func safeKeyValue(value reflect.Value) (output string) {
	defer func() {
		if recover() != nil {
			output = "invalid_field"
		}
	}()
	if value.Kind() == reflect.String {
		return value.String()
	}
	if value.CanInterface() {
		return safeKey(value.Interface())
	}
	return "invalid_field"
}

func safeError(value error) (output string) {
	defer func() {
		if recover() != nil {
			output = Redacted
		}
	}()
	return value.Error()
}

func safeString(value fmt.Stringer) (output string) {
	defer func() {
		if recover() != nil {
			output = Redacted
		}
	}()
	return value.String()
}
