package input

import (
	"bufio"
	"errors"
	"io"
	"os"
	"time"
)

type Reader interface {
	Read() ([]byte, error)
	End()
}

type File struct {
	file   *os.File
	name   string
	reader *bufio.Reader
	count  int
	done   chan struct{}
}

var ErrorDone = errors.New("DONE")

func NewFile(name string) (*File, error) {
	file, err := os.Open(name)

	if err != nil {
		return nil, err
	}
	// 将文件指针移到文件末尾，表示从文件末尾位置开始读取
	file.Seek(0, os.SEEK_END)

	// 创建一个文件读取器（bufio.Reader）用于逐行读取文件内容
	reader := bufio.NewReader(file)

	// 创建一个新的 File 实例并返回，初始化包括文件对象、读取器和一个关闭通道
	return &File{
		file:   file,
		reader: reader,
		done:   make(chan struct{}),
	}, nil
}

// 生成器逐行读取内容
func (f *File) Read() ([]byte, error) {
	for {
		select {
		case <-f.done:
			// 如果通道 f.done 被关闭，表示停止读取，返回 ErrorDone 错误
			return nil, ErrorDone
		default:
			// 使用 f.reader 逐行读取内容，直到遇到换行符 '\n'
			content, err := f.reader.ReadBytes('\n')
			if err != nil {
				// 如果遇到文件末尾（EOF），休眠 500 毫秒等待新内容写入文件
				if err == io.EOF {
					time.Sleep(500 * time.Millisecond)
				}

				break
			}

			return content, nil
		}
	}

	return nil, ErrorDone
}

func (f *File) End() {
	f.file.Close()
	// 关闭通道 f.done，通知读取过程停止
	close(f.done)
}
