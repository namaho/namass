这是个Shadowsocks多用户管理系统，分为ssweb和ssagent两个服务。

ssagent与Shadowsocks运行在同一个机器上，通过Unix Socket与Shadowsocks交互通讯，向Shadowsocks发送添加端口、删除端口命令，并接收Shadowsocks发来的流量统计信息。

ssweb与ssagent通讯，向ssagent发送命令来间接控制Shadowsocks，也负责连接数据库以及控制Web界面。

<img src='arch.png'>

# Getting Started

首先要搭建Go编译环境来编译代码（ https://golang.org/doc/install ），不过如果你是运行在linux-amd64上的话可以直接下载编译好的来用：https://github.com/namaho/namass/releases/download/1.0/namass.zip

## Building 
```bash
git clone https://github.com/namaho/namass.git && \
cd namass && \
./ssweb/control build && \
./ssagent/control build 
```

## Installing & Configuring (Ubuntu)
MySQL:
```bash
sudo apt-get update && sudo apt-get install mysql-server
mysql -uroot -p -e "create database namass"
mysql -uroot -p namass < ./ssweb/db.sql
```

Shadowsocks:
```bash
sudo apt-get update && \
sudo apt-get install -y python-pip --force-yes && \
pip install shadowsocks

cat > /etc/shadowsocks.json << EOF
{
    "method": "aes-256-cfb",
    "port_password": {
        "8081": "whatevernomatter"
    },
    "server": "::",
    "timeout": 300
}
EOF
```

ssweb:
```bash
cat > ssweb/cfg.json << EOF
{
        "listen": ":9991",
        "loglevel": "debug",
        "logfile": "stderr",
        "ss_password_prefix": "namaho",
        "ssweb_http_address":"www.namaho.com",
        "ssagent_http_port": "9992",
        "ss_server_domain": "bilibilitv.pw",
        "server_token": "d434bb120995b6a3b349",
        "db": {
                "driver": "mysql",
                "host": "localhost",
                "port": "3306",
                "user": "aaa",
                "password": "bbb",
                "database": "namass"

        },
        "smtp": {
                "user": "xxxx@gmail.com",
                "password": "xxxx",
                "address": "smtp.gmail.com",
                "port": "587"
        }
}
EOF
```

ssagent:
```bash
cat > ssagent/cfg.json << EOF
{
        "listen": ":9992",
        "loglevel": "debug",
        "logfile": "stderr",
        "ssagent_unix_sock": "/tmp/ssagent.sock",
        "ssdaemon_unix_sock": "/var/run/shadowsocks-manager.sock",
        "ssweb_http_address": "www.namaho.com",
        "server_token": "d434bb120995b6a3b349",
        "report_interface": "eth0",
        "area": "1"
}
EOF
```

Nginx:
```bash
sudo apt-get update && sudo apt-get install nginx
cat > /etc/nginx/nginx.conf << EOF
user www-data;
worker_processes auto;
pid /run/nginx.pid;
http {
    include /etc/nginx/mime.types;
}
server {
    listen 80;
    server_name www.namaho.com namaho.com;
    root /path/to/namass/html;
    location /ssweb/ {
        proxy_pass http://127.0.0.1:9991;
        proxy_set_header x-forwarded-for $remote_addr;
    }
}
EOF
```

## Running
```bash
# Shadowsocks
sudo rm -f /var/run/shadowsocks-manager.sock && \
sudo ssserver --manager-address /var/run/shadowsocks-manager.sock -c /etc/shadowsocks.json -d start

# namass
./ssweb/control start && \
./ssagent/control start

# check log
./ssweb/control tail
```
