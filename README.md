# Logdog

轻量化流式日志告警系统

## 功能

### 已经实现了自定义
* 1、多关键字告警
* 2、多日志文件告警
* 3、自定义Body
* 4、自定义Header
* 5、自定义 告警次数（每个工程颗粒度）
* 6、关键词在周期触发告警抑制
* 7、at 全体或指定人
* 8、日志文件配置支持glob

## 运行

```yaml
go run main.go -c config_demo.yaml
```

## 配置说明

使用yaml格式

```yaml

inputs:
  # 项目名称，没啥用
  - name: project-MacOS
    # paths扫描频率(秒)
    scan_frequency: 60
    alert_count: 2
    # 监控的文件，支持glob
    paths:
      - ./*.log
    # 监控内容，包含内容即告警
    include_lines: ['error', 'warning']
    # 排除监控内容，包含不告警
    exclude_lines:
      - "success"
  - name: project-001
    paths:
      - /var/log/*.log
    include_lines: ['success']
output.http:
  method: POST
  # 这里是企微机器人的地址
  url: https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=*
  # Header头
  headers:
    - Content-Type application/json;charset=UTF-8
  # 留着扩展用
  format: json
  # 请求内容(%{content}会替换为日志告警行的内容)
  body: >
    {
      "msgtype": "markdown",
      "markdown": {
        "content": "DIY报警内容\n<font color=\"warning\">%{content}</font>"
      }
    }
```
```bash
// at 指定用户
	<at id="ou_xxx"></at> //使用open_id at指定人
	<at id="b6xxxxg8"></at> //使用user_id at指定人
	<at email="test@email.com"></at> //使用邮箱地址 at指定人
// at 所有人
	<at id=all></at>
```

## app/app.go

这是应用程序的入口，它包含了初始化配置、创建 `Alert` 实例以及运行应用的逻辑。

- `InitConfig`函数：读取配置文件并解析为一个 `Config` 对象。
- `NewAlert`函数：创建一个新的 `Alert` 实例，初始化配置。
- `Run`方法：启动应用，创建输入（`input.NewKeeper`）和输出（`publisher.NewPublisher`），然后运行输入，并传递给输出。

#### input/config.go

> 这个文件定义了配置结构体，包括监控的输入信息和HTTP输出信息。

#### input/file.go

> 这个文件包含了文件读取相关的代码，用于打开文件、读取文件内容、以及关闭文件。

#### input/input.go

> 这个文件包含了输入的核心逻辑。它负责监控指定的文件，检查是否包含特定的关键字，如果包含则触发告警，并将告警信息传递给输出。

#### input/keeper.go

> `Keeper` 是一个用于管理多个输入的结构。它初始化并运行多个输入，并可以协调它们的操作。

#### output/http.go

> 这个文件包含HTTP输出的配置结构体，定义了输出信息的URL、请求方法、请求头、格式以及请求体。

#### publisher/publisher.go

> Publisher` 负责将告警消息发送到HTTP端点。它接收来自输入的告警消息，根据配置格式化消息内容，然后通过HTTP请求发送到指定URL。

#### main.go

> 这是应用程序的入口点。它解析命令行参数，读取配置文件路径，创建 `Alert` 实例，然后运行应用。

整个流程如下：

1. `main.go` 解析命令行参数，读取配置文件路径。
2. 创建 `Alert` 实例，初始化配置。
3. 在 `Alert` 实例的 `Run` 方法中，创建输入（`input.NewKeeper`）和输出（`publisher.NewPublisher`）。
4. 启动输入，每个输入负责监控文件，检查关键字，如果匹配则触发告警。
5. 如果触发告警，告警信息被传递给输出，输出将其格式化并发送到指定URL。
6. 应用程序继续运行，循环检测文件变化和告警触发。

这个应用程序的主要功能是监控指定文件，检查文件内容中是否包含特定关键字，如果包含则触发告警，并将告警信息发送到指定的HTTP端点。它的核心逻辑是输入和输出之间的协作，输入负责监控文件和触发告警，输出负责将告警信息发送出去。


```bash
cat /var/log/message
2023/11/06 15:41:07 Counter: 1, Filename: 123.log
2023/11/06 15:41:08 Counter: 2, Filename: 123.log
2023/11/06 15:41:16 Counter: 1, Filename: 1.log
2023/11/06 15:41:18 Counter: 2, Filename: 1.log
2023/11/06 15:41:19 Counter: 3, Filename: 1.log
2023/11/06 15:41:19 Counter: 4, Filename: 1.log
2023/11/06 15:41:24 Counter: 1, Filename: 1222.log
2023/11/06 15:41:25 Counter: 2, Filename: 1222.log
每分钟执行一次这个脚本

#!/bin/bash

# 将count.log按文件名进行分组，并找到每个文件名的最大数值

one_minute_ago = $(date -d '1 minute ago' +'%Y/%m/%d %H:%M')
cat /var/log/message | grep $one_minute_ago >> /var/log/message
# $ 4 是Counter 的值 $6  Filename 的值 
awk '
{
    file = $6 
    number = $4
    if (number > max[file]) {
        max[file] = number
    }
}
END {
    for (file in max) {
        print max[file], file
    }
}
' count.log # 这个日志 根据 你写的程序 输出到到地方 如果是/var/log/message 那就更换成/var/log/message

# 然后 配置成计划任务一分中执行一次即完成, 通知一分钟之前 每个日志文件的告警总数 
```