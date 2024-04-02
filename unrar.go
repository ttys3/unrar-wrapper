package unrarwrapper

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	timeLayout = "2006-01-02 15:04:05,000000000"
)

var (
	// ErrUnRARNotAvailable is returned if unrar executable is not available
	ErrUnRARNotAvailable = errors.New("unrar executable not available")

	// ErrNoEntries is returned if the archive has no files
	ErrNoEntries = errors.New("no entries in rar file")

	mu        sync.Mutex
	unrarPath string
)

// Archive describes a single .rar archive
type Archive struct {
	Path     string
	Entries  []Entry
	password *string
	maxCpuNum int
}

type EntryType string

const (
	EntryTypeFile      EntryType = "File"
	EntryTypeDirectory EntryType = "Directory"
)

// Entry describes a single file inside .rar,.zip archive
type Entry struct {
	Name            string
	Size            int64 // extracted size in bytes
	PackedSize      int64 // -1 means "size unknown"
	Ratio		   string // Ratio: 54%

	Modified        time.Time

	Attributes      string // Attributes: -rw-r--r--
	CRC             string
	HostOS          string // zip, rar
	Compression     string

	Flags       string // Flags: encrypted (directory does not have this field)

	Type       EntryType // Type: Directory or File
}

func detectUnRARCached() error {
	mu.Lock()
	defer mu.Unlock()
	if unrarPath == "" {
		if p, err := exec.LookPath("unrar"); err == nil {
			unrarPath = p
		}
	}
	if unrarPath != "" {
		// checked and present
		return nil
	}
	// checked and not present
	return ErrUnRARNotAvailable
}

const NormalSep = "\n"

/*
UNRAR 7.00 freeware      Copyright (c) 1993-2024 Alexander Roshal

Archive: myfile-header-enc.rar
Details: RAR 5, encrypted headers

*/
var detailsRe = regexp.MustCompile(`^Details: RAR(.+)$`)

func advanceToFirstEntry(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		s := scanner.Text()
		if detailsRe.MatchString(s) {
			scanner.Scan()
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
		if s == NormalSep {
			break
		}
		res = append(res, s)
	}
	err := scanner.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/*
        Name: rar/rarfiles.lst
        Type: File
        Size: 1210
 Packed size: 656
       Ratio: 54%
       mtime: 2024-02-26 17:05:59,000000000
  Attributes: -rw-r--r--
       CRC32: A5866556
     Host OS: Unix
 Compression: RAR 5.0(v50) -m3 -md=1m
       Flags: encrypted

*/
func parseEntryLines(lines []string) (Entry, error) {
	var e Entry
	var err error
	for _, s := range lines {
		parts := strings.SplitN(s, ": ", 2)
		if len(parts) != 2 {
			return e, fmt.Errorf("unexpected line, invalid key value pair, parts_len=%v raw_line=%s", len(parts), s)
		}
		name := strings.ToLower(parts[0])
		v := strings.TrimSpace(parts[1])

		switch name {
		case "name":
			e.Name = v
			if e.Name == "" {
				err = fmt.Errorf("Name field can not be empty")
			}
		case "type":
			e.Type = EntryType(v)
		case "size":
			e.Size, err = strconv.ParseInt(v, 10, 64)
		case "packed size":
			e.PackedSize = -1
			if v != "" {
				e.PackedSize, err = strconv.ParseInt(v, 10, 64)
			}
		case "ratio":
			e.Ratio = v
		case "mtime":
			// mtime: 2024-02-26 17:05:59,000000000
			e.Modified, _ = time.Parse(timeLayout, v)
		case "attributes":
			e.Attributes = v
		case "crc32":
			e.CRC = v
		case "host os":
			e.HostOS = v
		case "compression":
			e.Compression = v

		case "flags":
			e.Flags = v

		}
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func parseUnRARListOutput(d []byte) ([]Entry, error) {
	var res []Entry
	r := bytes.NewBuffer(d)
	scanner := bufio.NewScanner(r)
	err := advanceToFirstEntry(scanner)
	if err != nil {
		return nil, err
	}
	for {
		fields, err := getEntryLines(scanner)
		if err != nil {
			return nil, err
		}
		if len(fields) == 0 {
			// last entry
			break
		}
		e, err := parseEntryLines(fields)
		if err != nil {
			return nil, err
		}

		// skip Name field empty item, which maybe invalid
		if e.Name == "" {
			slog.Error("skip Name field empty item, which maybe invalid")
			continue
		}

		res = append(res, e)
	}
	return res, nil
}

func SetUnRARPath(path string) {
	unrarPath = path
}

func NewArchive(path string) (*Archive, error) {
	return newArchive(path, nil)
}

func NewEncryptedArchive(path string, password string) (*Archive, error) {
	return newArchive(path, &password)
}

// NewArchive uses 7z to extract a list of files in .7z archive
func newArchive(path string, password *string) (*Archive, error) {
	err := detectUnRARCached()
	if err != nil {
		return nil, err
	}
	numCPU := runtime.NumCPU()

	params := []string{"lt", fmt.Sprintf("-mt%d", numCPU)}
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

	cmd := exec.Command(unrarPath, params...)
	fixupEncoding(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = strings.TrimSpace(stdout.String())
		}
		re := regexp.MustCompile(`ERROR = Missing volume : .+`)
		subs := re.FindStringSubmatch(errMsg)
		if len(subs) > 0 {
			errMsg = strings.Join(subs, ",")
		}
		return nil, fmt.Errorf("err=%s cmd_run_err=%w", errMsg, err)
	}
	entries, err := parseUnRARListOutput(stdout.Bytes())
	if err != nil {
		return nil, err
	}
	return &Archive{
		Path:     path,
		Entries:  entries,
		password: &tmpPassword,
		maxCpuNum: numCPU,
	}, nil
}

func fixupEncoding(cmd *exec.Cmd) {
	// 在 alpine linux下, 如果用 LANG=C , 7z l 列文件会变成�乱码, 这不是latin1编码,
	// 因此，在alpine linux下不能用 LANG=C, 可以不指定，或指定为LANG=en_US.UTF-8 都OK
	// 而经测试, 在ArchLinux下和QNAP系统下, LANG=C 或 LANG=en_US.UTF-8 都表现一致
	envUpdateKV := "LANG=en_US.UTF-8"
	osEnv := os.Environ()
	if _, ok := os.LookupEnv("LANG"); ok {
		for k, e := range osEnv {
			if strings.HasPrefix(e, "LANG=") {
				osEnv[k] = envUpdateKV
				break
			}
		}
	} else {
		osEnv = append(osEnv, envUpdateKV)
	}
	cmd.Env = osEnv
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
		if e.Name == name {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("file not in the archive")
	}

	params := []string{"p", fmt.Sprintf("-mt%d", a.maxCpuNum)}
	if a.password != nil {
		params = append(params, fmt.Sprintf("-p%s", *a.password))
	}
	params = append(params, a.Path, name)

	cmd := exec.Command(unrarPath, params...)
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
