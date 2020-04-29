package wmic

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/jszwec/csvutil"
)

// QueryAll returns all items with all columns
func QueryAll(class string, out interface{}) error {
	return Query(class, []string{}, "", out)
}

// QueryColumns returns all items with specific columns
func QueryColumns(class string, columns []string, out interface{}) error {
	return Query(class, columns, "", out)
}

// QueryWhere returns all columns for where clause
func QueryWhere(class, where string, out interface{}) error {
	return Query(class, []string{}, where, out)
}

// Query returns a WMI query with the given parameters
func Query(class string, columns []string, where string, out interface{}) error {
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
		return err
	}
	if stderr.Len() > 0 {
		return errors.New(string(stderr.Bytes()))
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

	outerValue := reflect.ValueOf(out)
	if outerValue.Kind() == reflect.Ptr {
		outerValue = outerValue.Elem()
	}

	innerType := outerValue.Type().Elem()
	innerTypeIsPointer := false
	if innerType.Kind() == reflect.Ptr {
		innerTypeIsPointer = true
		innerType = innerType.Elem()
	}

	source := sb.String()

	csvReader := csv.NewReader(strings.NewReader(source))
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		return err
	}

	result := make([]interface{}, 0)

	for {
		i := reflect.New(innerType).Interface()
		if err := dec.Decode(&i); err == io.EOF {
			break
		} else if _, ok := err.(*csv.ParseError); ok {
			// Ignore parsing error
			items := dec.Record()
			if os.Getenv("WMIDebug") != "" {
				log.Println(class + " " + err.Error() + ": " + strings.Join(items, ","))
			}
			continue
		} else if err != nil {
			return err
		}
		result = append(result, i)
	}

	outerValue.Set(reflect.MakeSlice(outerValue.Type(), len(result), len(result)))

	for i, val := range result {
		v := reflect.ValueOf(val)
		if innerTypeIsPointer {
			outerValue.Index(i).Set(v)
		} else {
			outerValue.Index(i).Set(v.Elem())
		}
	}

	return nil
}
