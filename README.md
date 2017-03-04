# GMinZip
Statically minify and gzip your web site files (html, css, js, jpeg, etc...)

This project relies on Taco de Wolff's minify project: https://github.com/tdewolff/minify  
You may also want to look at minify-cmd project: https://github.com/tdewolff/minify/tree/master/cmd/minify

## Usage

	Usage: gminzip [options] inputs

	Options:
	  -m, --min
			Files to minify (ex -m html,css) (supported: css,htm,html,js,json,svg,xml)
	  -z, --zip
			Files to zip (ex: -z html,js,swf,jpg) (ex: -z all)
	  -s, --size
			Minimum file size in bytes for zip (default: 0)
	  -x, --maxsize
			Maximum file size in bytes for minify and zip
	  -l, --list
			List all file extensions and count files in inputs
	  --delete
			Delete the original file after zip
	  --silent
			Do not display info, but show the errors
	  --clean
			Delete the ziped files (.gz, .br) before process

	Inputs:
	  Files or directories

## Examples

Gminzip works with extentions and file extentions must be specified with -m and/or -z options

Do nothing. You must specify the extentions with -m and/or -z options:

    gminzip index.html

Minify and zip all "html" files in current working directory:

    gminzip -m html -z html .

Minify all "css,html" files in /var/www:

	gminzip -m css,html /var/www

Zip every file with an extention ("all" must be alone) but do not minify in /var/www:

	gminzip -z all /var/www

Minify and zip "css,html" files in /var/www/site1 and /var/www/site2 and index.html file in the current directory:

	gminzip -m css,html -z css,html /var/www/site1 /var/www/site2 index.html

Minify "css,html" files, but zip only json files (and *.all files, not every file) in /var/www/site:

	gminzip -m css,html -z all,json /var/www/site

Minify all "svg,html,xml" files (not "jpg"s) and gzip every file with extention if file size is larger than 120 bytes:

	gminzip -m svg,html,xml,jpg -m all -s 120 /var/www

List all extensions and file counts and zip "js" files:

	gminzip -l -z js /var/www

## Notes

* ALWAYS take backups
* Only "css,htm,html,js,json,svg,xml" files can minified
* Minifying a file overwrite the original
* To minify the minified files may cause problems
* GMinZip is recursive by default
* No Default for -m and -z options

## TODO

* Add brotli support as soon as a native brotli package added to golang.  
  see issue: https://github.com/google/brotli/issues/182
