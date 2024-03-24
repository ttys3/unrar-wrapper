package unrarwrapper

import "testing"

//testing .zip
func TestNewArchiveZip(t *testing.T) {
	fpath := "./test-data/example.zip"
	if _, err := NewArchive(fpath); err == nil {
		t.Errorf("NewArchive: %s should fail but not", fpath)
		t.Fail()
	}
}

//testing .rar
func TestNewArchiveRar(t *testing.T) {
	fpath := "./test-data/example.rar"
	if a, err := NewArchive(fpath); err != nil {
		t.Errorf("NewArchive: %s, err: %s", fpath, err)
		t.Fail()
	} else {
		for _,e := range a.Entries {
			t.Logf("item: %+v",e)
		}
	}
}

func TestNewArchiveRarEncrypted(t *testing.T) {
	fpath := "./test-data/encrypted.rar"
	if a, err := NewArchive(fpath); err != nil {
		t.Errorf("NewArchive: %s, err: %s", fpath, err)
		t.Fail()
	} else {
		for _,e := range a.Entries {
			t.Logf("item: %+v",e)
		}
	}
}

//7z can not list files whitout a password when rar encrypted with param:
// -hp[password]  Encrypt both file data and headers
func TestNewArchiveRarEncryptedHeader(t *testing.T) {
	fpath := "./test-data/encrypted.inc.headers.rar"
	passwd := "helloworld"
	if a, err := NewEncryptedArchive(fpath, passwd); err != nil {
		t.Errorf("NewArchive: %s, err: %s", fpath, err)
		t.Fail()
	} else {
		for _,e := range a.Entries {
			t.Logf("item: %+v",e)
		}
	}
}