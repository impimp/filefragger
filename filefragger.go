package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"syscall"
	"time"
)

const (
	BufferSize int = 4096
)

func main() {
	source := flag.String("source", "", "source file")
	destination := flag.String("dest", "", "source file")
	flag.Parse()

	if *source == "" || *destination == "" {
		flag.PrintDefaults()
	}

	if _, err := os.Stat(*source); err != nil {
		panic(err)
	}

	if _, err := os.Stat(*destination); err == nil {
		panic("Destination file already exists")
	}

	if err := cp(*source, *destination); err != nil {
		panic(err)
	}

}

func progress(completed, size int) {
	if completed%10 == 0 {
		fmt.Printf("\rOn %d/%d", completed, size)

	}
}

func makeRanges(size int64, blocksize int) []int64 {

	var ret []int64
	for i := int64(size / int64(blocksize)); i >= 0; i -= 1 {
		ret = append(ret, i)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(ret), func(i, j int) { ret[i], ret[j] = ret[j], ret[i] })

	return ret
}

func cp(source, destination string) error {
	sourceFd, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFd.Close()

	size, err := sourceFd.Seek(0, io.SeekEnd)
	if err != nil {
		return (err)
	}
	_, err = sourceFd.Seek(0, io.SeekStart)
	if err != nil {
		return (err)
	}

	destinationFd, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return (err)
	}
	defer destinationFd.Close()
	destinationFd.Truncate(size)
	c_ := 0

	buffer := make([]byte, BufferSize)
	for _, block := range makeRanges(size, BufferSize) {
		progress(c_, int(size)/BufferSize)
		bytes_ := block * int64(BufferSize)
		sourceFd.Seek(bytes_, io.SeekStart)
		destinationFd.Seek(bytes_, io.SeekStart)
		n, serr := sourceFd.Read(buffer)
		if serr != nil && !errors.Is(serr, io.EOF) {
			return serr
		}

		_, werr := destinationFd.Write(buffer[:n])
		if werr != nil {
			return werr
		}

		syscall.Fdatasync(int(destinationFd.Fd()))
		c_ = c_ + 1
	}

	return nil
}
