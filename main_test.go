package main

import (
	"io"
	"net/http"
	"os"
	"testing"
)

// unit tests to play with go file types and interfaces

func TestOsFileEmbeddedStruct(t *testing.T) {
	t.Log("in TestOsFileStruct")

	type myFile struct {
		*os.File
	}

	osFile, err := os.Open("/usr/share/dict/words")
	if err != nil {
		t.Errorf("file open error error = %v", err)
	}
	defer osFile.Close()

	var myFileInstance = myFile{
		File: osFile,
	}

	var httpFile http.File = myFileInstance
	_, myFileIsOsFile := httpFile.(*os.File)

	// myFileInstance is not *os.File
	if myFileIsOsFile {
		t.Errorf("myFileIsOsFile got %v expected false", myFileIsOsFile)
	}

	var _ io.WriterTo = osFile
	httpFile = osFile
	_, osFileIsIoWriterTo := httpFile.(io.WriterTo)

	// *os.File is io.WriterTo
	if !osFileIsIoWriterTo {
		t.Errorf("osFileIsIoWriterTo got %v expected true", osFileIsIoWriterTo)
	}

	httpFile = myFileInstance
	_, myFileIsWriterTo := httpFile.(io.WriterTo)

	// myFile is io.WriterTo
	if !myFileIsWriterTo {
		t.Errorf("myFileIsWriterTo got %v expected true", osFileIsIoWriterTo)
	}
}

func TestOsFileEmbeddedInterface(t *testing.T) {
	t.Log("in TestOsFileEmbeddedInterface")

	type myFile struct {
		http.File
	}

	osFile, err := os.Open("/usr/share/dict/words")
	if err != nil {
		t.Errorf("file open error error = %v", err)
	}
	defer osFile.Close()

	var myFileInstance = myFile{
		File: osFile,
	}

	var httpFile http.File = myFileInstance
	_, myFileIsOsFile := httpFile.(*os.File)

	// myFileInstance is not *os.File
	if myFileIsOsFile {
		t.Errorf("myFileIsOsFile got %v expected false", myFileIsOsFile)
	}

	var _ io.WriterTo = osFile
	httpFile = osFile
	_, osFileIsIoWriterTo := httpFile.(io.WriterTo)

	// *os.File is io.WriterTo
	if !osFileIsIoWriterTo {
		t.Errorf("osFileIsIoWriterTo got %v expected true", osFileIsIoWriterTo)
	}

	httpFile = myFileInstance
	_, myFileIsWriterTo := httpFile.(io.WriterTo)

	// myFile is not io.WriterTo
	if myFileIsWriterTo {
		t.Errorf("myFileIsWriterTo got %v expected false", osFileIsIoWriterTo)
	}
}

func TestOsFileEmbeddedMyHttpFileInterface(t *testing.T) {
	t.Log("in TestOsFileEmbeddedMyHttpFileInterface")

	type myHttpFileInterface interface {
		http.File
		io.WriterTo
	}

	type myFile struct {
		myHttpFileInterface
	}

	osFile, err := os.Open("/usr/share/dict/words")
	if err != nil {
		t.Errorf("file open error error = %v", err)
	}
	defer osFile.Close()

	var myFileInstance = myFile{
		myHttpFileInterface: osFile,
	}

	var httpFile http.File = myFileInstance
	_, myFileIsOsFile := httpFile.(*os.File)

	// myFileInstance is not *os.File
	if myFileIsOsFile {
		t.Errorf("myFileIsOsFile got %v expected false", myFileIsOsFile)
	}

	var _ io.WriterTo = osFile
	httpFile = osFile
	_, osFileIsIoWriterTo := httpFile.(io.WriterTo)

	// *os.File is io.WriterTo
	if !osFileIsIoWriterTo {
		t.Errorf("osFileIsIoWriterTo got %v expected true", osFileIsIoWriterTo)
	}

	httpFile = myFileInstance
	_, myFileIsWriterTo := httpFile.(io.WriterTo)

	// myFile is io.WriterTo
	if !myFileIsWriterTo {
		t.Errorf("myFileIsWriterTo got %v expected true", osFileIsIoWriterTo)
	}
}
