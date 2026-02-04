package wordtmpl

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMsWordSaveDocxAllKeys fills all placeholders and writes result.docx.
func TestMsWordSaveDocxAllKeys(t *testing.T) {
	ctx := context.Background()
	tmplPath := "testdata/template.docx"

	doc, err := Open(ctx, tmplPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	values := map[string]string{
		"award_1":               "Employee of the Year 2022",
		"award_2":               "Hackathon Winner 2021",
		"cert_1":                "AWS Certified Solutions Architect",
		"cert_2":                "Google Data Analytics Certificate",
		"city":                  "Austin",
		"country":               "USA",
		"education_1_city":      "Austin",
		"education_1_degree":    "BSc Computer Science",
		"education_1_details":   "GPA 3.8, Dean's List",
		"education_1_end":       "May 2018",
		"education_1_school":    "University of Texas",
		"education_1_start":     "Sep 2014",
		"education_2_city":      "Online",
		"education_2_degree":    "Professional Certificate",
		"education_2_details":   "Data Engineering Track",
		"education_2_end":       "Nov 2020",
		"education_2_school":    "Coursera",
		"education_2_start":     "Jun 2020",
		"email":                 "jordan.lee@example.com",
		"experience_1_bullet_1": "Built ETL pipelines processing 2TB/day",
		"experience_1_bullet_2": "Reduced query latency by 45%",
		"experience_1_city":     "Austin",
		"experience_1_company":  "Northwind Analytics",
		"experience_1_end":      "Aug 2023",
		"experience_1_position": "Senior Data Engineer",
		"experience_1_result_1": "Cut infrastructure costs by 20%",
		"experience_1_start":    "Jan 2021",
		"experience_2_bullet_1": "Migrated legacy SQL to cloud data warehouse",
		"experience_2_bullet_2": "Implemented CI for data models",
		"experience_2_city":     "Dallas",
		"experience_2_company":  "Bluejay Systems",
		"experience_2_end":      "Dec 2020",
		"experience_2_position": "Data Engineer",
		"experience_2_result_1": "Improved reporting freshness to hourly",
		"experience_2_start":    "Jul 2018",
		"extra":                 "Open-source contributor and conference speaker",
		"first_name":            "Jordan",
		"job_title":             "Data Engineer",
		"language_1":            "English",
		"language_1_level":      "Native",
		"language_2":            "Spanish",
		"language_2_level":      "Professional",
		"last_name":             "Lee",
		"linkedin":              "linkedin.com/in/jordanlee",
		"phone":                 "+1 512-555-0199",
		"project_1_description": "Designed a streaming ingestion system on Kafka",
		"project_1_link":        "github.com/jordanlee/streaming-ingest",
		"project_1_name":        "StreamForge",
		"project_1_note":        "Open-source",
		"project_1_stack":       "Go, Kafka, PostgreSQL, Docker",
		"skill_1":               "Go",
		"skill_2":               "SQL",
		"skill_3":               "Kafka",
		"skill_4":               "Airflow",
		"skill_5":               "Docker",
		"summary":               "Data engineer with 6+ years of experience building reliable pipelines.",
		"tool_1":                "dbt",
		"tool_2":                "Terraform",
		"tool_3":                "Grafana",
		"website":               "jordanlee.dev",
	}

	for key, value := range values {
		doc.Val(key, value)
	}

	for _, key := range doc.Keys() {
		if _, ok := values[key]; ok {
			continue
		}
		doc.Val(key, "VAL_"+key)
	}

	outPath := filepath.Join(filepath.Dir(tmplPath), "result.docx")
	if err := doc.Save(ctx, outPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	t.Logf("output saved to %s", outPath)

	content, err := readDocxPart(outPath, "word/document.xml")
	if err != nil {
		t.Fatalf("read output document.xml: %v", err)
	}

	if strings.Contains(content, "{first_name}") || strings.Contains(content, "{last_name}") {
		t.Fatalf("placeholders were not replaced in output")
	}
	if !strings.Contains(content, "Jordan") || !strings.Contains(content, "Lee") {
		t.Fatalf("expected replacement text not found in output")
	}
}

// TestMsWordSaveBytes validates the binary API for docx content.
func TestMsWordSaveBytes(t *testing.T) {
	ctx := context.Background()
	tmplPath := "testdata/template.docx"

	data, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	doc, err := OpenBytes(ctx, data)
	if err != nil {
		t.Fatalf("OpenBytes failed: %v", err)
	}

	doc.Val("first_name", "Jordan")
	doc.Val("last_name", "Lee")

	out, err := doc.SaveBytes(ctx)
	if err != nil {
		t.Fatalf("SaveBytes failed: %v", err)
	}

	content, err := readDocxPartFromBytes(out, "word/document.xml")
	if err != nil {
		t.Fatalf("read output document.xml: %v", err)
	}
	if strings.Contains(content, "{first_name}") || strings.Contains(content, "{last_name}") {
		t.Fatalf("placeholders were not replaced in output")
	}
	if !strings.Contains(content, "Jordan") || !strings.Contains(content, "Lee") {
		t.Fatalf("expected replacement text not found in output")
	}
}

// readDocxPart reads a named XML file from a .docx zip container.
func readDocxPart(path, name string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", io.EOF
}

// readDocxPartFromBytes reads a named XML file from an in-memory .docx.
func readDocxPartFromBytes(data []byte, name string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		content, err := io.ReadAll(rc)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	return "", io.EOF
}
