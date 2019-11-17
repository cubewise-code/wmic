package wmic

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/gocarina/gocsv"
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
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.LazyQuotes = true
		return r
	})
	err = gocsv.UnmarshalString(sb.String(), out)
	if err != nil {
		return err
	}
	return nil
}
