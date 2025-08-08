本项目仅为后端项目，前端没写，需使用Postman

账号密码方面有校验，示例:
注册：
{"email":"128@qq.com","password":"a2345678?","confirmPassword":"a2345678?"}
<img width="1301" height="256" alt="image" src="https://github.com/user-attachments/assets/d2d657b8-7bf3-4a37-bb8f-2499612fd1c1" />
登录：
{"email":"128@qq.com","password":"a2345678?"}
<img width="1276" height="296" alt="image" src="https://github.com/user-attachments/assets/6643a38b-68da-4bdb-8b24-8b223b3fbf58" />
在注册登录界面无需token，但使用其他Path时需携带token，密钥：yflnlfSKD6nYhF4n
<img width="1255" height="419" alt="image" src="https://github.com/user-attachments/assets/1d0117c5-1f5d-44f5-beb7-5ae19fa6db55" />


使用docker将服务器部署在本地，启动前只需要执行命令
docker compose up -d

正式运行项目前还需要配置路径，选择当前工作目录，在程序实参上填写--config=config/dev.yaml
<img width="1561" height="1265" alt="image" src="https://github.com/user-attachments/assets/1c3bc46c-8ecf-4e8b-a008-08a1385c2e4b" />

