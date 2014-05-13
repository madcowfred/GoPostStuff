GoPostStuff
===========

GoPostStuff is a simple client for posting binaries to Usenet, written in Go. If you've 
seen/used [newsmangler] [1], imagine that but faster (and maybe better one day).

  [1]: https://github.com/madcowfred/newsmangler/ "newsmangler"

Features
--------
* Multiple server support with multiple connections per server.
* Native TLS support so you don't need to use stunnel or equivalent frippery.
* Fast: a basic Linode VPS can push *250Mbit* of TLS-encrypted data while using 50-60%
  of a single CPU (Intel(R) Xeon(R) CPU E5-2680 v2 @ 2.80GHz).


Requirements
------------
* A working [Go installation] [2]
* A Usenet server that allows posting

  [2]: http://golang.org/doc/install  "Getting Started - The Go Programming Language"

Installation
------------
1. Initalise a directory to store Go files:

        mkdir ~/go
        export GOPATH="~/go"

1.  Get and install GoPostStuff - this will make a ~/go/bin/GoPostStuff binary:

        go get github.com/madcowfred/GoPostStuff
        go install github.com/madcowfred/GoPostStuff

3. Copy sample.conf to ~/.gopoststuff.conf and edit the options as appropriate.

        cp sample.conf ~/.gopoststuff.conf
        vim ~/.gopoststuff.conf

4. Run GoPostStuff!

Usage
-----

``gopoststuff [-c "CONFIG"] [-d] [-g "GROUP"] [-s "SUBJECT"] [-v] file1 file2 ... fileN``

* -c "CONFIG": Use an alternate configuration file.
* -d: Use directory posting mode. Each fileN argument _must_ be a directory. All files in each
  directory will be posted using the _directory name_ as the subject.
* -g "GROUP": Post to GROUP instead of the global/DefaultGroup config option.
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
