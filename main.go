package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func readFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func readURL(name string) ([]byte, error) {
	return nil, nil
}

const defaultTemplate = "{{ .Sha384 }}"

// for externally served files for SRI
const autoExternalTemplate = `
{{- if eq $.Ext ".css" -}}
<link rel="stylesheet" src="{{ $.Source }}" integrity="{{ .Sha384 }}" crossorigin="anonymous"></link>
{{- else -}}
<script src="{{ $.Source }}"{{ $.Async }} integrity="{{ .Sha384 }}" crossorigin="anonymous"></script>
{{- end -}}
`

// for locally served files with Cache Buster
const autoLocalTemplate = `
{{- if eq $.Ext ".css" -}}
<link rel="stylesheet" src="{{ $.Source }}?v={{ $.Checksum }}"></link>
{{- else -}}
<script src="{{ $.Source }}?v={{ $.Checksum }}"{{ $.Async }}></script>
{{- end }}
`

type TData struct {
	Source   string
	Base     string
	Ext      string
	Sha384   string
	Sha512   string
	Checksum string
	Async    string
}

func main() {
	//templateName := flag.String("t", "auto", "template")
	attrSrc := flag.String("src", "", "src name")
	attrAsync := flag.Bool("async", false, "use async tag in javascript")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("No filename specified")
	}
	name := args[0]
	if *attrSrc == "" {
		*attrSrc = name
	}
	var raw []byte
	var err error

	// read file
	raw, err = readFile(name)
	if err != nil {
		log.Fatalf("fail for %q: %s", name, err)
	}

	async := ""
	if *attrAsync {
		// note leading space
		async = " async"
	}

	tob64 := base64.StdEncoding.EncodeToString
	tmp384 := sha512.Sum384(raw)
	raw384 := []byte(tmp384[:])
	tmp512 := sha512.Sum512(raw)
	raw512 := []byte(tmp512[:])
	pagedata := TData{
		Source:   *attrSrc,
		Base:     filepath.Base(*attrSrc),
		Ext:      filepath.Ext(*attrSrc),
		Sha384:   tob64(raw384),
		Sha512:   tob64(raw512),
		Checksum: hex.EncodeToString(raw512[:6]),
		Async:    async,
	}

	tmpl, err := template.New("render").Parse(autoLocalTemplate)
	if err != nil {
		log.Fatalf("unable to parse template: %s", err)
	}
	err = tmpl.Execute(os.Stdout, pagedata)
	if err != nil {
		log.Fatalf("unable to render template: %s", err)
	}
}
