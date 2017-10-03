package common

import (
	"errors"
	"fmt"
)

const (
	ROOT = "ROOT"
	MAP  = "MAP"
	LIST = "LIST"

	INVALID_FIELD_PATH        = "Invalid fieldPath '%s' at char '%s'"
	INVALID_FIELD_PATH_REASON = "Invalid fieldPath '%s' at char '%s' (%s)"
	REASON_EMPTY_FIELD_NAME   = "field name can't be empty"
	REASON_INVALID_START      = "field path needs to start with '[' or '/'"
	REASON_NOT_A_NUMBER       = "only numbers and '*' allowed between '[' and ']'"
	REASON_QUOTES             = "quotes are not properly closed"
	INVALID_FIELD_PATH_NUMBER = "Invalid fieldPath '%s' at char '%s' ('%s' needs to be a number or '*')"
)

type PathElement struct {
	Type string
	Name string
	Idx  int
}

var ROOT_PATH_ELEMENT = &PathElement{
	Type: ROOT,
	Name: "",
	Idx:  0,
}

func CreateMapElement(name string) PathElement {
	return PathElement{
		Type: MAP,
		Name: name,
		Idx:  0,
	}
}

func CreateListElement(idx int) PathElement {
	return PathElement{
		Type: LIST,
		Name: "",
		Idx:  idx,
	}
}

func ParseFieldPath(fieldPath string, isSingleQuoteEscaped bool) ([]PathElement, error) {
	pathElementList := make([]PathElement, 0)
	pathElementList = append(pathElementList, *ROOT_PATH_ELEMENT)

	if len(fieldPath) > 0 {
		requiresStart := true
		requiresName := false
		requiresIndex := false
		singleQuote := false
		doubleQuote := false
		collector := ""
		pos := 0
		for pos = 0; pos < len(fieldPath); pos++ {
			fmt.Println(fieldPath[pos])
			if requiresStart {
				requiresStart = false
				requiresName = false
				requiresIndex = false
				singleQuote = false
				doubleQuote = false
				switch fieldPath[pos] {
				case '/':
					requiresName = true
					break
				case '[':
					requiresIndex = true
					break
				default:
					return nil, errors.New(fmt.Sprintf(INVALID_FIELD_PATH_REASON, fieldPath, 0, REASON_INVALID_START))
				}
			} else {
				if requiresName {
					switch fieldPath[pos] {
					case '\\':
						if pos == 0 || fieldPath[pos-1] != '\\' {
							if !doubleQuote {
								singleQuote = !singleQuote
							} else {
								collector += string(fieldPath[pos])
							}
						} else {
							collector = collector[0 : len(collector)-1]
							collector += string(fieldPath[pos])
						}
					case '"':
						if pos == 0 || fieldPath[pos] != '\\' {
							if !singleQuote {
								doubleQuote = !doubleQuote
							} else {
								collector += string(fieldPath[pos])
							}
						} else {
							collector = collector[0 : len(collector)-1]
							collector += string(fieldPath[pos])
						}
					case '/':
						fallthrough
					case '[':
						fallthrough
					case ']':
						if singleQuote || doubleQuote {
							collector += string(fieldPath[pos])
						} else {
							if len(fieldPath) <= pos+1 {
								return nil, errors.New(
									fmt.Sprintf(INVALID_FIELD_PATH_REASON, fieldPath, pos, REASON_EMPTY_FIELD_NAME),
								)
							}
							if fieldPath[pos] == fieldPath[pos+1] {
								collector += string(fieldPath[pos])
								pos++
							} else {
								pathElementList = append(pathElementList, CreateMapElement(collector))
								requiresStart = true
								collector = ""
								//not very kosher, we need to replay the current char as start of path element
								pos--
							}
						}
					default:
						collector += string(fieldPath[pos])
					}
				} else if requiresIndex {
					switch fieldPath[pos] {
					case '0':
						fallthrough
					case '1':
						fallthrough
					case '2':
						fallthrough
					case '3':
						fallthrough
					case '4':
						fallthrough
					case '5':
						fallthrough
					case '6':
						fallthrough
					case '7':
						fallthrough
					case '8':
						fallthrough
					case '9':
						fallthrough
					case '*': //wildcard character
						collector += string(fieldPath[pos])
					case ']':
						// TODO: add code
					default:
						return nil, errors.New(
							fmt.Sprintf(INVALID_FIELD_PATH_REASON, fieldPath, pos, REASON_NOT_A_NUMBER),
						)
					}
				}
			}
		}

		if singleQuote || doubleQuote {
			// If there is no matching quote
			return nil, errors.New(fmt.Sprintf(INVALID_FIELD_PATH_REASON, fieldPath, 0, REASON_QUOTES))
		} else if pos < len(fieldPath) {
			return nil, errors.New(fmt.Sprintf(INVALID_FIELD_PATH, fieldPath, pos))
		} else if len(collector) > 0 {
			// the last path element was a map entry, we need to create it.
			pathElementList = append(pathElementList, CreateMapElement(collector))
		}
	}

	return pathElementList, nil
}
