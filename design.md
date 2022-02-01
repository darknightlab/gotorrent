# 功能设计

后端

-   下载
    -   按指定顺序
-   播种
-   重新强制校验

webui

-   下载
-   播种
-   设置
    -   修改 tracker
    -   修改端口
    -   比率

输入源

-   带钩子
-   RSS 订阅
-   监视文件夹

输出

-   日志
-   带钩子

存储

-   sqlite

```
POST:
/api
    /add
        /magnet
        {
            "Auth": {
                "Secret": "..."
            },
            "Magnet": "..."
        }
    /delete
    {
        "Auth": {
            "Secret": "..."
        },
        "Hash": "...",
        "DeleteFile": "yes or no"
    }
    /status
    {
        "Auth": {
            "Secret": "..."
        }
    }
    /torrent
        /{hash:string}
            /delete
            {
                "Auth": {
                    "Secret": "..."
                },
                "DeleteFile": "yes or no"
            }
            /status
            {
                "Auth": {
                    "Secret": "..."
                }
            }
```
