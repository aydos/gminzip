# GMinZip
Statically minify and gzip your web site files (html, css, js, jpeg, etc...)

## Usage

	Usage: gminzip [options] inputs

	Options:
	  -m, --min
			Files to minify (ex: -m css,html,js) (default: css,htm,html,js,json,svg,xml)
	  -z, --zip
			Files to gzip (ex: -z jpg,js) (ex: -z all) (default: min option)
	  -s, --size 
			Min file size in bytes for gzip (default: 0)

	Inputs:
	  Files or directories

## Examples

	gminzip /var/www
		Minify and gzip all "css,htm,html,js,json,svg,xml" files in /var/www

	gminzip -m none -z all /var/www
		Gzip every file in /var/www with an extention (.jpeg, .swf, .html, ...) but do not minify

	gminzip -m css,html /var/www/site1 /var/www/site2
		Minify and gzip css and html files in /var/www/site1 and /var/www/site2

	gminzip -m css,html -z json /var/www/site1
		Minify css and html files, gzip only json files in /var/www/site1

	gminzip -s 120 /var/www
		Minify all "css,htm,html,js,json,svg,xml" files and gzip file size larger than 120 bytes in /var/www
