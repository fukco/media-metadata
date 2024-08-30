package box

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type (
	stringType uint8
	fieldFlag  uint8
)

/*
Name Semantics
utf8string UTF-8 string as defined in IETF RFC 3629, null-terminated.
utfstring null-terminated string encoded using either UTF-8 or UTF-16.
If UTF-16 is used, the sequence of bytes shall start with a byte order mark (BOM) and the
null termination shall be 2 bytes set to 0.
utf8list null-terminated list of space-separated UTF-8 strings
base64string null-terminated base64 encoded data
*/

const (
	stringType_utf8string stringType = iota + 1
	stringType_utfstring
)

const (
	fieldString fieldFlag = 1 << iota
	fieldDec
	fieldHex
	fieldLengthDynamic
)

const LengthUnlimited = math.MaxUint32

type field struct {
	children []*field
	name     string
	constant string
	size     uint // bits size
	length   uint
	flags    fieldFlag
	strType  stringType
	version  uint8
}

func (f *field) set(flag fieldFlag) {
	f.flags |= flag
}

func (f *field) is(flag fieldFlag) bool {
	return f.flags&flag != 0
}

func buildFields(payload Boxer) []*field {
	t := reflect.TypeOf(payload).Elem()
	return buildFieldsStruct(t)
}

func buildFieldsStruct(t reflect.Type) []*field {
	fs := make([]*field, 0, 8)
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i).Type
		tag, ok := t.Field(i).Tag.Lookup("mp4")
		if !ok {
			continue
		}
		f := buildField(t.Field(i).Name, tag)
		f.children = buildFieldsAny(ft)
		fs = append(fs, f)
	}
	return fs
}

func buildFieldsAny(t reflect.Type) []*field {
	switch t.Kind() {
	case reflect.Struct:
		return buildFieldsStruct(t)
	case reflect.Ptr, reflect.Array, reflect.Slice:
		return buildFieldsAny(t.Elem())
	default:
		return nil
	}
}

func buildField(fieldName string, tag string) *field {
	f := &field{
		name: fieldName,
	}
	tagMap := parseFieldTag(tag)

	if val, contained := tagMap["string"]; contained {
		f.set(fieldString)
		if val == "utfstring" {
			f.strType = stringType_utfstring
		} else if val == "utf8string" {
			f.strType = stringType_utf8string
		}
	}

	if val, contained := tagMap["const"]; contained {
		f.constant = val
	}

	f.version = anyVersion
	if val, contained := tagMap["ver"]; contained {
		ver, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		if ver > math.MaxUint8 {
			panic("ver-tag must be <=255")
		}
		f.version = uint8(ver)
	}

	if val, contained := tagMap["size"]; contained {
		size, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			panic(err)
		}
		f.size = uint(size)
	}

	f.length = LengthUnlimited
	if val, contained := tagMap["len"]; contained {
		if val == "dynamic" {
			f.set(fieldLengthDynamic)
		} else {
			l, err := strconv.ParseUint(val, 10, 32)
			if err != nil {
				panic(err)
			}
			f.length = uint(l)
		}
	}

	return f
}

func parseFieldTag(str string) map[string]string {
	tag := make(map[string]string, 8)

	list := strings.Split(str, ",")
	for _, e := range list {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) == 2 {
			tag[strings.Trim(kv[0], " ")] = strings.Trim(kv[1], " ")
		} else {
			tag[strings.Trim(kv[0], " ")] = ""
		}
	}

	return tag
}

func resolveFieldInstance(f *field, box Boxer, parent reflect.Value, ctx *Context) {
	fielder, ok := parent.Addr().Interface().(CustomFielder)
	var customFielder CustomFielder
	if ok {
		customFielder = fielder
	} else {
		customFielder = box
	}

	if f.is(fieldLengthDynamic) {
		f.length = customFielder.GetFieldLength(f.name, ctx)
	}
}

func isTargetField(bi *BoxInfo, fi *field) bool {
	if bi.FullBoxHeader != nil {
		if fi.version != anyVersion && bi.FullBoxHeader.Version != fi.version {
			return false
		}
	}
	return true
}

func checkFieldSize(f *field) error {
	if f.size == 0 {
		return fmt.Errorf("size must not be zero: %s", f.name)
	}
	if f.size%8 != 0 {
		return fmt.Errorf("size must be multiple of 8: %s", f.name)
	}
	return nil
}
