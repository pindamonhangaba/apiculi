package quick_schema

import (
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
)

var (
	marshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
)

type Node struct {
	Package     string
	Type        string
	Format      string
	Name        string
	Description string
	Example     string
	Children    []Node
}

func marshalerEncoder(v reflect.Value) Node {
	result := ""
	if v.Kind() == reflect.Pointer && v.IsNil() {
		//e.WriteString("null")
		return Node{
			Type:    v.Type().Name(),
			Package: v.Type().PkgPath(),
			Format:  "any",
		}
	}
	m, ok := v.Interface().(json.Marshaler)
	if !ok {
		//e.WriteString("null")
		return Node{
			Type:    v.Type().Name(),
			Package: v.Type().PkgPath(),
			Format:  "any",
		}
	}
	b, err := m.MarshalJSON()
	if err != nil {
		// copy JSON into buffer, checking validity.
		//err = compact(&e.Buffer, b, opts.escapeHTML)
		return Node{
			Type:    v.Type().Name(),
			Package: v.Type().PkgPath(),
			Format:  "any",
		}
	}
	result = string(b)
	isNullable := string(b) == "null"
	if string(b) == "null" && v.Kind() == reflect.Struct {
		v.FieldByName("Valid").Set(reflect.ValueOf(true))
		m, ok := v.Interface().(json.Marshaler)
		if !ok {
			//e.WriteString("null")
			return Node{
				Type:    v.Type().Name(),
				Package: v.Type().PkgPath(),
				Format:  "any",
			}
		}
		b, err = m.MarshalJSON()
		if err != nil {
			// copy JSON into buffer, checking validity.
			//err = compact(&e.Buffer, b, opts.escapeHTML)
			return Node{
				Type:    v.Type().Name(),
				Package: v.Type().PkgPath(),
				Format:  "any",
			}
		}
		result = string(b)
	}
	it := Node{
		Type:    v.Type().Name(),
		Package: v.Type().PkgPath(),
		Format:  "any",
	}
	if strings.HasPrefix(result, "\"") && strings.HasSuffix(result, "\"") {
		it.Format = "string"
	} else {
		isNotDigit := func(c rune) bool { return c < '0' || c > '9' }
		isNumeric := strings.IndexFunc(result, isNotDigit) == -1
		if isNumeric {
			it.Format = "number"
		} else if strings.HasPrefix(result, "[") && strings.HasSuffix(result, "]") {
			it.Format = "array"
			if strings.HasPrefix(result, "[true") || strings.HasPrefix(result, "[false") {
				it.Children = []Node{{
					Type:   "bolean",
					Format: "bolean",
				}}
			}
			if strings.HasPrefix(result, "[{") || strings.HasSuffix(result, "}]") {
				it.Children = []Node{{
					Type:   "object",
					Format: "object",
				}}
			}
			if strings.HasPrefix(result, "[\"") || strings.HasSuffix(result, "\"]") {
				it.Children = []Node{{
					Type:   "string",
					Format: "string",
				}}
			}
			if strings.HasPrefix(result, "[[") || strings.HasSuffix(result, "]]") {
				it.Children = []Node{{
					Type:   "array",
					Format: "array",
				}}
			}
			isDigitOrComma := func(c rune) bool { return c >= '0' || c <= '9' || c == ',' }
			isNumericArr := strings.IndexFunc(result, isDigitOrComma) == -1
			if isNumericArr {
				it.Children = []Node{{
					Type:   "number",
					Format: "number",
				}}
			}
		}
	}

	if isNullable {
		return Node{
			Type:     v.Type().Name(),
			Package:  v.Type().PkgPath(),
			Format:   "pointer",
			Children: []Node{it},
		}
	}

	return it

}

