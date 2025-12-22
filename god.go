package god

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/*
Package god implements encoding and decoding of GOD (Grounded Object Data),
a compact, human-readable data serialization format designed as a JSON alternative.

GOD emphasizes "grounded" data - using zero values instead of null/undefined,
ensuring all data is type-safe and grounded in concrete values.

GOD Features:
- Compact syntax: 20-50% smaller than JSON for structured data
- Native tabular format: (header:rows) for efficient array-of-struct encoding
- Grounded values: No null/nil - uses type-specific zero values
- Struct reflection: Automatic encoding/decoding with `god:"fieldname"` tags
- Optional semicolons: Whitespace-insignificant syntax
- Human-readable: Clean, readable format
- Type safety: Zero values prevent null pointer errors

Example usage:

	type Person struct {
		Name string `god:"name"`
		Age  int    `god:"age"`
	}

	person := Person{Name: "John", Age: 30}
	encoded, _ := god.Marshal(person)
	// Output: {name="John";age=30}

	var decoded Person
	god.Unmarshal(encoded, &decoded)

For tabular data:

	people := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
	encoded, _ := god.Marshal(people)
	// Output: {(name,age:"Alice",30;"Bob",25;)}

Zero values (grounded data):
	int/float: 0
	string: ""
	bool: false
	array: []
	object: {}
*/

// Table represents the key = (header:rows;...) syntax.
type Table struct {
	Header []string
	Rows   [][]string
}

// ===================== ENCODING =====================

// Marshal encodes any Go value into GOD format (compact, no extra whitespace).
// Rule 2: Root must always be an object. Non-object types are wrapped with a default key.
func Marshal(v interface{}) ([]byte, error) {
	return marshalWithCompact(v, true)
}

// MarshalBeautify encodes any Go value into formatted GOD (readable with indentation).
// Rule 2: Root must always be an object. Non-object types are wrapped with a default key.
func MarshalBeautify(v interface{}) ([]byte, error) {
	return marshalWithCompact(v, false)
}

func marshalWithCompact(v interface{}, compact bool) ([]byte, error) {
	var b strings.Builder
	rv := reflect.ValueOf(v)
	
	// Handle pointers
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	
	// Rule 2: Root must always be an object {}
	// Rule 5: Root can contain either:
	//   - A single raw value: {"string"}, {[...]}, {(table)}, etc.
	//   - Key-value pairs: {key=value;key2=value2}
	//   - But NOT both mixed together
	
	// If it's already a map or struct, encode normally (key-value pairs)
	if rv.Kind() == reflect.Map || rv.Kind() == reflect.Struct {
		if err := encodeValue(&b, rv, 0, compact); err != nil {
			return nil, err
		}
		return []byte(b.String()), nil
	}
	
	// Otherwise, wrap as single raw value in {}
	b.WriteByte('{')
	if !compact {
		b.WriteByte('\n')
		b.WriteString("  ")
	}
	
	if err := encodeValue(&b, rv, 1, compact); err != nil {
		return nil, err
	}
	
	if !compact {
		b.WriteByte('\n')
	}
	b.WriteByte('}')
	
	return []byte(b.String()), nil
}


func encodeValue(b *strings.Builder, v reflect.Value, level int, compact bool) error {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return encodeStruct(b, v, level, compact)
	case reflect.Map:
		return encodeMap(b, v, level, compact)
	case reflect.Slice, reflect.Array:
		return encodeSlice(b, v, level, compact)
	case reflect.String:
		return encodeString(b, v.String(), compact)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(fmt.Sprintf("%d", v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(fmt.Sprintf("%d", v.Uint()))
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if float64(int64(f)) == f {
			b.WriteString(fmt.Sprintf("%d", int64(f)))
		} else {
			b.WriteString(fmt.Sprintf("%v", f))
		}
	case reflect.Bool:
		if v.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return encodeValue(b, v.Elem(), level, compact)
	default:
		return fmt.Errorf("unsupported type: %v", v.Kind())
	}
	return nil
}

func encodeStruct(b *strings.Builder, v reflect.Value, level int, compact bool) error {
	t := v.Type()
	
	b.WriteByte('{')
	if !compact {
		b.WriteByte('\n')
	}
	
	first := true
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get field name from tag or use field name
		fieldName := field.Tag.Get("god")
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}
		
		if !first && compact {
			b.WriteByte(';')
		}
		first = false
		
		if !compact {
			b.WriteString(indent(level))
		}
		
		b.WriteString(fieldName)
		b.WriteByte('=')
		
		// Handle nil/zero values
		if isZeroValue(fieldValue) {
			if !compact {
				b.WriteString(";\n")
			}
			continue
		}
		
		if err := encodeValue(b, fieldValue, level+1, compact); err != nil {
			return err
		}
		
		if !compact {
			b.WriteString(";\n")
		}
	}
	
	if !compact {
		b.WriteString(indent(level - 1))
	}
	b.WriteByte('}')
	return nil
}

