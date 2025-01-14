package input

import (
	"bytes"
	"fmt"
	"github.com/zhjx922/alert/publisher"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Input struct {
	inputs         *Inputs
	files          map[string]*File
	done           chan struct{}
	publisher      *publisher.Publisher
	alertCounter   int       // 增加告警计数器
	alertStartTime time.Time // 增加告警开始时间
}

func NewInput(inputs *Inputs) *Input {
	return &Input{
		inputs:         inputs,
		files:          make(map[string]*File),
		done:           make(chan struct{}),
		alertCounter:   0,          // 初始化告警计数器
		alertStartTime: time.Now(), // 初始化告警开始时间
	}
}

// 关键字监控计算逻辑
func (i Input) read(reader Reader, filename string) {
	for {
		// 读取一行日志
		line, err := reader.Read()
		// 如果读取完成，退出循环
		if err == ErrorDone {
			return
		}
		// 初始化告警标志
		alert := false
		// 检查每个关键字是否在日志行中
	AlertFor:
		for _, word := range i.inputs.IncludeLines {
			if bytes.Contains(line, []byte(word)) {
				// 如果包含关键字，设置告警标志为true
				alert = true

				if len(i.inputs.ExcludeLines) > 0 {
					for _, exWord := range i.inputs.ExcludeLines {
						if bytes.Contains(line, []byte(exWord)) {
							// 如果包含排除关键字，取消告警标志，跳出循环
							alert = false
							break AlertFor
						}
					}
				}
			}
		}
		// 如果触发告警条件
		if alert {
			// 获取当前时间
			currentTime := time.Now()

			// 如果在一分钟内达到了告警次数阈值
			if currentTime.Sub(i.alertStartTime) <= time.Minute {

				// 增加计数器
				i.alertCounter++
				log.Printf("Counter: %d, Filename: %s\n", i.alertCounter, filename)

			} else {
				// 如果时间已经过了一分钟，重置计数器和时间戳
				i.alertCounter = 0
				i.alertStartTime = currentTime
				log.Println(i.alertCounter, "超过周期，重置计数器和时间戳")

			}

			if i.alertCounter == i.inputs.AlertCount {

				var message strings.Builder
				message.WriteString(fmt.Sprintf("告警项目: [%s]\n", i.inputs.Name))
				message.WriteString(fmt.Sprintf("告警日志路径: %s\n", filename))
				message.WriteString(fmt.Sprintf("告警关键字触发次数: %d次\n", i.alertCounter))
				message.WriteString(fmt.Sprintf("告警内容: %s", line))

				// 触发告警
				i.publisher.Write([]byte(message.String()))
			}
		}
	}
}

func (i *Input) AddFile(name string) {
	if _, ok := i.files[name]; !ok {
		if file, err := NewFile(name); err == nil {
			i.files[name] = file

			go i.read(i.files[name], name)
			log.Printf("Add File:%s\n", name)

		}
	}
}

// // 调用文件对象的 End 方法，结束对文件的读取

func (i *Input) RemoveFile(name string) {
	if _, ok := i.files[name]; ok {
		log.Printf("Remove File:%s\n", name)
		i.files[name].End()
		delete(i.files, name)
	}
}

func (i Input) scan() {

	log.Println("Scanning.....")

	// 创建一个用于存储文件名的映射，初始化为当前监控文件列表的副本
	files := make(map[string]bool)

	for name, _ := range i.files {
		files[name] = true
	}

	for _, path := range i.inputs.Paths {
		// 使用 filepath.Glob 获取匹配路径模式的文件列表
		pList, err := filepath.Glob(path)
		if err != nil {
			continue
		}

		for _, filename := range pList {

			delete(files, filename)

			fileInfo, err := os.Stat(filename)

			if err != nil {
				continue
			}

			if fileInfo.IsDir() {
				continue
			}

			i.AddFile(filename)
		}
	}
	// 遍历剩余的未被删除的文件，从监控列表中移除这些文件
	for name, _ := range files {
		i.RemoveFile(name)
	}

}

// 监控的文件列表与文件系统中的文件保持同步

func (i Input) Run(publisher *publisher.Publisher) {
	if i.inputs.ScanFrequency < 1 {
		i.inputs.ScanFrequency = 10
	}
	i.publisher = publisher
	i.scan()
	// 无限循环，等待退出信号或定时执行文件扫描
	for {
		select {
		case <-i.done:
			return // 如果接收到退出信号，则退出循环
		case <-time.After(time.Second * time.Duration(i.inputs.ScanFrequency)):
			i.scan() // 定时执行文件扫描操作
		}
	}
}
