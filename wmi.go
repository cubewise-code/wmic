package wmic

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"

	"github.com/jszwec/csvutil"
)

// RecordError holds information about an error for record in the WMI result
type RecordError struct {
	Class   string
	Line    int
	Items   []string
	Message string
}

// Record returns the line as a csv
func (e *RecordError) Record() string {
	return strings.Join(e.Items, ",")
}

// QueryAll returns all items with all columns
func QueryAll(class string, out interface{}) ([]RecordError, error) {
	return Query(class, []string{}, "", out)
}

// QueryColumns returns all items with specific columns
func QueryColumns(class string, columns []string, out interface{}) ([]RecordError, error) {
	return Query(class, columns, "", out)
}

// QueryWhere returns all columns for where clause
func QueryWhere(class, where string, out interface{}) ([]RecordError, error) {
	return Query(class, []string{}, where, out)
}

// Query returns a WMI query with the given parameters
func Query(class string, columns []string, where string, out interface{}) ([]RecordError, error) {

	recordErrors := []RecordError{}

	query := []string{"PATH", class}
	if where != "" {
		parts := strings.Split(strings.TrimSpace(where), " ")
		query = append(query, "WHERE")
		if !strings.HasPrefix(parts[0], "(") {
			query = append(query, "(")
		}
		query = append(query, parts...)
		if !strings.HasSuffix(parts[len(parts)-1], ")") {
			query = append(query, ")")
		}
	}
	query = append(query, "GET")
	if len(columns) > 0 {
		query = append(query, strings.Join(columns, ","))
	}
	query = append(query, "/format:csv")
	cmd := exec.Command("wmic", query...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return recordErrors, err
	}
	if stderr.Len() > 0 {
		return recordErrors, errors.New(string(stderr.Bytes()))
	}
	str := string(stdout.Bytes())
	scanner := bufio.NewScanner(strings.NewReader(str))
	var sb strings.Builder
	for scanner.Scan() {
		s := scanner.Text()
		if strings.TrimSpace(s) != "" {
			sb.WriteString(strings.ReplaceAll(s, "\"", ""))
			sb.WriteString("\n")
		}
	}

	// Get the outer type (needs to be a slice)
	outerValue := reflect.ValueOf(out)
	if outerValue.Kind() == reflect.Ptr {
		outerValue = outerValue.Elem()
	}

	if outerValue.Kind() != reflect.Slice {
		return recordErrors, fmt.Errorf("You must provide a slice to the out argument")
	}

	// Get the inner type of the slice
	innerType := outerValue.Type().Elem()
	innerTypeIsPointer := false
	if innerType.Kind() == reflect.Ptr {
		// If a pointer get the underlying type
		innerTypeIsPointer = true
		innerType = innerType.Elem()
	}

	if innerType.Kind() != reflect.Struct {
		return recordErrors, fmt.Errorf("You must provide a struct as the type of the out slice")
	}

	source := sb.String()

	csvReader := csv.NewReader(strings.NewReader(source))
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		return recordErrors, err
	}

	result := make([]interface{}, 0)

	for {
		// Loop through all of the results and populate result slice
		i := reflect.New(innerType).Interface()
		if err := dec.Decode(&i); err == io.EOF {
			break
		} else if csvError, ok := err.(*csv.ParseError); ok {
			// Ignore parsing error
			items := dec.Record()
			recordErrors = append(recordErrors, RecordError{Class: class, Items: items, Line: csvError.Line, Message: csvError.Error()})
			continue
		} else if err != nil {
			// Error so exit function
			return recordErrors, err
		}
		result = append(result, i)
	}

	// Resize the out slice to match the number of records
	outerValue.Set(reflect.MakeSlice(outerValue.Type(), len(result), len(result)))

	for i, val := range result {
		// Update the out slice with each item
		v := reflect.ValueOf(val)
		if innerTypeIsPointer {
			outerValue.Index(i).Set(v)
		} else {
			outerValue.Index(i).Set(v.Elem())
		}
	}

	return recordErrors, nil
}
