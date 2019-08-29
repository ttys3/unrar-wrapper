#!/bin/sh

7z a -R example.7z this-is-dir
7z a -phelloworld -R encrypted.7z this-is-dir


zip -r example.zip this-is-dir
zip -r -P helloworld encrypted.zip this-is-dir

#wget https://rarlab.com/rar/rarlinux-x64-5.7.1.tar.gz -O/tmp/rarlinux-x64-5.7.1.tar.gz
#tar xvzf /tmp/rarlinux-x64-5.7.1.tar.gz -C/tmp/
#mv /tmp/rarlinux-x64-5.7.1/rar /usr/lib/rarlab
/usr/lib/rarlab/rar a -r example.rar this-is-dir
/usr/lib/rarlab/rar a -r -phelloworld encrypted.rar this-is-dir
/usr/lib/rarlab/rar a -r -hphelloworld encrypted.inc.headers.rar this-is-dir

