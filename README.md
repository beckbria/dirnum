# dirnum

Dirnum is a utility script designed to keep directories of image files well-numbered.  It validates that file names correspond to the following schema:

* Major version numbers are 4 digits (0000.jpg, 0001.jpg, etc.)
* Minor version numbers are optional and most start from 0 (0000-0.png, 0000-1.gif, etc.)
* Text tags are allowed at the end of files (0000-foo.jpg, 0001-0-bar.jpg)
* All major and minor version numbers appear in strictly increasing order with no gaps

If any divergence from the schema is found, the tool prints errors.  It is also capable of automatically fixing some basic mistakes such as using underscores instead of hyphens.

