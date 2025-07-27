package odata

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"os"
)

const (
	Filename = "odata.go"

	DateTimeOffsetMarshalJSONText = `
		func (d *DateTimeOffset) MarshalJSON() ([]byte, error) {
			str := "\""+d.Time.Format(time.RFC3339)+"\""
			return []byte(str), nil
		}

		func (d *DateTimeOffset) UnmarshalJSON(b []byte) error {
			str := string(b)
			t, err := time.Parse(time.RFC3339, str[1:len(str)-1])
			if err != nil {
				return err
			}
			d.Time = t
			return nil
		}
	`
	DurationMarshalJSONText = `
		type ParseError struct {
			input []byte
		}

		func (e *ParseError) Error() string {
			return "invalid duration \"" + string(e.input) + "\""
		}

		// Not shared out side the package just an indicator that there was an error so that the
		// parent function can return correctly
		var errInternal = errors.New("internal error")

		func (d *Duration) MarshalJSON() ([]byte, error) {
			buf := bytes.NewBuffer(nil)
			buf.WriteString("P")
			hours := d.Duration.Hours()
			if hours >= 24 {
				days := int(hours) / 24
				buf.WriteString(strconv.Itoa(days))
				buf.WriteString("D")
				hours = hours - float64(days*24)
			}
			if hours == 0 {
				return buf.Bytes(), nil
			}
			buf.WriteString("T")
			hoursInt := int(hours)
			if hoursInt > 0 {
				buf.WriteString(strconv.Itoa(hoursInt))
				buf.WriteString("H")
			}
			minutes := d.Duration.Minutes()
			minutesInt := int(minutes)
			if minutesInt > 0 && minutesInt%60 > 0 {
				buf.WriteString(strconv.Itoa(minutesInt % 60))
				buf.WriteString("M")
			}
			seconds := d.Duration.Seconds()
			seconds = seconds - float64(minutesInt*60)
			if seconds > 0 {
				buf.WriteString(strconv.FormatFloat(seconds, 'f', -1, 64))
				buf.WriteString("S")
			}
			return buf.Bytes(), nil
		}

		func (d *Duration) UnmarshalJSON(b []byte) error {
			if bytes.Equal([]byte("null"), b) {
				return nil
			}
			b = bytes.Trim(b, "\"")
			// Per Redfish Spec: No negative values...
			if b[0] != 'P' {
				return &ParseError{b}
			}
			days, rest, err := getDays(b[1:])
			if err != nil {
				return &ParseError{b}
			}
			d.Duration = time.Duration(days*24*int(time.Hour))
			if len(rest) == 0 {
				return nil
			}
			// Next character must be "T"
			if rest[0] != 'T' {
				return &ParseError{b}
			}
			hours, rest, err := getHours(rest[1:])
			if err != nil {
				return &ParseError{b}
			}
			d.Duration += time.Duration(hours * int(time.Hour))
			if len(rest) == 0 {
				return nil
			}
			mins, rest, err := getMinutes(rest)
			if err != nil {
				return &ParseError{b}
			}
			d.Duration += time.Duration(mins * int(time.Minute))
			if len(rest) == 0 {
				return nil
			}
			seconds, fraction, err := getSeconds(rest)
			if err != nil {
				return &ParseError{b}
			}
			d.Duration += time.Duration(seconds*int(time.Second) + int(fraction*float64(time.Second)))
			return nil
		}

		func getType(input []byte, char rune) (int, []byte, error) {
			index := bytes.IndexRune(input, char)
			if index == -1 {
				// This is fine, only seconds are required (if we get that far)
				return 0, input, nil
			}
			ret, err := strconv.Atoi(string(input[:index]))
			if err != nil {
				return 0, nil, err
			}
			return ret, input[index+1:], nil
		}

		func getDays(input []byte) (int, []byte, error) {
			return getType(input, 'D')
		}

		func getHours(input []byte) (int, []byte, error) {
			return getType(input, 'H')
		}

		func getMinutes(input []byte) (int, []byte, error) {
			return getType(input, 'M')
		}

		func getSeconds(input []byte) (int, float64, error) {
			// This one is different, if we got here and there is no S that is an error
			index := bytes.IndexRune(input, 'S')
			if index == -1 || index != len(input)-1 {
				return 0, 0, errInternal
			}
			// Seconds can have a fractional part, so check if we have that...
			fractionIndex := bytes.IndexRune(input, '.')
			if fractionIndex == -1 {
				// This is fine, no fractional seconds is valid
				seconds, err := strconv.Atoi(string(input[:index]))
				if err != nil {
					return 0, 0, err
				}
				return seconds, 0, nil
			}
			floatVal, err := strconv.ParseFloat(string(input[:index]), 64)
			if err != nil {
				return 0, 0, err
			}
			return int(floatVal), float64(floatVal - float64(int(floatVal))), nil
		}
	`

	UUIDMarshalJSONText = `
		func (u *UUID) MarshalJSON() ([]byte, error) {
		}
	`
)

func GenBoilerPlate(packageName string) error {
	fileToken := &ast.File{
		Name: ast.NewIdent(packageName),
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: `"bytes"`,
						},
					},
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: `"errors"`,
						},
					},
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: `"math/big"`,
						},
					},
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: `"strconv"`,
						},
					},
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: `"time"`,
						},
					},
				},
			},
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: ast.NewIdent("Action"),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{ast.NewIdent("Target")},
										Type:  &ast.Ident{Name: "string"},
										Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"target\"`"},
									},
									{
										Names: []*ast.Ident{ast.NewIdent("ActionInfo")},
										Type:  &ast.Ident{Name: "string"},
										Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"@Redfish.ActionInfo,omitempty\"`"},
									},
								},
							},
						},
					},
					&ast.TypeSpec{
						Name: ast.NewIdent("OdataID"),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{ast.NewIdent("ID")},
										Type:  &ast.Ident{Name: "string"},
										Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"@odata.id\"`"},
									},
								},
							},
						},
					},
					&ast.TypeSpec{
						Name: ast.NewIdent("DateTimeOffset"),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{ast.NewIdent("Time")},
										Type:  &ast.Ident{Name: "time.Time"},
									},
								},
							},
						},
					},
					&ast.TypeSpec{
						Name: ast.NewIdent("Duration"),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{ast.NewIdent("Duration")},
										Type:  &ast.Ident{Name: "time.Duration"},
									},
								},
							},
						},
					},
					&ast.TypeSpec{
						Name: ast.NewIdent("UUID"),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{ast.NewIdent("Data")},
										Type:  &ast.Ident{Name: "big.Int"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	buf := bytes.NewBuffer(nil)
	fileSet := token.NewFileSet()
	err := format.Node(buf, fileSet, fileToken)
	if err != nil {
		return err
	}
	_, err = buf.WriteString(DateTimeOffsetMarshalJSONText)
	if err != nil {
		return err
	}
	_, err = buf.WriteString(DurationMarshalJSONText)
	if err != nil {
		return err
	}
	file, err := os.Create(Filename)
	if err != nil {
		return err
	}
	//nolint:errcheck // Ignore error on close, not sure what we can do about it
	defer file.Close()
	content, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	_, err = file.Write(content)
	return err
}
