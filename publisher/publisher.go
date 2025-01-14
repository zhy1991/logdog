package publisher

import (
	"encoding/json"
	"fmt"
	"github.com/zhjx922/alert/output"
	"net/http"
	"regexp"
	"strings"
)

type Publisher struct {
	http    *output.Http   // 保存 HTTP 配置的结构体
	regexp  *regexp.Regexp // 替换正则对象
	message chan []byte    // 用于接收消息的通道
}

// NewPublisher 创建一个新的发布者实例，接收一个 output.Http 结构体作为参数
func NewPublisher(http *output.Http) *Publisher {
	return &Publisher{
		http:    http,
		regexp:  regexp.MustCompile(`%{content}`), // 创建正则表达式对象，匹配 "%{content}"
		message: make(chan []byte),                // 创建消息通道
	}
}

// Write 将消息写入发布者的消息通道

func (p *Publisher) Write(message []byte) {
	p.message <- message
}

// Monitor 监控消息通道，当有消息时执行发布操作
func (p *Publisher) Monitor() {
	for {
		select {
		case m := <-p.message:
			p.curl(m)
		}
	}
}

// format 根据数据类型格式化数据
func (p *Publisher) format(data []byte, body []byte) []byte {
	var j interface{}
	json.Unmarshal([]byte(p.http.Body), &j)

	formatted := p.formatNested(j, body)
	b, _ := json.Marshal(formatted)
	return b
}

// formatNested 递归处理嵌套的 JSON 结构
func (p *Publisher) formatNested(data interface{}, body []byte) interface{} {
	switch data.(type) {
	case string:
		// 如果是字符串类型，进行正则表达式替换
		return p.regexp.ReplaceAllString(data.(string), string(body))
	case map[string]interface{}:
		// 如果是对象类型，递归处理
		formatted := make(map[string]interface{})
		for key, value := range data.(map[string]interface{}) {
			formatted[key] = p.formatNested(value, body)
		}
		return formatted
	case []interface{}:
		// 如果是数组类型，递归处理
		formatted := make([]interface{}, len(data.([]interface{})))
		for i, item := range data.([]interface{}) {
			formatted[i] = p.formatNested(item, body)
		}
		return formatted
	default:
		// 其他类型直接返回
		return data
	}
}

func (p *Publisher) curl(body []byte) {

	if p.http.Format == "json" {
		body = p.format([]byte(p.http.Body), body)
	} else {
		body = p.regexp.ReplaceAll([]byte(p.http.Body), body)
	}

	content := string(body)

	reader := strings.NewReader(content)
	request, err := http.NewRequest(p.http.Method, p.http.Url, reader)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, header := range p.http.Headers {
		hs := strings.SplitN(strings.Trim(header, " "), " ", 2)
		request.Header.Set(hs[0], hs[1])
	}

	client := http.Client{}

	client.Do(request)
}
