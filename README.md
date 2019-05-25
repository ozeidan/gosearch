gosearch
========
gosearch is an attempt to create the Windows file search tool Everything for Linux 5.1+, written in Go.

Existing file search tools for Linux haven't been satisfying since they didn't fulfill the following criteria:

* Super fast
* Finds all files on the system
* Requires few system resources
* Always up-to-date

gosearch aims to be the perfect file searching program for Linux and to provide blazing fast fuzzy/substring/prefix-searches on ALL files at ALL times.

What's different about gosearch?
--------------------------------
Other file search tools on Linux have failed to provide a fast and up to date index, since the Linux kernel has not provided the means to monitor file changes accross the whole filesystem. With [recent changes](https://lkml.org/lkml/2019/3/1/400 "fanotify patch") of the Linux kernel, that add filesystem monitoring for creations/deletions/moves to the fanotify system, it is possible for gosearch to keep a file index up to date in real time and using few system resources. These changes have made it into kernel version 5.1, which is required to use this program (or you can patch old kernel versions).

Performance
-----------
Since the filesystem change events sent by fanotify only report the parent directory of a created/deleted/moved file, the whole filesystem has to be indexed in a tree-like structure. This index is then used to compare the current directory contents to the last-known state and the created/deleted files are updated from the actual filename index.

The filename index, which provides the fast searches on filename across the filesystem, is built using a patricia trie. For this purpose, a memory-optimized version of go-patricia was created, which can be found [here](https://github.com/ozeidan/fuzzy-patricia/) .

Nevertheless, on my system gosearch uses 250mb of memory and most fuzzy/substring queries are processed in less then 100ms. Prefix queries are processed in a matter of microseconds. These benchmarks were conducted on ~1.1 million indexed files and ~130 thousand directories, which amount to ~250GB of data. The indexing, which has to be run once everytime the system restarts, takes roughly 6 seconds.
TODO: actual benchmarks

Upcoming Features
------------------
gosearch is still in early developement, but usable (need more testers). Things, that I would like to see implemented in gosearch include:

* Case-insensitive searching
* Fuzzy searching on whole file paths
* More performance optimizations/faster indexing
* Porperly deal with multiple drives, listen to mount events
* Integration with other tools (i.e. rofi)
* Maybe a GUI

Installation
============
**Important:** gosearch requires a kernel of version >= 5.1 or [this patch](https://lkml.org/lkml/2019/3/1/400) applied to your current kernel. I have not applied the patch to older kernel versions and don't know if it works or how hard it is.

Installing via Package Manager
------------------------------
A gosearch package is available for the following distributions:
* Arch Linux (AUR): [gosearch-git](https://aur.archlinux.org/packages/gosearch-git/)

More to come!

Installing manually
-------------------
To build the program you need a working Go installation (version > 1.11) and properly set `$GOPATH`.

Clone the repo and install by running

	make install

This command will build the server and client binaries, will put them in their appropriate directories. It will also install a systemd service and run it. To use the program, you can now run

	gosearch

If you don't want to install the systemd script, you can build the binaries by running

	make build

and start the server binary `gosearchServer` by hand/use whatever system you're using.
Contributions to support alternatives to systemd are appreciated!

Configuration
-------------
The server will create a configuration file at `/etc/gosearch/config`, the first time it is run. You should probably edit it to set some filters in there, so some useless directories are not indexed (e.g. .cache, /proc, /dev...).

Usage
=====
After the server is started and has indexed your files (takes a couple of seconds, depending on the amount of files on your system), you use the `gosearch` command send queries.

The default searching mode is substring searching (e.g. the query 'sea' will match the file 'gosearch'):

	gosearch [query]

For fuzzy searching, set the `-f` flag (e.g. the query 'grch' will match the file 'gosearch'):

	gosearch -f [query]

Prefix searching can be conducted by setting the `-p` flag, this is the fastes of the search options:

	gosearch -p [query]

To reverse the sorting order, the `-r` flag can be set, and sorting can be disabled by setting the `-nosort` flag.


Contributing
============
I am hoping for some contributions to this project. Please test the software and create plenty issues for its shortcommings. Any kinds of pull requests are always welcome. Hopefully, we can build a performant and stable tool together and I can stop writing in first person in this readme file. :)
