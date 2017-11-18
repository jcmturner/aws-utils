# aws-utils

Small single purpose executables in the sytle of Unix CLI commands.

* ec2inst - Query information about the EC2 instance the command is running on.

## Building
```
go build -ldflags "-X main.buildstamp=`date -u '+%FT%T%Z'` -X main.githash=`git rev-parse HEAD`"
```