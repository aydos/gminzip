package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	//"io"
	"io/ioutil"
	//"log"
	//"net/url"
	"os"
	//"os/signal"
	"path"
	"path/filepath"
	"regexp"
	//"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/pflag"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

var mimetypes = map[string]string{
	"css":  "text/css",
	"htm":  "text/html",
	"html": "text/html",
	"js":   "text/javascript",
	"json": "application/json",
	"svg":  "image/svg+xml",
	"xml":  "text/xml",
}

type task struct {
	name string
	mime string
	ext  string
	min  bool
	zip  bool
}

var (
	m     *minify.M
	tasks []task
	//recursive bool
	//delete bool

	minextsall = []string{"css", "htm", "html", "js", "json", "svg", "xml"}
	zipextsall = false
	minexts    []string
	zipexts    []string
	size       = 0
)

func main() {

	minfiles := ""
	zipfiles := ""

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gminzip [options] inputs\n\nOptions:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories\n\n")
	}
	//pflag.StringVarP(&output, "output", "o", "", "Output file or directory (must have trailing slash)")
	pflag.StringVarP(&minfiles, "min", "m", "", "Files to minify (ex: -m css,html,js) (default: css,htm,html,js,json,svg,xml)")
	pflag.StringVarP(&zipfiles, "zip", "z", "", "Files to gzip (ex: -z jpg,js) (ex: -z all) (default: min option)")
	pflag.IntVarP(&size, "size", "s", 0, "Min file size in bytes for gzip (default: 0)")
	//pflag.BoolVarP(&recursive, "recursive", "r", true, "Recursively gminzip directories (true by default)")
	//pflag.BoolVarP(&delete, "delete", "", false, "Delete the original file")

	pflag.Parse()
	inputs := pflag.Args()

	if len(inputs) == 0 {
		pflag.Usage()
		return
	}

	// clearify minimize file types
	if minfiles == "" {
		minexts = minextsall
	} else {
		minexts = strings.Split(minfiles, ",")
		for i, ext := range minexts {
			if !contains(minextsall, ext) {
				minexts[i] = "" // delete unsupported type
			}
		}
	}

	// clearify gzip file types
	if zipfiles == "" {
		zipexts = minexts
	} else if zipfiles == "all" {
		zipextsall = true
		zipexts = []string{""}
	} else {
		zipexts = strings.Split(zipfiles, ",")
	}

	//fmt.Printf("minexts: %v\n\n", minexts)
	//fmt.Printf("zipexts: %v\n\n", zipexts)
	for _, input := range inputs {
		_ = filepath.Walk(input, visitfiles)
		//fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	m = minify.New()
	htmlMinifier := &html.Minifier{KeepDefaultAttrVals: false, KeepWhitespace: false, KeepDocumentTags: true}
	xmlMinifier := &xml.Minifier{KeepWhitespace: false}
	m.Add("text/html", htmlMinifier)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), xmlMinifier)

	var fails int32
	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(t task) {
			defer wg.Done()
			fmt.Printf("%#v\n", t)
			if ok := gminzip(t); !ok {
				atomic.AddInt32(&fails, 1)
			}
		}(t)
	}
	wg.Wait()

	if fails > 0 {
		os.Exit(1)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func visitfiles(p string, f os.FileInfo, err error) error {
	if f == nil {
		return nil
	}
	if len(f.Name()) == 0 || f.Name()[0] == '.' {
		return nil
	}
	if f.Mode().IsDir() { // directory
	}
	if f.Mode().IsRegular() { // file
		t := task{}
		t.name = p
		t.ext = path.Ext(f.Name())
		if len(t.ext) > 0 {
			t.ext = t.ext[1:]
			if contains(minexts, t.ext) {
				t.min = true
				t.mime = mimetypes[t.ext]
			}
			if zipextsall || (contains(zipexts, t.ext) && f.Size() > int64(size)) {
				t.zip = true
			}
		}
		if t.min || t.zip {
			tasks = append(tasks, t)
		}
	}
	return nil
}

func gminzip(t task) bool {

	success := true

	if t.min {
		mi := t.name
		mo := mi + ".bak"
		fi, _ := os.Open(mi)
		defer fi.Close()
		fo, _ := os.Create(mo)
		defer fo.Close()
		r := bufio.NewReader(fi)
		w := bufio.NewWriter(fo)
		_ = m.Minify(t.mime, w, r)
		w.Flush()
		_ = os.Remove(mi)
		_ = os.Rename(mo, mi)
	}

	if t.zip {
		zi := t.name
		zo := zi + ".gz"
		fi, _ := os.Open(zi)
		defer fi.Close()
		fo, _ := os.Create(zo)
		defer fo.Close()
		r := bufio.NewReader(fi)
		c, _ := ioutil.ReadAll(r)
		w, _ := gzip.NewWriterLevel(fo, gzip.BestCompression)
		defer w.Close()
		w.Write(c)
	}

	return success
}
