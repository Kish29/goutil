package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gookit/color"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/strutil"
)

var (
	hidden = []string{
		"netutil",
		"numutil",
		"internal",
	}
	nameMap = map[string]string{
		"arr":     "array/Slice",
		"str":     "string",
		"sys":     "system",
		"math":    "math/Number",
		"fs":      "fileSystem",
		"fmt":     "formatting",
		"test":    "testing",
		"dump":    "dump",
		"structs": "struct",
		"json":    "JSON",
		"cli":     "CLI",
		"env":     "ENV",
		"std":     "standard",
	}

	pkgDesc = map[string]map[string]string{
		"en": {
			"arr": "Package arrutil provides some util functions for array, slice",
		},
		"zh-CN": {
			"arr": "`arrutil` 包提供一些辅助函数，用于数组、切片处理",
		},
	}

	allowLang = map[string]int{
		"en":    1,
		"zh-CN": 1,
	}
)

type genOptsSt struct {
	lang     string
	output   string
	template string
	tplDir   string
}

func (o genOptsSt) tplFilename() string {
	if o.lang == "en" {
		return "README.md.tpl"
	}

	return fmt.Sprintf("README.%s.md.tpl", o.lang)
}

func (o genOptsSt) tplFilepath(givePath string) string {
	if givePath != "" {
		return path.Join(o.tplDir, givePath)
	}
	return path.Join(o.tplDir, o.tplFilename())
}

var (
	genOpts = genOptsSt{}
	// collected sub package names.
	// short name => full name.
	pkgNames = make(map[string]string, 16)

	partDocTplS = "part-%s-s%s.md"
	partDocTplE = "part-%s%s.md"
)

func bindingFlags() {
	flag.StringVar(&genOpts.lang, "l", "en", "package desc message language. allow: en, zh-CN")
	flag.StringVar(&genOpts.output,
		"o",
		"./metadata.log",
		"the result output file. if is 'stdout', will direct print it.",
	)
	flag.StringVar(&genOpts.tplDir,
		"t",
		"./internal/template",
		"the template file dir, use for generate, will inject metadata to the template.\nsee ./internal/template/*.tpl",
	)

	flag.Usage = func() {
		color.Info.Println("Collect and dump all exported functions for goutil\n")

		color.Comment.Println("Options:")
		flag.PrintDefaults()

		color.Comment.Println("Example:")
		fmt.Println(`  go run ./internal/gendoc -o stdout
  go run ./internal/gendoc -o stdout -l zh-CN
  go run ./internal/gendoc -o README.md
  go run ./internal/gendoc -o README.zh-CN.md -l zh-CN`)
	}
}

// go run ./internal/gendoc -h
// go run ./internal/gendoc
func main() {
	bindingFlags()
	flag.Parse()

	ms, err := filepath.Glob("./*/*.go")
	goutil.PanicIfErr(err)

	var out io.Writer
	var toFile bool

	if genOpts.output == "stdout" {
		out = os.Stdout
	} else {
		toFile = true
		out, err = os.OpenFile(genOpts.output, os.O_CREATE|os.O_WRONLY, fsutil.DefaultFilePerm)
		goutil.PanicIfErr(err)

		// close after handle
		defer out.(*os.File).Close()
	}

	// want output by template file
	// var tplFile *os.File
	var tplBody []byte
	if genOpts.tplDir != "" {
		tplFile := genOpts.tplFilepath("")
		color.Info.Println("- read template file contents from", tplFile)
		tplBody = fsutil.MustReadFile(tplFile)
	}

	basePkg := "github.com/gookit/goutil"

	// collect functions
	buf := collectPgkFunc(ms, basePkg)

	// write to output
	if len(tplBody) > 0 {
		_, err = fmt.Fprint(out, strings.Replace(string(tplBody), "{{pgkFuncs}}", buf.String(), 1))
	} else {
		_, err = buf.WriteTo(out)
	}

	goutil.PanicIfErr(err)

	color.Cyanln("Collected packages:")
	dump.Clear(pkgNames)

	if toFile {
		color.Info.Println("OK. write result to the", genOpts.output)
	}
}

func collectPgkFunc(ms []string, basePkg string) *bytes.Buffer {
	var name, dirname string
	var pkgFuncs = make(map[string][]string)

	// match func
	reg := regexp.MustCompile(`func [A-Z]\w+\(.*\).*`)
	buf := new(bytes.Buffer)

	color.Info.Println("- find and collect exported functions...")
	for _, filename := range ms { // for each go file
		// "jsonutil/jsonutil_test.go"
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}

		// "sysutil/sysutil_windows.go"
		if strings.HasSuffix(filename, "_windows.go") {
			continue
		}

		idx := strings.IndexRune(filename, '/')
		dir := filename[:idx] // sub pkg name.

		if arrutil.StringsHas(hidden, dir) {
			continue
		}

		pkgPath := basePkg + "/" + dir
		pkgNames[dir] = pkgPath

		if ss, ok := pkgFuncs[pkgPath]; ok {
			pkgFuncs[pkgPath] = append(ss, "added")
		} else {
			if len(pkgFuncs) > 0 { // end of prev package.
				bufWriteln(buf, "```")

				// load prev sub-pkg doc file.
				bufWriteDoc(buf, partDocTplE, dirname)
			}

			dirname = dir
			name = dir
			if strings.HasSuffix(dir, "util") {
				name = dir[:len(dir)-4]
			}

			if setTitle, ok := nameMap[name]; ok {
				name = setTitle
			}

			// now: name is package name.
			bufWriteln(buf, "\n###", strutil.UpperFirst(name))
			bufWritef(buf, "\n> Package `%s`\n\n", pkgPath)
			pkgFuncs[pkgPath] = []string{"xx"}

			// load sub-pkg start doc file.
			bufWriteDoc(buf, partDocTplS, name)

			bufWriteln(buf, "```go")
		}

		// read contents
		text := fsutil.MustReadFile(filename)
		lines := reg.FindAllString(string(text), -1)

		if len(lines) > 0 {
			bufWriteln(buf, "// source at", filename)
			for _, line := range lines {
				bufWriteln(buf, strings.TrimRight(line, "{ "))
			}
		}
	}

	if len(pkgFuncs) > 0 {
		bufWriteln(buf, "```")
		// load last sub-pkg doc file.
		bufWriteDoc(buf, partDocTplE, dirname)
	}

	return buf
}

func bufWritef(buf *bytes.Buffer, f string, a ...interface{}) {
	_, _ = fmt.Fprintf(buf, f, a...)
}

func bufWriteln(buf *bytes.Buffer, a ...interface{}) {
	_, _ = fmt.Fprintln(buf, a...)
}

func bufWriteDoc(buf *bytes.Buffer, partNameTpl, pkgName string) {
	var lang string
	if genOpts.lang != "en" {
		lang = "." + genOpts.lang
	}

	filename := fmt.Sprintf(partNameTpl, pkgName, lang)

	if !doWriteDoc2buf(buf, filename) {
		// fallback use en docs
		filename = fmt.Sprintf(partNameTpl, pkgName, "")
		doWriteDoc2buf(buf, filename)
	}
}

func doWriteDoc2buf(buf *bytes.Buffer, filename string) bool {
	partFile := genOpts.tplDir + "/" + filename
	// color.Infoln("- try read part readme from", partFile)
	partBody := fsutil.ReadExistFile(partFile)

	if len(partBody) > 0 {
		color.Infoln("- find and inject sub-package doc:", filename)
		_, _ = fmt.Fprintln(buf, string(partBody))
		return true
	}

	return false
}
