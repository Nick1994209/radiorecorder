# Script for downloading radio stream every tuesday and thursday


For running script local
```shell script
    go mod download
    go run main.go
```

For building and deploy to server
```shell script
    GOOS=linux GOARCH=amd64 go build -o recorder
    
    export SERVER_PATH="server_username@server_ip"
    cat ~/.ssh/id_rsa.pub | ssh $SERVER_PATH 'cat >> ~/.ssh/authorized_keys'
    # insert server password

    ssh $SERVER_PATH mkdir -p /home/records
    scp recorder $SERVER_PATH:/home/records/recorder
    ssh $SERVER_PATH kill $(ps aux | grep recorder | awk '{print $2}')
    ssh $SERVER_PATH cd /home/records && nohup ./recorder &
```

For downloading all records from server
```shell script
    scp "$SERVER_PATH:/home/records/*.mp3" .
```