func schemaIt(t reflect.Type, f *reflect.Value) (d *Node) {
	defer func() {
		if r := recover(); r != nil {
			d = &Node{
				Type:   "panic",
				Format: "any",
			}
		}
	}()
	switch t.Kind() {
	case reflect.Map:
		if t.Key().Name() != reflect.String.String() {
			return &Node{
				Type:   "non-string-keys-map",
				Format: "any",
			}
		}
		fv := reflect.New(t.Elem())
		e := fv.Elem()
		items := []Node{}
		itm := schemaIt(t.Elem(), &e)
		if itm != nil {
			items = append(items, *itm)
		}
		return &Node{
			Type:     t.Name(),
			Package:  t.PkgPath(),
			Format:   "map",
			Children: items,
		}
	case reflect.Slice:
		s1 := reflect.SliceOf(t)
		tt := s1.Elem().Elem()
		fv := reflect.New(tt)
		e := fv.Elem()

		items := []Node{}
		itm := schemaIt(tt, &e)
		if itm != nil {
			items = append(items, *itm)
		}
		return &Node{
			Type:     t.Name(),
			Package:  t.PkgPath(),
			Format:   "array",
			Children: items,
		}
	case reflect.Array:
		s1 := reflect.ArrayOf(1, t)
		tt := s1.Elem().Elem()
		fv := reflect.New(tt.Elem())
		e := fv.Elem()

		items := []Node{}
		itm := schemaIt(tt, &e)
		if itm != nil {
			items = append(items, *itm)
		}
		return &Node{
			Type:     t.Name(),
			Package:  t.PkgPath(),
			Format:   "array",
			Children: items,
		}
	case reflect.Ptr:
		fv := reflect.New(t.Elem())
		e := fv.Elem()
		items := []Node{}
		itm := schemaIt(e.Type(), &e)
		if itm != nil {
			items = append(items, *itm)
		}
		return &Node{
			Type:     t.Name(),
			Package:  t.PkgPath(),
			Format:   "pointer",
			Children: items,
		}
	case reflect.String:
		return &Node{
			Type:    t.Name(),
			Package: t.PkgPath(),
			Format:  "string",
		}
	case reflect.Bool:
		return &Node{
			Type:    t.Name(),
			Package: t.PkgPath(),
			Format:  "boolean",
		}
	case reflect.Struct:
		if t.Implements(marshalerType) {
			enc := marshalerEncoder(*f)
			return &enc
		}
		items := []Node{}
		for i := 0; i < f.NumField(); i++ {
			v := f.Field(i)
			vv := t.Field(i)
			if !vv.IsExported() {
				continue
			}
			jsontag := strings.TrimSpace(vv.Tag.Get("json"))
			if jsontag == "-" {
				continue
			}
			name, _ := parseTag(jsontag)
			if !isValidTag(name) {
				name = vv.Name
			}

			itm := schemaIt(vv.Type, &v)
			if itm != nil {
				if !val(itm.Name) {
					itm.Name = name
				}
				items = append(items, *itm)
			}

			d := strings.TrimSpace(vv.Tag.Get("description"))
			if val(d) {
				itm.Description = d
			}
			e := strings.TrimSpace(vv.Tag.Get("example"))
			if val(e) {
				itm.Example = e
			}
			f := strings.TrimSpace(vv.Tag.Get("format"))
			if val(f) {
				itm.Format = f
			}

		}
		return &Node{
			Type:     t.Name(),
			Package:  t.PkgPath(),
			Format:   "object",
			Children: items,
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		return &Node{
			Type:    t.Name(),
			Package: t.PkgPath(),
			Format:  "number",
		}
	default:
	}
	return d
}

func GetSchema[T any]() *Node {
	t := *new(T)
	tt := reflect.TypeOf(t)
	if tt == nil {
		return nil
	}
	ptr := reflect.New(tt)
	e := ptr.Elem()
	its := schemaIt(tt, &e)
	return its
}

func parseTag(tag string) (string, string) {
	tag, opt, _ := strings.Cut(tag, ",")
	return tag, (opt)
}

func val(s string) bool {
	return len(s) > 0
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}
