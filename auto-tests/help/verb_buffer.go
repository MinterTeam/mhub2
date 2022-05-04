package help

import (
	"bytes"
	"os"
)

type VBuffer struct {
	buf  bytes.Buffer
	file *os.File
}

func NewVBuffer(filepath string) *VBuffer {
	os.Remove(filepath)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		println(err.Error())
	}
	return &VBuffer{file: file}
}

func (V *VBuffer) Write(p []byte) (n int, err error) {
	V.file.Write(p)
	return V.buf.Write(p)
}

func (V VBuffer) String() string {
	return V.buf.String()
}
