GoPostStuff
===========

GoPostStuff is a simple client for posting binaries to Usenet. It's pretty much an
attempt at a modernised version of [newsmangler] [1].

  [1]: https://github.com/madcowfred/newsmangler/ "newsmangler"

Features
--------
* Multiple server support with multiple connections per server.
* Encrypted connection support if you like that sort of thing.
* Maxes a 100Mbit connection with relatively light CPU usage - encrypted connections will
    use around twice as much CPU (~18% vs ~40% of a single core on my test machine).

Requirements
------------
* A working [Go installation] [2]
* A Usenet server that allows posting

  [2]: http://golang.org/doc/install  "Getting Started - The Go Programming Language"

Installation
------------
1.  Get the source:

        git clone git://github.com/madcowfred/GoPostStuff.git

2. We'll be lazy and build the app in the current directory:

        export GOPATH=`pwd`
        go get
        go build

3. Copy the binary to a bin directory somewhere:

        cp gopoststuff ~/bin
        sudo cp gopoststuff /usr/local/bin

4. Copy sample.conf to ~/.gopoststuff.conf and edit the options as appropriate.

        cp sample.conf ~/.gopoststuff.conf
        vim ~/.gopoststuff.conf

Usage
-----

``gopoststuff [-c "CONFIG"] [-d] [-s "SUBJECT"] [-v] file1 file2 ... fileN``

* -c "CONFIG": Use an alternate configuration file.
* -d: Use directory posting mode. Each fileN argument _must_ be a directory. All files in each
  directory will be posted using the _directory name_ as the subject.
* -s "SUBJECT": Use subject posting mode. All files will be posted using SUBJECT as the subject.
  Directories supplied as arguments are always recursed into.
* -v: Verbose mode. This will spam a lot of extra debug information.

Example
-------
Let's say you have some files that you would like to post:

* Cool Files/
    + cool.rar
    + cool.r00
    + cool.r01
    + cool.sfv

You can post it with the subject "Cool Files" like so:

``gopoststuff -d "Cool Files"``

or with a different subject like so:

``gopoststuff -s "This is a different subject" "Cool Files"``