func encodeMap(b *strings.Builder, v reflect.Value, level int, compact bool) error {
	b.WriteByte('{')
	if !compact {
		b.WriteByte('\n')
	}
	
	first := true
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		
		if !first && compact {
			b.WriteByte(';')
		}
		first = false
		
		if !compact {
			b.WriteString(indent(level))
		}
		
		// Key must be string
		b.WriteString(fmt.Sprintf("%v", key.Interface()))
		b.WriteByte('=')
		
		if isZeroValue(val) || (val.Kind() == reflect.Interface && val.IsNil()) {
			if !compact {
				b.WriteString(";\n")
			}
			continue
		}
		
		if err := encodeValue(b, val, level+1, compact); err != nil {
			return err
		}
		
		if !compact {
			b.WriteString(";\n")
		}
	}
	
	if !compact {
		b.WriteString(indent(level - 1))
	}
	b.WriteByte('}')
	return nil
}

func encodeSlice(b *strings.Builder, v reflect.Value, level int, compact bool) error {
	if v.Len() == 0 {
		b.WriteString("[]")
		return nil
	}
	
	// Check if slice of structs -> use table format
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Struct {
		return encodeStructSliceAsTable(b, v, compact)
	}
	
	// Regular list
	b.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := encodeValue(b, v.Index(i), level, compact); err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

func encodeStructSliceAsTable(b *strings.Builder, v reflect.Value, compact bool) error {
	if v.Len() == 0 {
		b.WriteString("()")
		return nil
	}
	
	elemType := v.Type().Elem()
	
	// Build header from struct fields
	var headers []string
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName := field.Tag.Get("god")
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}
		headers = append(headers, fieldName)
	}
	
	b.WriteByte('(')
	
	// Write header
	for i, h := range headers {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(h)
	}
	b.WriteByte(':')
	
	// Write rows
	for i := 0; i < v.Len(); i++ {
		structVal := v.Index(i)
		for j := 0; j < structVal.NumField(); j++ {
			field := elemType.Field(j)
			if !field.IsExported() {
				continue
			}
			
			if j > 0 {
				b.WriteByte(',')
			}
			
			fieldVal := structVal.Field(j)
			if err := encodeTableCell(b, fieldVal); err != nil {
				return err
			}
		}
		b.WriteByte(';')
	}
	
	b.WriteByte(')')
	return nil
}

func encodeTableCell(b *strings.Builder, v reflect.Value) error {
	if !v.IsValid() {
		b.WriteString("\\0")
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			b.WriteString("\"\"") // Rule 18: string = ""
			return nil
		}
		b.WriteString(strconv.Quote(s))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(fmt.Sprintf("%d", v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(fmt.Sprintf("%d", v.Uint()))
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if float64(int64(f)) == f {
			b.WriteString(fmt.Sprintf("%d", int64(f)))
		} else {
			b.WriteString(fmt.Sprintf("%v", f))
		}
	case reflect.Bool:
		if v.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false") // Rule 18: bool = false
		}
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			b.WriteString("\\0") // Rule 18: if unsure/empty then \0
			return nil
		}
		return encodeTableCell(b, v.Elem())
	default:
		if isZeroValue(v) {
			b.WriteString("\\0")
			return nil
		}
		b.WriteString(strconv.Quote(fmt.Sprintf("%v", v.Interface())))
	}
	return nil
}

func encodeString(b *strings.Builder, s string, compact bool) error {
	if strings.Contains(s, "\n") {
		b.WriteString(`"""`)
		b.WriteString(s)
		b.WriteString(`"""`)
	} else {
		b.WriteString(strconv.Quote(s))
	}
	return nil
}

func indent(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat("  ", level)
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// ===================== DECODING =====================

// Unmarshal parses GOD data into a Go value.
// v must be a pointer to the target type.
// Special case: {(table...)} decodes directly to a slice if target is a slice.
func Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("unmarshal target must be a non-nil pointer")
	}
	
	p := &parser{src: data, pos: 0}
	p.skipSpaces()
	
	// Special case: Check if it's a bare table in root object {(table...)}
	// This should decode directly to a slice
	if rv.Elem().Kind() == reflect.Slice && p.peek() == '{' {
		// Peek ahead to see if it's {(
		savedPos := p.pos
		p.next() // consume '{'
		p.skipSpaces()
		if p.peek() == '(' {
			// It's a bare table! Decode it directly
			return decodeValue(p, rv.Elem())
		}
		// Not a bare table, restore position
		p.pos = savedPos
	}
	
	return decodeValue(p, rv.Elem())
}

