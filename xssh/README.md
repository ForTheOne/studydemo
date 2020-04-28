`编译命令`:

`go build -x -v -ldflags "-s -w" -buildmode=c-shared -o xssh.so main.go`



example：

输入为：

```json
[{"user":"root","password":"****","host":"172.16.10.137","key":"","cmds":["ls","ps -ef | grep kube"],"port":22}]

key 与password，user 可缺省一个用“”空字符串代替即可
```



```python
from ctypes import *
import json
import sys


if __name__ == "__main__":
    # mac
    if sys.platform=="darwin":
        ssh = cdll.LoadLibrary("./darwin.so")
    # linux
    elif sys.platform=="linux":
        ssh = cdll.LoadLibrary("./linux.so")
    else:
        raise Exception("unknow platform")

    # 函数返回值类型
    ssh.SSH.restype  = c_char_p

    # 输入类型需为字节
    # [{"user":"root","password":"abc123","host":"172.16.10.137","key":"","cmds":["ps x | grep kube | grep -v grep | awk '{print $1}'"],"port":22}]
    p=json.dumps([{"user":"root","password":"****","host":"172.16.10.137","key":"","cmds":["ps x | grep kube | grep -v grep | awk '{print $1}'"],"port":22}]).encode()

    res = ssh.SSH(p)

    # res 为[] 列表 json反序列话就行
    res1 = json.loads(res)
    for i in res1:
        print(i)

```