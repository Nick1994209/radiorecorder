# Script for downloading radio stream every tuesday and thursday


For running script local
```shell script
    go mod download
    go run main.go
```

For first building and deploying to server
```shell script
    GOOS=linux GOARCH=amd64 go build -o radiorecorder
    
    export SERVER_PATH="server_username@server_ip"
    cat ~/.ssh/id_rsa.pub | ssh $SERVER_PATH 'cat >> ~/.ssh/authorized_keys'
    # insert server password

    ssh $SERVER_PATH mkdir -p /home/records
    scp radiorecorder $SERVER_PATH:/home/records/radiorecorder
    scp .env $SERVER_PATH:/home/records/.env  # in .env keep security variables (for example SENTRY_DSN)
```

```shell script
    ssh $SERVER_PATH kill $(ps aux | grep recorder | awk '{print $2}')
    ssh $SERVER_PATH cd /home/records && nohup ./radiorecorder &
```

For downloading all records from server
```shell script
    scp "$SERVER_PATH:/home/records/*.mp3" .
```

Setup supervisor
```shell script
    ssh $SERVER_PATH apt-get install supervisor -y
    cat infra/supervisor.conf | ssh $SERVER_PATH 'cat > /etc/supervisor/conf.d/radiorecorder.conf'
    ssh $SERVER_PATH supervisorctl reread \
        && ssh $SERVER_PATH supervisorctl update \
        && ssh $SERVER_PATH supervisorctl restart radiorecorder 
```

Redeploy with supervisor
```shell script
    GOOS=linux GOARCH=amd64 go build -o radiorecorder \
      && ssh $SERVER_PATH rm /home/records/radiorecorder \
      && scp radiorecorder $SERVER_PATH:/home/records/radiorecorder \
      && ssh $SERVER_PATH supervisorctl restart radiorecorder
```
