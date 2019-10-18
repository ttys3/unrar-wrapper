package lzmadec

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	timeLayout = "2006-01-02 15:04:05"
)

var (
	// Err7zNotAvailable is returned if 7z executable is not available
	Err7zNotAvailable = errors.New("7z executable not available")

	// ErrNoEntries is returned if the archive has no files
	ErrNoEntries = errors.New("no entries in 7z file")

	errUnexpectedLines = errors.New("unexpected number of lines")

	mu                 sync.Mutex
	p7zPath string
)

// Archive describes a single .7z archive
type Archive struct {
	Path     string
	Entries  []Entry
	password *string
}

// Entry describes a single file inside .7z,.rar,.zip archive
type Entry struct {
	Path            string
	Size            int64
	PackedSize      int // -1 means "size unknown"
	Modified        time.Time
	Created         time.Time //20190828 added, ArchLinux version has this https://git.archlinux.org/svntogit/packages.git/tree/trunk?h=packages/p7zip
	Accessed        time.Time //20190828 added, ArchLinux version has this
	Attributes      string
	CRC             string
	Encrypted       string
	Method          string
	Block           int
	Comment         string //zip only
	VolumeIndex     int //zip only
	Characteristics string //zip only
	Offset			int //zip only
	Solid           string //rar only
	Commented       string //rar lagecy only
	SplitBefore     string //rar only
	SplitAfter      string //rar only
	AlternateStream string //rar only, rar 5.7.1
	SymbolicLink    string //rar only, rar 5.7.1
	HardLink        string //rar only, rar 5.7.1
	CopyLink        string //rar only, rar 5.7.1
	Checksum        string //rar only, rar 5.7.1
	NTSecurity      string //rar only, rar 5.7.1
	Folder          string //zip, rar
	HostOS          string //zip, rar
	Version         string //zip, rar lagecy
}

func detect7zCached() error {
	mu.Lock()
	defer mu.Unlock()
	if p7zPath == "" {
		if p, err := exec.LookPath("7z"); err == nil {
			p7zPath = p
		}
	}
	if p7zPath != "" {
		// checked and present
		return nil
	}
	// checked and not present
	return Err7zNotAvailable
}

/*
----------
Path = Badges.xml
Size = 4065633
Packed Size = 18990516
Modified = 2015-03-09 14:30:49
Attributes = ....A
CRC = 2C468F32
Encrypted = -
Method = BZip2
Block = 0
*/
func advanceToFirstEntry(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		s := scanner.Text()
		if s == "----------" {
			return nil
		}
	}
	err := scanner.Err()
	if err == nil {
		err = ErrNoEntries
	}
	return err
}

func getEntryLines(scanner *bufio.Scanner) ([]string, error) {
	var res []string
	for scanner.Scan() {
		s := scanner.Text()
		s = strings.TrimSpace(s)
		if s == "" {
			break
		}
		res = append(res, s)
	}
	err := scanner.Err()
	if err != nil {
		return nil, err
	}
	//.iso may have 5 or 6, .7z may have 9 or 11, .zip may have 15, .rar may have 17, 21 or 23
	// len(res) == 6 || len(res) == 9 || len(res) == 11 || len(res) == 15 || len(res) == 17 || len(res) == 21 || len(res) == 23 || len(res) == 0
	if (len(res) >= 5 && len(res) <= 23) || len(res) == 0 {
		return res, nil
	}
	fmt.Printf("err: has lines: %d", len(res))
	return nil, errUnexpectedLines
}

