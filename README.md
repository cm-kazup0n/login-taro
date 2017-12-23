login-taro
------

`login-taro` generates commandline snipets for connecting to AWS EC2 instance with ssh.

## install

```
go get -u github.com/cm-kazup0n/login-taro
```

## usage

```
login-taro --region NAME_OF_REGION
```


### output

```
<Value of Name tag> | <ssh command>
```

### Run with Peco

With [peco](https://github.com/peco/peco), output can be filtered interactively, and filtered output can be eval like below.

#### bash/zsh

```
`login-taro --region ap-northeast-1 | peco`
```

#### fish

```
eval(login-taro --region ap-northeast-1 | peco)
```


## example

```
$login-taro --region ap-northeast-1
windows | ssh -o ProxyCommand='ssh -i ~/.ssh/id-rsa.pem -W %h:%p ec2-user@ec2-XX-XX-XX-XX.ap-northeast-1.compute.amazonaws.com' -i ~/.ssh/id-rsa.pem ec2-user@ip-YY-YY-YY-YY.ap-northeast-1.compute.internal
bastion | ssh -i ~/.ssh/id-rsa.pem ec2-user@ec2-XX-XX-XX-XX.ap-northeast-1.compute.amazonaws.com
```