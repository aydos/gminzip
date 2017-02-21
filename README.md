# GMinZip
Statically minify and gzip your web site files (html, css, js, jpeg, etc...)

This project relies on Taco de Wolff's minify project: https://github.com/tdewolff/minify  
You may also want to look at minify-cmd project: https://github.com/tdewolff/minify/tree/master/cmd/minify

## Usage

	Usage: gminzip [options] inputs

	Options:
	  -m, --min
			Files to minify (ex: -m css,html,js) (default: css,htm,html,js,json,svg,xml)
	  -z, --zip
			Files to zip (gzip) (ex: -z all) (default: copy of min option)
	  -s, --size
			Min file size in bytes for zip (default: 0)
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

	gminzip /var/www
		Minify and gzip all "css,htm,html,js,json,svg,xml" files in /var/www

	gminzip -m none -z all /var/www
		Gzip every file in /var/www with an extention (.jpeg, .swf, .html, ...) but do not minify

	gminzip -m css,html /var/www/site1 /var/www/site2 index.html
		Minify and gzip css and html files in /var/www/site1 and /var/www/site2 and index.html file in the current directory

	gminzip -m css,html -z json /var/www/site1
		Minify css and html files, gzip only json files in /var/www/site1

	gminzip -s 120 /var/www
		Minify all "css,htm,html,js,json,svg,xml" files and gzip the result if size larger than 120 bytes

	gminzip -l -m none /var/www
		List all extensions and file counts in /var/www/ but no minify, no gzip

	gminzip -l -m none -z js /var/www
		List all extensions and file counts in /var/www/ and gzip js files, but no minify

## Notes

* GMinZip is recursive by default
* Minifying a file overwrite the original
* If no -z file specified, -m option is used
* May have problems with large files (use maxsize option)

## TODO

* Add brotli support as soon as a native brotli package added to golang.  
  see issue: https://github.com/google/brotli/issues/182

