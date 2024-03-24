#!/bin/sh

zip -r example.zip this-is-dir

#curl -LZ https://www.rarlab.com/rar/rarlinux-x64-700.tar.gz -o/tmp/rarlinux.tar.gz
#tar xvzf /tmp/rarlinux.tar.gz -C/tmp/
#mv /tmp/rarlinux*/rar /usr/local/bin/rar
/usr/local/bin/rar a -r example.rar this-is-dir
/usr/local/bin/rar a -r -phelloworld encrypted.rar this-is-dir
/usr/local/bin/rar a -r -hphelloworld encrypted.inc.headers.rar this-is-dir

