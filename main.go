package main

import (
	"bufio"
	//"compress/brotli"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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
	dst  string
	min  bool
	zip  bool
}

var (
	m          *minify.M
	tasks      []task
	delete     bool
	silent     bool
	brotli     bool
	clean      bool
	minextsall = []string{"css", "htm", "html", "js", "json", "svg", "xml"}
	zipextsall bool
	minexts    []string
	zipexts    []string
	minsize    int64
	maxsize    int64
	list       bool
	listexts   map[string]int
	mincount   int
	zipcount   int
)

func main() {
	minfiles := ""
	zipfiles := ""
	listexts = make(map[string]int)

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gminzip [options] inputs\n\nOptions:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories\n\n")
		fmt.Fprintf(os.Stderr, "Visit https://github.com/aydos/gminzip for more example.\n\n")
	}
	pflag.StringVarP(&minfiles, "min", "m", "", "Files to minify (ex -m html,css) (supported: css,htm,html,js,json,svg,xml)")
	pflag.StringVarP(&zipfiles, "zip", "z", "", "Files to zip (ex: -z html,js,swf,jpg) (ex: -z all)")
	pflag.Int64VarP(&minsize, "size", "s", 0, "Minimum file size in bytes for zip (default: 0)")
	pflag.Int64VarP(&maxsize, "maxsize", "x", 0, "Maximum file size in bytes for minify and zip")
	pflag.BoolVarP(&list, "list", "l", false, "List all file extensions and count files in inputs")
	pflag.BoolVarP(&delete, "delete", "", false, "Delete the original file after zip")
	pflag.BoolVarP(&silent, "silent", "", false, "Do not display info, but show the errors")
	//pflag.BoolVarP(&brotli, "brotli", "b", false, "Use brotli to zip instead of gzip")
	pflag.BoolVarP(&clean, "clean", "", false, "Delete the ziped files (.gz, .br) before process")

	pflag.Parse()
	inputs := pflag.Args()

	if len(inputs) == 0 {
		pflag.Usage()
		return
	}

	// files to minify
	minexts = strings.Split(minfiles, ",")
	for i, ext := range minexts {
		if !contains(minextsall, ext) {
			minexts[i] = "" // delete unsupported type
		}
	}

	// files to zip
	if zipfiles == "all" {
		zipextsall = true
		zipexts = []string{""}
	} else {
		zipexts = strings.Split(zipfiles, ",")
	}

	// visit each input and build the task list
	for _, input := range inputs {
		_ = filepath.Walk(input, visitfiles)
	}

	//minifier
	m = minify.New()
	htmlMinifier := &html.Minifier{KeepDefaultAttrVals: false, KeepWhitespace: false, KeepDocumentTags: true}
	xmlMinifier := &xml.Minifier{KeepWhitespace: false}
	m.Add("text/html", htmlMinifier)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), xmlMinifier)

	// processing tasks
	fails := 0
	concurrency := 16
	sem := make(chan bool, concurrency)
	for _, t := range tasks {
		sem <- true
		go func(t task) {
			defer func() { <-sem }()
			if !silent {
				info := ""
				if t.min {
					info += "M"
				} else {
					info += " "
				}
				if t.zip {
					info += "Z "
				} else {
					info += "  "
				}
				info += t.name
				fmt.Println(info)
			}
			if ok := gminzip(t); !ok {
				fails++
			}
		}(t)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	if !silent && !(list && len(tasks) == 0) {
		fmt.Printf("%6d files were processed\n", len(tasks))
		fmt.Printf("%6d files were minified\n", mincount)
		fmt.Printf("%6d files were zipped\n", zipcount)
	}

	// show file extensions and file counts
	if list {
		var exts []string
		for e := range listexts {
			exts = append(exts, e)
		}
		sort.Strings(exts)
		for _, ext := range exts {
			fmt.Printf("%6d %s\n", listexts[ext], ext)
		}
	}

	if fails > 0 {
		fmt.Println("\nCAUTION: There are ERRORs...")
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
			// count files via extensions
			if list {
				listexts[t.ext]++
			}
			// dont minify or zip zipped files
			if t.ext == "gz" || t.ext == "br" {
				if clean { // but delete them if you want
					err = os.Remove(t.name)
					if err != nil {
						fmt.Println("ERROR : Cant delete zipped file", t.name)
					}
				}
				return nil
			}
			// if maxsize defined
			if maxsize > 0 {
				if f.Size() > maxsize {
					return nil
				}
			}
			// check for minify
			if contains(minexts, t.ext) {
				t.min = true
				t.mime = mimetypes[t.ext]
			}
			// check for zip
			if contains(zipexts, t.ext) || zipextsall {
				if f.Size() > minsize {
					t.zip = true
				}
			}
		}
		if t.min || t.zip {
			tasks = append(tasks, t)
		}
	}
	return nil
}

func gminzip(t task) bool {

	if t.min {
		mi := t.name
		mo := mi + ".bak"
		fi, err := os.Open(mi)
		if err != nil {
			fmt.Println("ERROR : Cant open", mi)
			return false
		}
		defer fi.Close()
		fo, err := os.Create(mo)
		if err != nil {
			fmt.Println("ERROR : Cant create", mo)
			return false
		}
		defer fo.Close()
		r := bufio.NewReader(fi)
		w := bufio.NewWriter(fo)
		err = m.Minify(t.mime, w, r)
		if err != nil {
			fmt.Println("ERROR : Cant minify", mi)
			return false
		}
		w.Flush()
		err = os.Remove(mi)
		if err != nil {
			fmt.Println("ERROR : Cant delete original file", mi)
			return false
		}
		err = os.Rename(mo, mi)
		if err != nil {
			fmt.Println("ERROR : Cant rename ", mo)
			return false
		}
		mincount++
	}

	if t.zip {
		zi := t.name
		zo := zi + ".gz"
		fi, err := os.Open(zi)
		defer fi.Close()
		if err != nil {
			fmt.Println("ERROR : Cant open", zi)
			return false
		}
		fo, err := os.Create(zo)
		if err != nil {
			fmt.Println("ERROR : Cant create", zo)
			return false
		}
		defer fo.Close()
		if !brotli { // gzip
			gz, err := gzip.NewWriterLevel(fo, gzip.BestCompression)
			if err != nil {
				fmt.Println("ERROR : Cant gzip file", zi)
				return false
			}
			defer gz.Close()
			_, err = io.Copy(gz, fi)
			if err != nil {
				fmt.Println("ERROR : Cant zip file", zi)
				return false
			}
		} else {
			//br, err := brotli.NewWriter(fo)
		}

		zipcount++
		if delete {
			err = os.Remove(zi)
			if err != nil {
				fmt.Println("ERROR : Cant delete after zip", zi)
				return false
			}
		}
	}
	return true
}
