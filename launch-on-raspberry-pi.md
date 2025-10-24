# Raspberry Pi でのAPIアプリ起動方法
Raspberry Pi上でGoアプリをサービス化してバックグラウンドで実行する

## 手順
### 1. Raspberry Pi用にバイナリをビルド
```
GOOS=linux GOARCH=arm GOARM=6 go build -o cellar-app
```
### 2. バイナリファイルをホームディレクトリに配置
ビルドしたバイナリファイルをscpでRaspberry Piに転送
```
scp cellar-app pi@<HOST_IP>:~
```
(注意): バイナリをRaspberry Pi上で実行中の場合は停止してからscpでファイルを転送する  
### 3. ユニットファイル作成
`/etc/systemd/system/cellar-app.service`に以下を記載して保存  
DB_PASSWORDは適宜変更する
```
[Unit]
Description=Cellar App API Service
After=network.target

[Service]
ExecStart=/usr/local/bin/cellar-app
WorkingDirectory=/usr/local/bin
Restart=on-failure
RestartSec=5s
User=pi
Group=pi
StandardOutput=append:/var/log/cellar-app.log
StandardError=append:/var/log/cellar-app.log
Environment=GIN_MODE=release
Environment=DB_PASSWORD=<REPLACE_HERE>
Environment=SSL_CERT_FILE=<REPLACE_HERE>
Environment=SSL_KEY_FILE=<REPLACE_HERE>

[Install]
WantedBy=multi-user.target

```
### 3. シンボリックリンク作成
```
sudo ln -s /home/pi/cellar-app /usr/local/bin/cellar-app
```

### 4. サービス起動
```
sudo systemctl daemon-reload
sudo sysctemctl enable cellar-app.service
sudo systemctl start cellar-app.service
```