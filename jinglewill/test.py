import json
import requests
from datetime import datetime, timedelta

# 读取 JSON 文件
with open('/opt/Logdog/alert_results.json', "r", encoding="utf-8") as f:
    alert_data = json.load(f)

result = []
time_format = "%Y-%m-%d %H:%M:%S"
time_only_format = "%H:%M"  # 仅保留小时和分钟

# 获取当前小时
# 获取当前小时
current_hour = datetime.now().hour
# 获取上一个小时（处理跨天情况）
previous_hour = current_hour - 1 if current_hour > 0 else 23

# 用字典来存储每个时间点的累计值
time_counts = {}

# 遍历原始数据
for entry in alert_data:
    timestamp = entry["timestamp"]
    timestamp_obj = datetime.strptime(timestamp, time_format)

    # 获取当前数据的时间的小时和分钟
    time_only = timestamp_obj.strftime(time_only_format)

    # 如果该时间在上一个小时内，保存数据
    if timestamp_obj.hour == previous_hour:
        for count in entry["alert_results"].values():
            if time_only in time_counts:
                time_counts[time_only] += count
            else:
                time_counts[time_only] = count

# 将字典转换为列表
for time, count in time_counts.items():
    result.append({
        "time": time,
        "value": count
    })

# 按照时间排序
result_sorted = sorted(result, key=lambda x: datetime.strptime(x["time"], time_only_format))

# 输出排序后的结果
print(result_sorted)

# 输出排序后的结果
print(json.dumps(result_sorted))


# 飞书机器人 Webhook 地址
url = 'https://open.feishu.cn/open-apis/bot/v2/hook/baa43d2e-c330-4030-8449-466af5053847'
# 消息内容
data = {
  "msg_type": "interactive",
  "card": {
    "elements": [
      {
        "tag": "chart",
        "chart_spec": {
          "type": "area",
          "title": {
            "text": "上个周期日志错误统计"
          },
          "data": {
            "values": result_sorted
          },
          "xField": "time",
          "yField": "value"
        }
      }
    ],
    "header": {
      "template": "purple",
      "title": {
        "content": "巴基斯坦生产日志错误通知",
        "tag": "plain_text"
      }
    }
  }
}

# 发送请求
response = requests.post(url, json=data)

# 打印返回的响应
print(response.status_code)
print(response.text)


# 0 9-18 * * * /usr/bin/python3 /opt/Logdog/test.py