func decodeValue(p *parser, target reflect.Value) error {
	p.skipSpaces()
	
	switch target.Kind() {
	case reflect.Ptr:
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return decodeValue(p, target.Elem())
		
	case reflect.Struct:
		return decodeStruct(p, target)
		
	case reflect.Map:
		return decodeMap(p, target)
		
	case reflect.Slice:
		return decodeSlice(p, target)
		
	case reflect.String:
		val, err := parseStringValue(p)
		if err != nil {
			return err
		}
		target.SetString(val)
		return nil
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := parseNumber(p)
		if err != nil {
			return err
		}
		target.SetInt(int64(val))
		return nil
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := parseNumber(p)
		if err != nil {
			return err
		}
		target.SetUint(uint64(val))
		return nil
		
	case reflect.Float32, reflect.Float64:
		val, err := parseNumber(p)
		if err != nil {
			return err
		}
		target.SetFloat(val)
		return nil
		
	case reflect.Bool:
		val, err := parseBool(p)
		if err != nil {
			return err
		}
		target.SetBool(val)
		return nil
		
	case reflect.Interface:
		// Decode as generic value
		val, err := parseGenericValue(p)
		if err != nil {
			return err
		}
		if val == nil {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}
		target.Set(reflect.ValueOf(val))
		return nil
		
	default:
		return fmt.Errorf("unsupported target type: %v", target.Kind())
	}
}

func decodeStruct(p *parser, target reflect.Value) error {
	if p.peek() != '{' {
		return fmt.Errorf("expected '{' for struct, got '%c'", p.peek())
	}
	p.next() // consume '{'
	p.skipSpaces()
	
	t := target.Type()
	fieldMap := make(map[string]int) // field name -> field index
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName := field.Tag.Get("god")
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}
		fieldMap[fieldName] = i
	}
	
	for !p.eof() && p.peek() != '}' {
		// Parse key
		key := p.readBareToken()
		p.skipSpaces()
		
		if p.peek() != '=' {
			return fmt.Errorf("expected '=' after key '%s'", key)
		}
		p.next() // consume '='
		p.skipSpaces()
		
		// Check for empty value
		if p.peek() == ';' || p.peek() == '}' {
			if p.peek() == ';' {
				p.next()
			}
			p.skipSpaces()
			// Find field and set zero value
			fieldIdx, ok := fieldMap[key]
			if ok {
				fieldVal := target.Field(fieldIdx)
				fieldVal.Set(reflect.Zero(fieldVal.Type()))
			}
			continue
		}
		
		// Find field
		fieldIdx, ok := fieldMap[key]
		if !ok {
			// Skip unknown field
			if err := skipValue(p); err != nil {
				return err
			}
		} else {
			fieldVal := target.Field(fieldIdx)
			if err := decodeValue(p, fieldVal); err != nil {
				return err
			}
		}
		
		p.skipSpaces()
		// Optional semicolon (rule 17)
		if p.peek() == ';' {
			p.next()
		}
		p.skipSpaces()
	}
	
	if p.peek() != '}' {
		return errors.New("expected '}' at end of struct")
	}
	p.next() // consume '}'
	
	return nil
}

func decodeMap(p *parser, target reflect.Value) error {
	if p.peek() != '{' {
		return fmt.Errorf("expected '{' for map, got '%c'", p.peek())
	}
	p.next() // consume '{'
	p.skipSpaces()
	
	if target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	}
	
	keyType := target.Type().Key()
	valType := target.Type().Elem()
	
	for !p.eof() && p.peek() != '}' {
		// Parse key
		keyStr := p.readBareToken()
		p.skipSpaces()
		
		// Skip empty keys (can happen with extra whitespace/semicolons)
		if keyStr == "" {
			if p.peek() == ';' {
				p.next()
				p.skipSpaces()
			}
			continue
		}
		
		if p.peek() != '=' {
			return fmt.Errorf("expected '=' after key '%s', got '%c' at position %d", keyStr, p.peek(), p.pos)
		}
		p.next() // consume '='
		p.skipSpaces()
		
		// Create key value
		keyVal := reflect.New(keyType).Elem()
		keyVal.SetString(keyStr) // Assuming string keys
		
		// Check for empty value
		if p.peek() == ';' || p.peek() == '}' {
			if p.peek() == ';' {
				p.next()
			}
			p.skipSpaces()
			// Set zero value in map
			target.SetMapIndex(keyVal, reflect.Zero(valType))
			continue
		}
		
		// Parse value
		val := reflect.New(valType).Elem()
		if err := decodeValue(p, val); err != nil {
			return err
		}
		
		target.SetMapIndex(keyVal, val)
		
		p.skipSpaces()
		// Optional semicolon
		if p.peek() == ';' {
			p.next()
		}
		p.skipSpaces()
	}
	
	if p.peek() != '}' {
		return errors.New("expected '}' at end of map")
	}
	p.next() // consume '}'
	
	return nil
}