func parseEntryLines(lines []string) (Entry, error) {
	var e Entry
	var err error
	for _, s := range lines {
		parts := strings.SplitN(s, " =", 2)
		if len(parts) != 2 {
			return e, fmt.Errorf("unexpected line: '%s'", s)
		}
		name := strings.ToLower(parts[0])
		v := strings.TrimSpace(parts[1])
		if v == "" {
			v = "0"
		}
		switch name {
		case "path":
			e.Path = v
			if e.Path == "" {
				err = fmt.Errorf("Path field can not be empty")
			}
		case "size":
			e.Size, err = strconv.ParseInt(v, 10, 64)
		case "packed size":
			e.PackedSize = -1
			if v != "" {
				e.PackedSize, err = strconv.Atoi(v)
			}
		case "modified":
			e.Modified, _ = time.Parse(timeLayout, v)
		case "created":
			e.Created, _ = time.Parse(timeLayout, v)
		case "accessed":
			e.Accessed, _ = time.Parse(timeLayout, v)
		case "attributes":
			e.Attributes = v
			if e.Attributes == "" {
				err = fmt.Errorf("Attributes field can not be empty")
			}
		case "crc":
			e.CRC = v
		case "encrypted":
			e.Encrypted = v
		case "method":
			e.Method = v
		case "block":
			e.Block, err = strconv.Atoi(v)
		case "comment":
			e.Comment = v
		case "volume index":
			e.VolumeIndex, err = strconv.Atoi(v)
		case "solid":
			e.Solid = v
		case "commented":
			e.Commented = v
		case "split before":
			e.SplitBefore = v
		case "split after":
			e.SplitAfter = v
		case "folder":
			e.Folder = v
		case "host os":
			e.HostOS = v
		case "version":
			e.Version = v
		//rar 5.7.1
		case "alternate stream":
			e.AlternateStream = v
		case "symbolic link":
			e.SymbolicLink = v
		case "hard link":
			e.HardLink = v
		case "copy link":
			e.CopyLink = v
		case "checksum":
			e.Checksum = v
		case "nt security":
			e.NTSecurity = v
		case "characteristics":
			e.Characteristics = v
		case "offset":
			e.Offset, err = strconv.Atoi(v)
		default:
			err = fmt.Errorf("unexpected entry line '%s'", name)
		}
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func parse7zListOutput(d []byte) ([]Entry, error) {
	var res []Entry
	r := bytes.NewBuffer(d)
	scanner := bufio.NewScanner(r)
	err := advanceToFirstEntry(scanner)
	if err != nil {
		return nil, err
	}
	for {
		lines, err := getEntryLines(scanner)
		if err != nil {
			return nil, err
		}
		if len(lines) == 0 {
			// last entry
			break
		}
		e, err := parseEntryLines(lines)
		if err != nil {
			return nil, err
		}

		// fixup empty Attributes for .iso
		if e.Attributes == "" && e.Folder != "" {
			if e.Folder == "+" {
				e.Attributes = "D"
			} else if e.Folder == "-" {
				e.Attributes = "A"
			}
		}

		// check e.Attributes
		if e.Attributes == "" {
			err = fmt.Errorf("Attributes field can not be empty")
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func Set7zPath(path string) {
    p7zPath = path
}

func NewArchive(path string) (*Archive, error) {
	return newArchive(path, nil)
}

func NewEncryptedArchive(path string, password string) (*Archive, error) {
	return newArchive(path, &password)
}

// NewArchive uses 7z to extract a list of files in .7z archive
func newArchive(path string, password *string) (*Archive, error) {
	err := detect7zCached()
	if err != nil {
		return nil, err
	}

	params := []string{"l", "-slt", "-sccUTF-8"}
	var tmpPassword string
	if password == nil || *password == "" {
		// 7z interactively asks for a password when an archive is encrypted
		// and no password has been supplied. But it has no problems when
		// a password has been supplied and the archive is not encrypted.
		// So if no password has been provided, use a non-sensical one to
		// prevent 7z from blocking on encrypted archives and instead fail
		tmpPassword = "                  "
	} else {
		tmpPassword = *password
	}
	params = append(params, fmt.Sprintf("-p%s", tmpPassword))
	params = append(params, path)

	/*
	here we must use fullpath 7z
	on QNAP nas
		7-Zip [64] 16.02 : Copyright (c) 1999-2016 Igor Pavlov : 2016-05-21
		p7zip Version 16.02 (locale=en_US.UTF-8,Utf16=on,HugeFiles=on,64 bits,4 CPUs x64)
	if we do not use fullpath to exec 7z, it will result in error:
		Can't load './7z.dll' (./7z.so: cannot open shared object file: No such file or directory)
		ERROR:
			7-Zip cannot find the code that works with archives.
	I finally understand why most Linux use a shell file named 7z and its contents is just:
		#! /bin/sh
		"/usr/lib/p7zip/7z" "$@"
	 */
	cmd := exec.Command(p7zPath, params...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("err: %s, out: %s", err.Error(), out)
	}
	entries, err := parse7zListOutput(out)
	if err != nil {
		return nil, err
	}
	return &Archive{
		Path:     path,
		Entries:  entries,
		password: &tmpPassword,
	}, nil
}

type readCloser struct {
	rc  io.ReadCloser
	cmd *exec.Cmd
}

func (rc *readCloser) Read(p []byte) (int, error) {
	return rc.rc.Read(p)
}

func (rc *readCloser) Close() error {
	// if we want to finish before reading all the data, we need to Close()
	// stdout pipe, or else rc.cmd.Wait() will hang.
	// if it's already closed then Close() returns 'invalid argument',
	// which we can ignore
	rc.rc.Close()
	return rc.cmd.Wait()
}

// GetFileReader returns a reader for reading a given file
func (a *Archive) GetFileReader(name string) (io.ReadCloser, error) {
	found := false
	for _, e := range a.Entries {
		if e.Path == name {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("file not in the archive")
	}

	params := []string{"x", "-so"}
	if a.password != nil {
		params = append(params, fmt.Sprintf("-p%s", *a.password))
	}
	params = append(params, a.Path, name)

	cmd := exec.Command(p7zPath, params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	rc := &readCloser{
		rc:  stdout,
		cmd: cmd,
	}
	err = cmd.Start()
	if err != nil {
		stdout.Close()
		return nil, err
	}
	return rc, nil
}

// ExtractToWriter writes the content of a given file inside the archive to dst
func (a *Archive) ExtractToWriter(dst io.Writer, name string) error {
	r, err := a.GetFileReader(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, r)
	err2 := r.Close()
	if err != nil {
		return err
	}
	return err2
}

// ExtractToFile extracts a given file from the archive to a file on disk
func (a *Archive) ExtractToFile(dstPath string, name string) error {
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return a.ExtractToWriter(f, name)
}
