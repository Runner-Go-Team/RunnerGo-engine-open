### 扩展

`php-xml \ php-xmlrpc`

### 命令行

` * * * * * php /data/Application/ApiPost/artisan schedule:run >> /data/logs/ApiPost/logs/commands.log 2>&1`

### 队列

`php //data/Application/ApiPost/artisan queue:work --queue=add_project_after_join_team,bind_project_after_pay,file_upload_oss,send_phone_sms,email_invite_register,email_invite_success,email --tries=1`

## 私有化部署

1. 配置环境变量
2. 初始化;
   php artisan init:system fsdfhuui1123
3. 其他主进程;
   php artisan octane:start --host=127.0.0.1 --port=8000
4. 部署定时任务;
   php artisan schedule:work >> /dev/stdout 2>&1
5. 部署队列任务;
   php /data/Application/ApiPost/artisan queue:work database --queue=add_project_after_join_team,email_invite_register,email_invite_success,email --tries=1

### 私有化升级



## 配置说明
| key                | 是否必填 | 默认值                                     | 说明                                                                               |
|--------------------|------|-----------------------------------------|----------------------------------------------------------------------------------|
| mysql数据库           ||||
| DB_HOST            | 否    | 127.0.0.1                               | 数据库host                                                                          |
| DB_PORT            | 否    | 3306                                    | 数据库端口                                                                            |
| DB_DATABASE        | 否    | apipost                                 | 数据库库名                                                                            |
| DB_USERNAME        | 否    | root                                    | 数据库用户名                                                                           |
| DB_PASSWORD        | 否    | password                                | 数据库密码                                                                            |
| Redis              ||||
| REDIS_HOST         | 否    | 127.0.0.1                               | redis服务端host                                                                     |
| REDIS_PORT         | 否    | 6379                                    | redis服务端端口                                                                       |
| REDIS_PASSWORD     | 否    |                                         | redis服务端密码                                                                       |
| REDIS_DB           | 否    | 0                                       | redis数据库id                                                                       |
| Mongodb            |      |                                         |
| MONGODB_DSN        | 否    | mongodb://root:password@127.0.0.1:27017 | Mongodb连接url                                                                     |
| MONGODB_DATABASE   | 否    | apipost                                 | Mongodb数据库名称                                                                     |
| 上传文件存储             ||||
| FILESYSTEM_SYH_DIR | 否    | /home/api/app/apipost-static            | 上传文件存储地址                                                                         |

### 更多配置
| key                | 是否必填 | 默认值                                                | 说明                                                                                 |
|--------------------|------|----------------------------------------------------|------------------------------------------------------------------------------------|
| 自定义服务信息            ||||
| SOCKET_TOKEN       | 否    |                                                    | 连接协作服务的加密token，apis服务和socket服务此参数需要一致                                              |
| 邮箱服务               |      |                                                    | 用于找回密码                                                                             |
| MAIL_MAILER        | 否    | smtp                                               | 邮箱协议【"smtp", "sendmail", "mailgun", "ses", "postmark", "log", "array", "failover"】 |
| MAIL_HOST          | 否    |                                                    | 邮箱服务地址                                                                             |
| MAIL_PORT          | 否    | 465                                                | 邮箱服务端口                                                                             |
| MAIL_USERNAME      | 否    |                                                    | 邮箱账号                                                                               |
| MAIL_PASSWORD      | 否    |                                                    | 邮箱密码                                                                               |
| MAIL_ENCRYPTION    | 否    |                                                    | 加密方式【"ssl","tls"】                                                                  |
| MAIL_FROM_ADDRESS  | 否    |                                                    | 发件人地址                                                                              |
| MAIL_FROM_NAME     | 否    |                                                    | 发件人名称                                                                              |