func decodeSlice(p *parser, target reflect.Value) error {
	p.skipSpaces()
	
	// Check if it's a table format (for struct slices)
	if p.peek() == '(' {
		return decodeTable(p, target)
	}
	
	// Regular list format
	if p.peek() != '[' {
		return fmt.Errorf("expected '[' or '(' for slice, got '%c'", p.peek())
	}
	p.next() // consume '['
	p.skipSpaces()
	
	elemType := target.Type().Elem()
	slice := reflect.MakeSlice(target.Type(), 0, 0)
	
	for !p.eof() && p.peek() != ']' {
		elem := reflect.New(elemType).Elem()
		if err := decodeValue(p, elem); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
		
		p.skipSpaces()
		if p.peek() == ',' {
			p.next()
			p.skipSpaces()
		}
	}
	
	if p.peek() != ']' {
		return errors.New("expected ']' at end of list")
	}
	p.next() // consume ']'
	
	target.Set(slice)
	return nil
}

func decodeTable(p *parser, target reflect.Value) error {
	if p.peek() != '(' {
		return fmt.Errorf("expected '(' for table, got '%c'", p.peek())
	}
	p.next() // consume '('
	p.skipSpaces()
	
	elemType := target.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("table format only supported for struct slices")
	}
	
	// Parse header
	var headers []string
	for {
		p.skipSpaces()
		if p.peek() == ':' {
			p.next()
			break
		}
		if p.peek() == ')' {
			p.next()
			return nil // Empty table
		}
		
		token := p.readUntilAny(",:")
		token = strings.TrimSpace(token)
		if token != "" {
			headers = append(headers, token)
		}
		
		p.skipSpaces()
		if p.peek() == ',' {
			p.next()
		}
	}
	
	// Build field map
	fieldMap := make(map[string]int)
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName := field.Tag.Get("god")
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}
		fieldMap[fieldName] = i
	}
	
	// Parse rows
	slice := reflect.MakeSlice(target.Type(), 0, 0)
	
	for {
		p.skipSpaces()
		if p.peek() == ')' {
			p.next()
			break
		}
		
		// Create new struct
		structVal := reflect.New(elemType).Elem()
		
		// Parse cells
		cellIdx := 0
		for {
			p.skipSpaces()
			if p.peek() == ';' {
				p.next()
				break
			}
			if p.peek() == ')' {
				break
			}
			
			// Parse cell value
			var cellStr string
			if p.peek() == '"' {
				val, err := parseStringValue(p)
				if err != nil {
					return err
				}
				cellStr = val
			} else {
				cellStr = p.readUntilAny(",;)")
				cellStr = strings.TrimSpace(cellStr)
			}
			
			// Set field value
			if cellIdx < len(headers) {
				headerName := headers[cellIdx]
				if fieldIdx, ok := fieldMap[headerName]; ok {
					field := structVal.Field(fieldIdx)
					if err := setFieldFromString(field, cellStr); err != nil {
						return err
					}
				}
			}
			
			cellIdx++
			p.skipSpaces()
			if p.peek() == ',' {
				p.next()
			}
		}
		
		slice = reflect.Append(slice, structVal)
	}
	
	target.Set(slice)
	return nil
}

