/*
Package lzmadec implements extracting files from .7z archives. It requires
7z executable to be installed.

Short tutorial: http://blog.kowalczyk.info/article/g/Extracting-files-from-7z-archives-in-Go.html.
*/
package lzmadec

/*
//.7z ArchLinux version: https://git.archlinux.org/svntogit/packages.git/tree/trunk?h=packages/p7zip
Path = thefilename.txt
Size = 0
Packed Size = 0
Modified = 2019-08-28 16:38:07
Created = 2019-08-28 16:38:07
Accessed = 2019-08-28 16:37:54
Attributes = D_ drwxr-xr-x
CRC =
Encrypted = -
Method =
Block =
*/

/*
//zip, apk: 15 lines

Path = apkg-version
Folder = -
Size = 4
Packed Size = 4
Modified = 2018-04-05 23:11:28
Created =
Accessed =
Attributes = _ -rw-r--r--
Encrypted = -
Comment =
CRC = 76265B31
Method = Store
Host OS = Unix
Version = 20
Volume Index = 0
 */

/*
//.rar 17 lines
Path = example.txt
Folder = +
Size = 0
Packed Size = 0
Modified = 2017-05-10 22:49:26
Created =
Accessed =
Attributes = D
Encrypted = -
Solid = -
Commented = -
Split Before = -
Split After = -
CRC = 00000000
Host OS = Win32
Method = m0
Version = 20
 */

/*
//.rar 21 lines, rar 5.7.1
Path = this-is-dir
Folder = +
Size = 0
Packed Size = 0
Modified =
Created =
Accessed =
Attributes = D drwxr-xr-x
Alternate Stream = -
Encrypted = -
Solid = -
Split Before = -
Split After = -
CRC = 00000000
Host OS = Unix
Method = m0
Symbolic Link =
Hard Link =
Copy Link =
Checksum =
NT Security =
*/

/*
windows version 7z
zip: 17 lines
"C:\Program Files\7-Zip\7z.exe" l -slt  chromedriver_win32.zip

7-Zip 19.00 (x64) : Copyright (c) 1999-2018 Igor Pavlov : 2019-02-21

Scanning the drive for archives:
1 file, 4599757 bytes (4492 KiB)

Listing archive: chromedriver_win32.zip

--
Path = chromedriver_win32.zip
Type = zip
Physical Size = 4599757

----------
Path = chromedriver.exe
Folder = -
Size = 8713728
Packed Size = 4599627
Modified = 2018-12-10 14:54:46
Created =
Accessed =
Attributes =  -rwxrwxrwx
Encrypted = -
Comment =
CRC = 4A032CBD
Method = Deflate
Characteristics =
Host OS = FAT
Version = 20
Volume Index = 0
Offset = 0
 */

/*
rar: 23 lines
"C:\Program Files\7-Zip\7z.exe" l -slt IDM.xxxxx.rar

7-Zip 19.00 (x64) : Copyright (c) 1999-2018 Igor Pavlov : 2019-02-21

Path = IDM.xxxxxxxxx-AoRE.exe
Folder = -
Size = 39424
Packed Size = 36895
Modified = 2018-11-17 11:32:04
Created =
Accessed =
Attributes = A
Alternate Stream = -
Encrypted = -
Solid = -
Split Before = -
Split After = -
CRC = 4D1DA63A
Host OS = Windows
Method = m5:17
Characteristics = CRC Time:M
Symbolic Link =
Hard Link =
Copy Link =
Volume Index =
Checksum =
NT Security =
 */