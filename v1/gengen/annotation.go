package gengen

import (
	"strings"
	"text/scanner"
)

type Annotation struct {
	Name       string
	Attributes map[string]string
}

type status int

const (
	initial status = iota
	annotationName
	attributeName
	attributeValue
	done
)

func parseAnnotation(line string) *Annotation {
	withoutComment := strings.TrimLeft(strings.TrimSpace(line), "/")

	annotation := Annotation{
		Name:       "",
		Attributes: make(map[string]string),
	}

	var s scanner.Scanner
	s.Init(strings.NewReader(withoutComment))

	var tok rune
	currentStatus := initial
	var attrName string

	for tok != scanner.EOF && currentStatus < done {
		tok = s.Scan()
		switch tok {
		case '@':
			if currentStatus != initial {
				// fmt.Println(1, currentStatus)
				return nil
			}

			currentStatus = annotationName
		case '.':
			if currentStatus != annotationName {
				// fmt.Println(2, currentStatus)
				return nil
			}
			annotation.Name += "."
		case '(':
			if currentStatus != annotationName {
				// fmt.Println(3, currentStatus)
				return nil
			}
			currentStatus = attributeName
		case '=':
			if currentStatus != attributeName {
				// fmt.Println(4, currentStatus)
				return nil
			}

			currentStatus = attributeValue
		case ',':
			if currentStatus != attributeValue {
				// fmt.Println(5, currentStatus)
				return nil
			}

			currentStatus = attributeName
		case ')':
			if currentStatus != attributeName && currentStatus != attributeValue {
				// fmt.Println(6, currentStatus)
				return nil
			}

			currentStatus = done
		case scanner.Ident:
			switch currentStatus {
			case annotationName:
				annotation.Name += s.TokenText()
			case attributeName:
				attrName = s.TokenText()
			default:
				// fmt.Println(7, currentStatus, s.TokenText())
				return nil
			}
		default:
			switch currentStatus {
			case attributeValue:
				annotation.Attributes[strings.ToLower(attrName)] = strings.Trim(s.TokenText(), "\"")
			default:
				// fmt.Println(8, currentStatus)
				return nil
			}
		}
	}
	if currentStatus != done {
		// fmt.Println(9, currentStatus)
		return nil
	}

	// fmt.Println("ok", annotation)
	return &annotation
}