func setFieldFromString(field reflect.Value, s string) error {
	if s == "" {
		return nil
	}
	
	switch field.Kind() {
	case reflect.String:
		field.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}

// ===================== PARSER HELPERS =====================

type parser struct {
	src []byte
	pos int
}

func (p *parser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) next() byte {
	if p.eof() {
		return 0
	}
	c := p.src[p.pos]
	p.pos++
	return c
}

func (p *parser) skipSpaces() {
	for !p.eof() && (p.peek() == ' ' || p.peek() == '\n' || p.peek() == '\r' || p.peek() == '\t') {
		p.pos++
	}
}

func (p *parser) readBareToken() string {
	p.skipSpaces()
	var buf bytes.Buffer
	for !p.eof() {
		c := p.peek()
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '=' || c == ';' || c == '{' || c == '}' || c == '[' || c == ']' || c == '(' || c == ')' || c == ',' || c == ':' {
			break
		}
		buf.WriteByte(p.next())
	}
	return strings.TrimSpace(buf.String())
}

func (p *parser) readUntilAny(seps string) string {
	start := p.pos
	for !p.eof() {
		if strings.ContainsRune(seps, rune(p.peek())) {
			break
		}
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func parseStringValue(p *parser) (string, error) {
	if p.peek() == '"' {
		// Check for triple quote
		if p.peekAhead(3) == `"""` {
			return parseTripleString(p)
		}
		return parseString(p)
	}
	// Bare token
	return p.readBareToken(), nil
}

func parseString(p *parser) (string, error) {
	if p.next() != '"' {
		return "", errors.New("expected '\"' at start of string")
	}
	var buf bytes.Buffer
	for !p.eof() {
		c := p.next()
		if c == '\\' {
			if p.eof() {
				return "", errors.New("unterminated escape in string")
			}
			nc := p.next()
			switch nc {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case '\\':
				buf.WriteByte('\\')
			case '"':
				buf.WriteByte('"')
			default:
				buf.WriteByte(nc)
			}
			continue
		}
		if c == '"' {
			return buf.String(), nil
		}
		buf.WriteByte(c)
	}
	return "", errors.New("unterminated string")
}

func parseTripleString(p *parser) (string, error) {
	if p.peekAhead(3) != `"""` {
		return "", errors.New("expected triple quote")
	}
	p.pos += 3
	start := p.pos
	for !p.eof() {
		if p.peekAhead(3) == `"""` {
			segment := string(p.src[start:p.pos])
			p.pos += 3
			return segment, nil
		}
		p.pos++
	}
	return "", errors.New("unterminated triple-quoted string")
}

func (p *parser) peekAhead(n int) string {
	if p.pos+n > len(p.src) {
		return string(p.src[p.pos:])
	}
	return string(p.src[p.pos : p.pos+n])
}

func parseNumber(p *parser) (float64, error) {
	token := p.readBareToken()
	if token == "" {
		return 0, errors.New("expected number")
	}
	return strconv.ParseFloat(token, 64)
}

func parseBool(p *parser) (bool, error) {
	token := p.readBareToken()
	if token == "true" {
		return true, nil
	}
	if token == "false" {
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean: %s", token)
}

func parseGenericValue(p *parser) (interface{}, error) {
	p.skipSpaces()
	c := p.peek()
	if c == '{' {
		m := make(map[string]interface{})
		err := decodeMap(p, reflect.ValueOf(&m).Elem())
		return m, err
	}
	if c == '[' {
		var s []interface{}
		err := decodeSlice(p, reflect.ValueOf(&s).Elem())
		return s, err
	}
	if c == '(' {
		return nil, errors.New("generic table decoding not implemented yet")
	}
	if c == '"' {
		return parseStringValue(p)
	}
	if c == 't' || c == 'f' {
		return parseBool(p)
	}
	
	// Check for \0
	if p.pos+1 < len(p.src) && p.src[p.pos] == '\\' && p.src[p.pos+1] == '0' {
		p.pos += 2
		return nil, nil // Return nil for \0
	}

	return parseNumber(p)
}

func skipValue(p *parser) error {
	p.skipSpaces()
	c := p.peek()
	
	switch c {
	case '{':
		depth := 0
		for !p.eof() {
			if p.peek() == '{' {
				depth++
			} else if p.peek() == '}' {
				depth--
				p.next()
				if depth == 0 {
					return nil
				}
				continue
			}
			p.next()
		}
	case '[':
		depth := 0
		for !p.eof() {
			if p.peek() == '[' {
				depth++
			} else if p.peek() == ']' {
				depth--
				p.next()
				if depth == 0 {
					return nil
				}
				continue
			}
			p.next()
		}
	case '(':
		depth := 0
		for !p.eof() {
			if p.peek() == '(' {
				depth++
			} else if p.peek() == ')' {
				depth--
				p.next()
				if depth == 0 {
					return nil
				}
				continue
			}
			p.next()
		}
	case '"':
		parseStringValue(p)
	default:
		p.readBareToken()
	}
	return nil
}
